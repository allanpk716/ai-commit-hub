package service

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/git"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// StartupProgress 启动进度
type StartupProgress struct {
	Stage   string `json:"stage"`
	Percent int    `json:"percent"`
	Message string `json:"message"`
}

// StartupService 启动服务
type StartupService struct {
	ctx             context.Context
	gitProjectRepo  repository.GitProjectRepository
	pushoverService *pushover.Service
	db              *gorm.DB
}

// NewStartupService 创建启动服务
func NewStartupService(
	ctx context.Context,
	gitProjectRepo repository.GitProjectRepository,
	pushoverService *pushover.Service,
) *StartupService {
	return &StartupService{
		ctx:             ctx,
		gitProjectRepo:  gitProjectRepo,
		pushoverService: pushoverService,
		db:              repository.GetDB(),
	}
}

// Preload 预加载所有项目状态
func (s *StartupService) Preload() error {
	logger.Info("开始启动预加载...")

	// 添加总体超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 在新 goroutine 中执行预加载
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.doPreload()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.Warn("启动预加载超时，将进入主界面")
		s.emitProgress(StartupProgress{
			Stage:   "complete",
			Percent: 100,
			Message: "完成",
		})
		return nil
	}
}

// doPreload 实际执行预加载逻辑
func (s *StartupService) doPreload() error {
	// 阶段 1: 初始化
	s.emitProgress(StartupProgress{
		Stage:   "initializing",
		Percent: 10,
		Message: "正在初始化...",
	})
	time.Sleep(500 * time.Millisecond)

	// 阶段 2: 检查扩展
	s.emitProgress(StartupProgress{
		Stage:   "extension",
		Percent: 20,
		Message: "检查扩展...",
	})
	time.Sleep(300 * time.Millisecond)

	// 阶段 3: 扫描项目
	projects, err := s.gitProjectRepo.GetAll()
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %w", err)
	}

	totalProjects := len(projects)
	if totalProjects == 0 {
		s.emitProgress(StartupProgress{
			Stage:   "complete",
			Percent: 100,
			Message: "完成",
		})
		return nil
	}

	// 并发检查所有项目
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数为 5
	completed := 0
	var mu sync.Mutex

	// 使用指针切片来收集需要更新的项目
	projectsToUpdate := make([]*models.GitProject, 0, totalProjects)

	for _, project := range projects {
		wg.Add(1)
		go func(proj *models.GitProject) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 检查项目状态
			s.checkProjectStatus(proj)

			// 收集需要更新的项目
			mu.Lock()
			projectsToUpdate = append(projectsToUpdate, proj)
			completed++
			percent := 20 + int(float64(completed)/float64(totalProjects)*70)
			s.emitProgress(StartupProgress{
				Stage:   "scanning",
				Percent: percent,
				Message: fmt.Sprintf("扫描项目 %d/%d...", completed, totalProjects),
			})
			mu.Unlock()
		}(&project)
	}

	wg.Wait()

	// 批量保存项目状态到数据库（避免并发写入）
	if len(projectsToUpdate) > 0 {
		logger.Infof("开始批量保存 %d 个项目状态到数据库...", len(projectsToUpdate))
		if err := s.db.Save(projectsToUpdate).Error; err != nil {
			logger.Errorf("批量保存项目状态失败: %v", err)
			return fmt.Errorf("批量保存项目状态失败: %w", err)
		}
		logger.Infof("批量保存项目状态成功")
	}

	// 阶段 4: 完成
	s.emitProgress(StartupProgress{
		Stage:   "complete",
		Percent: 100,
		Message: "完成",
	})

	logger.Info("启动预加载完成")
	return nil
}

// checkProjectStatus 检查单个项目状态
func (s *StartupService) checkProjectStatus(project *models.GitProject) {
	projectName := project.Name

	// 验证项目路径是否存在
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		logger.Debugf("[%s] 项目路径不存在，跳过状态检查: %s", projectName, project.Path)
		return
	}

	// 检查 Pushover 更新状态
	if s.pushoverService != nil {
		status, err := s.pushoverService.GetHookStatus(project.Path)
		if err != nil {
			logger.Debugf("[%s] 获取 Pushover 状态失败: %v", projectName, err)
		} else if status.Installed {
			latestVersion, err := s.pushoverService.GetExtensionVersion()
			if err != nil {
				logger.Debugf("[%s] 获取扩展版本失败: %v", projectName, err)
			} else {
				project.PushoverNeedsUpdate = pushover.CompareVersions(status.Version, latestVersion) < 0
			}
		}
	}

	// 检查 Git 状态（使用 context 实现可取消的超时）
	type stagingResult struct {
		status *git.StagingStatus
		err    error
	}

	resultChan := make(chan stagingResult, 1)

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		status, err := git.GetStagingStatus(project.Path)
		select {
		case resultChan <- stagingResult{status, err}:
		case <-ctx.Done():
			// Context 已取消，不发送结果
			return
		}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.Debugf("[%s] 获取 Git 状态失败: %v", projectName, result.err)
		} else {
			project.HasUncommittedChanges = len(result.status.Staged) > 0 || len(result.status.Unstaged) > 0
			project.UntrackedCount = len(result.status.Untracked)
		}
	case <-ctx.Done():
		// 超时，跳过 Git 状态检查
		logger.Debugf("[%s] Git 状态检查超时", projectName)
	}

	// 注意：不在这里更新数据库，而是在 doPreload 中批量保存
}

// emitProgress 发送进度事件
func (s *StartupService) emitProgress(progress StartupProgress) {
	runtime.EventsEmit(s.ctx, "startup-progress", progress)
}
