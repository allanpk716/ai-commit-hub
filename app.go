package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/git"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/allanpk716/ai-commit-hub/pkg/service"
	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
	"gorm.io/gorm"

	// Provider 注册 - 匿名导入以触发 init()
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/anthropic"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/deepseek"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/google"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/ollama"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/openai"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/openrouter"
	_ "github.com/allanpk716/ai-commit-hub/pkg/provider/phind"
)

// Command creates a new exec.Cmd with hidden window on Windows
// This prevents console windows from popping up when running external commands
func Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// On Windows, hide the console window to prevent popups
	if stdruntime.GOOS == "windows" {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}

	return cmd
}

// App struct
type App struct {
	ctx                  context.Context
	dbPath               string
	gitProjectRepo       *repository.GitProjectRepository
	commitHistoryRepo    *repository.CommitHistoryRepository
	configService        *service.ConfigService
	projectConfigService *service.ProjectConfigService
	pushoverService      *pushover.Service
	errorService         *service.ErrorService
	initError            error
	// 系统托盘相关字段
	systrayReady   chan struct{} // systray 就绪信号
	systrayExit    *sync.Once    // 确保只退出一次
	windowVisible  bool          // 窗口可见状态
	windowMutex    sync.RWMutex  // 保护 windowVisible
	systrayRunning atomic.Bool   // systray 运行状态
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		systrayReady:  make(chan struct{}),
		systrayExit:   &sync.Once{},
		windowVisible: true, // 启动时窗口可见
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Info("AI Commit Hub starting up...")

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get home directory: %v", err)
		return
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Errorf("Failed to create config directory: %v", err)
		return
	}

	a.dbPath = filepath.Join(configDir, "ai-commit-hub.db")

	// Initialize database
	dbConfig := &repository.DatabaseConfig{Path: a.dbPath}
	if err := repository.InitializeDatabase(dbConfig); err != nil {
		a.initError = fmt.Errorf("database initialization failed: %w", err)
		logger.Errorf("Failed to initialize database: %v", err)
		return
	}

	// Initialize repositories (only if database init succeeded)
	a.gitProjectRepo = repository.NewGitProjectRepository()
	a.commitHistoryRepo = repository.NewCommitHistoryRepository()

	// Initialize config service and ensure default config exists
	a.configService = service.NewConfigService()
	if _, err := a.configService.LoadConfig(ctx); err != nil {
		logger.Errorf("Failed to initialize config: %v", err)
		// Continue anyway - config will be created when needed
	}

	// Initialize project config service
	cfg, _ := a.configService.LoadConfig(ctx)
	a.projectConfigService = service.NewProjectConfigService(a.gitProjectRepo, cfg)

	// Run database migrations
	db := repository.GetDB()
	if err := repository.MigrateAddProjectAIConfig(db); err != nil {
		logger.Warnf("数据库迁移失败: %v", err)
		// Continue anyway - migration may have already been applied
	}

	// Run Pushover Hook migration
	if err := repository.MigrateAddPushoverHookFields(db); err != nil {
		logger.Warnf("Pushover Hook 迁移失败: %v", err)
		// Continue anyway - migration may have already been applied
	}

	// Initialize pushover service
	// 获取可执行文件所在目录作为 appPath
	execPath, err := os.Executable()
	if err != nil {
		logger.Errorf("获取可执行文件路径失败: %v", err)
	} else {
		appPath := filepath.Dir(execPath)
		a.pushoverService = pushover.NewService(appPath)

		// 自动下载 cc-pushover-hook 扩展（如果不存在）
		if a.pushoverService != nil {
			if !a.pushoverService.IsExtensionDownloaded() {
				logger.Info("cc-pushover-hook 扩展未安装，开始自动下载...")
				if err := a.pushoverService.CloneExtension(); err != nil {
					logger.Errorf("自动下载 cc-pushover-hook 扩展失败: %v", err)
					// 不中断启动流程，继续运行
				} else {
					logger.Info("cc-pushover-hook 扩展下载成功")
				}
			} else {
				logger.Info("cc-pushover-hook 扩展已存在")
			}
		}
	}

	// Initialize error service
	a.errorService = service.NewErrorService()

	// 同步所有项目的 Hook 状态（阻塞执行，确保前端获取到最新状态）
	if a.pushoverService != nil {
		logger.Info("准备启动 Hook 状态同步...")
		a.syncAllProjectsHookStatus()
	} else {
		logger.Warn("Pushover service 未初始化，跳过 Hook 状态同步")
	}

	logger.Info("AI Commit Hub initialized successfully")

	// 启动预加载（异步）
	if a.pushoverService != nil && a.gitProjectRepo != nil {
		go func() {
			startupService := service.NewStartupService(ctx, a.gitProjectRepo, a.pushoverService)
			if err := startupService.Preload(); err != nil {
				logger.Errorf("启动预加载失败: %v", err)
				// 失败时也发送完成事件（不带数据）
				runtime.EventsEmit(ctx, "startup-complete", nil)
				return
			}

			// 预加载成功，批量获取所有项目状态
			projects, err := a.gitProjectRepo.GetAll()
			if err != nil {
				logger.Errorf("获取项目列表失败: %v", err)
				runtime.EventsEmit(ctx, "startup-complete", nil)
				return
			}

			if len(projects) == 0 {
				// 无项目，发送完成事件（不带数据）
				runtime.EventsEmit(ctx, "startup-complete", nil)
				return
			}

			// 提取所有项目路径
			projectPaths := make([]string, len(projects))
			for i, p := range projects {
				projectPaths[i] = p.Path
			}

			// 批量获取所有项目状态
			statuses, err := a.GetAllProjectStatuses(projectPaths)
			if err != nil {
				logger.Errorf("批量获取项目状态失败: %v", err)
				// 失败时仍发送完成事件（不带数据），让用户进入主界面
				runtime.EventsEmit(ctx, "startup-complete", nil)
				return
			}

			logger.Infof("成功预加载 %d 个项目的状态", len(statuses))

			// 发送完成事件（包含预加载的状态数据）
			runtime.EventsEmit(ctx, "startup-complete", map[string]interface{}{
				"success":  true,
				"statuses": statuses,
			})
		}()
	} else {
		// 无需预加载，直接完成（不带数据）
		runtime.EventsEmit(ctx, "startup-complete", nil)
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	logger.Info("AI Commit Hub shutting down...")

	// 退出系统托盘
	systray.Quit()
}

// getTrayIcon 根据平台返回合适的图标
func (a *App) getTrayIcon() []byte {
	if stdruntime.GOOS == "windows" {
		return appIconICO
	}
	// macOS 和 Linux 可以使用 PNG
	return appIconPNG
}

// runSystray 启动系统托盘 (在单独的 goroutine 中运行)
func (a *App) runSystray() {
	// 延迟初始化,避免与 Wails 启动冲突
	time.Sleep(500 * time.Millisecond)

	logger.Info("正在初始化系统托盘...")

	// 标记 systray 开始运行
	a.systrayRunning.Store(true)

	systray.Run(
		a.onSystrayReady,
		func() {
			// systray 退出时的清理
			a.systrayRunning.Store(false)
			a.onSystrayExit()
		},
	)
}

// onSystrayReady 在 systray 就绪时调用
func (a *App) onSystrayReady() {
	logger.Info("系统托盘初始化成功")

	// 设置托盘图标
	systray.SetIcon(appIcon)
	systray.SetTitle("AI Commit Hub")
	systray.SetTooltip("AI Commit Hub - 点击显示窗口")

	// 创建菜单
	menu := systray.AddMenuItem("显示窗口", "显示主窗口")
	go func() {
		for range menu.ClickedCh {
			a.showWindow()
		}
	}()

	// 添加分隔线
	systray.AddSeparator()

	// 退出菜单项
	quitMenu := systray.AddMenuItem("退出应用", "完全退出应用")
	go func() {
		for range quitMenu.ClickedCh {
			a.quitApplication()
		}
	}()

	// 通知 systray 已就绪
	close(a.systrayReady)
}

// onSystrayExit 在 systray 退出时调用
func (a *App) onSystrayExit() {
	logger.Info("系统托盘已退出")
}

// showWindow 显示窗口
func (a *App) showWindow() {
	if a.ctx == nil {
		logger.Warn("showWindow: context 未初始化")
		return
	}

	a.windowMutex.Lock()
	defer a.windowMutex.Unlock()

	if a.windowVisible {
		logger.Debug("窗口已可见,跳过显示")
		return
	}

	logger.Info("显示窗口")
	runtime.WindowShow(a.ctx)
	a.windowVisible = true

	// 发送事件到前端
	runtime.EventsEmit(a.ctx, "window-shown", map[string]interface{}{
		"timestamp": time.Now(),
	})

	// === 健康检查和自动重启 ===
	// 检查 systray 是否还在运行
	if !a.systrayRunning.Load() {
		logger.Warn("检测到 systray 已停止,重新启动...")
		go a.runSystray()

		// 等待 systray 重新初始化完成
		time.Sleep(1 * time.Second)
		logger.Info("systray 重新启动完成")
	}
}

// hideWindow 隐藏窗口
func (a *App) hideWindow() {
	if a.ctx == nil {
		logger.Warn("hideWindow: context 未初始化")
		return
	}

	a.windowMutex.Lock()
	defer a.windowMutex.Unlock()

	if !a.windowVisible {
		logger.Debug("窗口已隐藏,跳过隐藏")
		return
	}

	logger.Info("隐藏窗口到托盘")
	runtime.WindowHide(a.ctx)
	a.windowVisible = false

	// 发送事件到前端
	runtime.EventsEmit(a.ctx, "window-hidden", map[string]interface{}{
		"timestamp": time.Now(),
	})
}

// quitApplication 完全退出应用
func (a *App) quitApplication() {
	// 使用 sync.Once 确保只执行一次
	a.systrayExit.Do(func() {
		logger.Info("应用正在退出...")

		if a.ctx != nil {
			runtime.Quit(a.ctx)
		} else {
			// 如果 context 未初始化,强制退出
			logger.Warn("context 未初始化,使用 os.Exit")
			os.Exit(0)
		}
	})
}

// onBeforeClose 拦截窗口关闭事件,隐藏到托盘而非退出
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	logger.Info("窗口关闭事件被触发,将隐藏到托盘")

	// 隐藏窗口而非退出
	a.hideWindow()

	// 返回 true 阻止窗口关闭
	return true
}

// Greet returns a greeting
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, AI Commit Hub is ready!", name)
}

// OpenConfigFolder opens the config folder in system file manager
func (a *App) OpenConfigFolder() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")

	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "windows":
		cmd = Command("explorer", configDir)
	case "darwin":
		cmd = Command("open", configDir)
	default:
		cmd = Command("xdg-open", configDir)
	}

	return cmd.Start()
}

// OpenExtensionFolder opens the cc-pushover-hook extension folder in system file manager
func (a *App) OpenExtensionFolder() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}

	// 获取扩展路径
	extensionPath := a.pushoverService.GetExtensionPath()

	// 检查目录是否存在
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		return fmt.Errorf("extension directory not found: %s", extensionPath)
	}

	// 根据操作系统选择命令
	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "windows":
		cmd = Command("explorer", extensionPath)
	case "darwin":
		cmd = Command("open", extensionPath)
	default:
		cmd = Command("xdg-open", extensionPath)
	}

	return cmd.Start()
}

// Terminal 终端类型
type Terminal struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// OpenInFileExplorer 在系统文件管理器中打开项目路径
func (a *App) OpenInFileExplorer(projectPath string) error {
	// 转换为绝对路径并清理格式
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}
	absPath = filepath.Clean(absPath)

	logger.Debugf("OpenInFileExplorer: 原始路径=%s, 绝对路径=%s", projectPath, absPath)

	// 检查路径是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("项目路径不存在: %s", absPath)
	}

	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "windows":
		// 使用 rundll32 调用 Shell API，这是 Windows 打开文件管理器的标准方式
		// 不会打开命令行窗口，正确处理各种路径格式
		cmd = Command("rundll32.exe", "url.dll,FileProtocolHandler", absPath)
	case "darwin":
		cmd = Command("open", absPath)
	case "linux":
		cmd = Command("xdg-open", absPath)
	default:
		return fmt.Errorf("unsupported platform: %s", stdruntime.GOOS)
	}

	logger.Debugf("OpenInFileExplorer: 执行命令=%s %v", cmd.Path, cmd.Args)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("打开文件管理器失败: %w", err)
	}

	return nil
}

// OpenInTerminal 在指定终端中打开项目路径
func (a *App) OpenInTerminal(projectPath, terminalType string) error {
	// 检查路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	var cmd *exec.Cmd

	switch stdruntime.GOOS {
	case "windows":
		// 转换为绝对路径并清理格式
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return fmt.Errorf("获取绝对路径失败: %w", err)
		}
		absPath = filepath.Clean(absPath)

		switch terminalType {
		case "powershell":
			// 使用 cmd /c start 启动新的独立 PowerShell 窗口
			// "PowerShell" 是窗口标题
			cmd = Command("cmd", "/c", "start", "PowerShell", "powershell",
				"-NoExit", "-Command", fmt.Sprintf("Set-Location -LiteralPath '%s'", absPath))
		case "cmd":
			// 使用 cmd /c start 启动新的独立 CMD 窗口
			// "CMD" 是窗口标题
			cmd = Command("cmd", "/c", "start", "CMD", "/k", fmt.Sprintf("cd /d %s", absPath))
		case "windows-terminal":
			// 使用 Windows Terminal 的 -d 参数直接设置工作目录
			cmd = Command("wt", "-d", absPath)
		default:
			return fmt.Errorf("不支持的终端类型: %s", terminalType)
		}
	case "darwin":
		switch terminalType {
		case "terminal":
			// 使用 AppleScript 打开 Terminal 并执行 cd 命令
			script := fmt.Sprintf(`tell application "Terminal" to do script "cd %s"`, projectPath)
			cmd = Command("osascript", "-e", script)
		case "iterm2":
			// 使用 AppleScript 打开 iTerm2 并执行 cd 命令
			script := fmt.Sprintf(`tell application "iTerm" to tell current window to create tab with default profile and tell current session to write text "cd %s"`, projectPath)
			cmd = Command("osascript", "-e", script)
		default:
			return fmt.Errorf("不支持的终端类型: %s", terminalType)
		}
	case "linux":
		// Linux 默认使用系统默认终端
		switch terminalType {
		case "default":
			// 尝试使用常见的 Linux 终端模拟器
			cmd = Command("x-terminal-emulator", "-e", fmt.Sprintf("cd %s && exec $SHELL", projectPath))
		default:
			return fmt.Errorf("不支持的终端类型: %s", terminalType)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", stdruntime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("打开终端失败: %w", err)
	}

	return nil
}

// GetAvailableTerminals 返回当前平台可用的终端列表
func (a *App) GetAvailableTerminals() []Terminal {
	switch stdruntime.GOOS {
	case "windows":
		return []Terminal{
			{ID: "powershell", Name: "PowerShell", Icon: "💠"},
			{ID: "cmd", Name: "命令提示符", Icon: "📟"},
			{ID: "windows-terminal", Name: "Windows Terminal", Icon: "🪟"},
		}
	case "darwin":
		return []Terminal{
			{ID: "terminal", Name: "Terminal", Icon: "📟"},
			{ID: "iterm2", Name: "iTerm2", Icon: "🔷"},
		}
	case "linux":
		return []Terminal{
			{ID: "default", Name: "默认终端", Icon: "💻"},
		}
	default:
		return []Terminal{}
	}
}

// GetAllProjects retrieves all projects
func (a *App) GetAllProjects() ([]models.GitProject, error) {
	if a.initError != nil {
		return nil, fmt.Errorf("app not initialized: %w", a.initError)
	}
	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}

// GetProjectsWithStatus 获取带状态的项目列表
func (a *App) GetProjectsWithStatus() ([]models.GitProject, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}

	// 填充运行时状态字段
	for i := range projects {
		project := &projects[i]

		// 检查 Pushover 更新状态
		if a.pushoverService != nil {
			status, err := a.pushoverService.GetHookStatus(project.Path)
			if err == nil && status.Installed {
				latestVersion, err := a.pushoverService.GetExtensionVersion()
				if err == nil {
					project.PushoverNeedsUpdate = pushover.CompareVersions(status.Version, latestVersion) < 0
				}
			}
		}

		// 检查 Git 状态
		stagingStatus, err := git.GetStagingStatus(project.Path)
		if err == nil {
			project.HasUncommittedChanges = len(stagingStatus.Staged) > 0 || len(stagingStatus.Unstaged) > 0
			project.UntrackedCount = len(stagingStatus.Untracked)
		}
	}

	return projects, nil
}

// GetSingleProjectStatus 获取单个项目的运行时状态
// 用于增量更新，避免检查所有项目
func (a *App) GetSingleProjectStatus(projectPath string) (*models.SingleProjectStatus, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	if projectPath == "" {
		return nil, fmt.Errorf("项目路径不能为空")
	}

	status := &models.SingleProjectStatus{
		Path: projectPath,
	}

	// 检查 Pushover 更新状态
	if a.pushoverService != nil {
		hookStatus, err := a.pushoverService.GetHookStatus(projectPath)
		if err == nil && hookStatus.Installed {
			latestVersion, err := a.pushoverService.GetExtensionVersion()
			if err == nil {
				status.PushoverNeedsUpdate = pushover.CompareVersions(hookStatus.Version, latestVersion) < 0
			}
		}
	}

	// 检查 Git 状态
	stagingStatus, err := git.GetStagingStatus(projectPath)
	if err == nil {
		status.HasUncommittedChanges = len(stagingStatus.Staged) > 0 || len(stagingStatus.Unstaged) > 0
		status.UntrackedCount = len(stagingStatus.Untracked)
	}

	logger.Infof("[GetSingleProjectStatus] 项目 %s 状态: hasUncommitted=%v, untracked=%d, pushoverUpdate=%v",
		projectPath, status.HasUncommittedChanges, status.UntrackedCount, status.PushoverNeedsUpdate)

	return status, nil
}

// AddProject adds a new project
func (a *App) AddProject(path string) (models.GitProject, error) {
	if a.initError != nil {
		return models.GitProject{}, fmt.Errorf("app not initialized: %w", a.initError)
	}

	// Validate path
	project := &models.GitProject{Path: path}
	if err := project.Validate(); err != nil {
		return models.GitProject{}, fmt.Errorf("项目验证失败: %w", err)
	}

	// Detect name
	name, err := project.DetectName()
	if err != nil {
		return models.GitProject{}, fmt.Errorf("无法检测项目名称: %w", err)
	}
	project.Name = name

	// Get next sort order
	maxOrder, err := a.gitProjectRepo.GetMaxSortOrder()
	if err != nil {
		return models.GitProject{}, fmt.Errorf("无法获取排序: %w", err)
	}
	project.SortOrder = maxOrder + 1

	// Save to database
	if err := a.gitProjectRepo.Create(project); err != nil {
		return models.GitProject{}, fmt.Errorf("保存项目失败: %w", err)
	}

	return *project, nil
}

// DeleteProject deletes a project
func (a *App) DeleteProject(id uint) error {
	if a.initError != nil {
		return fmt.Errorf("app not initialized: %w", a.initError)
	}
	if err := a.gitProjectRepo.Delete(id); err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}
	return nil
}

// SelectProjectFolder opens a folder selection dialog
func (a *App) SelectProjectFolder() (string, error) {
	selectedFile, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Git 仓库",
	})
	if err != nil {
		return "", fmt.Errorf("打开文件夹选择对话框失败: %w", err)
	}
	if selectedFile == "" {
		return "", nil // User canceled - return empty string with no error
	}
	return selectedFile, nil
}

// MoveProject moves a project up or down
func (a *App) MoveProject(id uint, direction string) error {
	if a.initError != nil {
		return fmt.Errorf("app not initialized: %w", a.initError)
	}

	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		return fmt.Errorf("获取项目列表失败: %w", err)
	}

	// Find current project index
	var currentIndex int = -1
	for i, p := range projects {
		if p.ID == id {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return fmt.Errorf("项目不存在")
	}

	// Calculate new index
	newIndex := currentIndex
	if direction == "up" && currentIndex > 0 {
		newIndex = currentIndex - 1
	} else if direction == "down" && currentIndex < len(projects)-1 {
		newIndex = currentIndex + 1
	} else {
		return nil // No change needed
	}

	// Swap sort orders
	projects[currentIndex].SortOrder, projects[newIndex].SortOrder =
		projects[newIndex].SortOrder, projects[currentIndex].SortOrder

	// Update both projects in a transaction
	db := repository.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&projects[currentIndex]).Error; err != nil {
			return fmt.Errorf("更新项目失败: %w", err)
		}
		if err := tx.Save(&projects[newIndex]).Error; err != nil {
			return fmt.Errorf("更新项目失败: %w", err)
		}
		return nil
	})
}

// ReorderProjects reorders projects based on new order
func (a *App) ReorderProjects(projects []models.GitProject) error {
	if a.initError != nil {
		return fmt.Errorf("app not initialized: %w", a.initError)
	}

	for i := range projects {
		projects[i].SortOrder = i
		if err := a.gitProjectRepo.Update(&projects[i]); err != nil {
			return fmt.Errorf("更新项目排序失败: %w", err)
		}
	}
	return nil
}

// GetProjectStatus retrieves the git status of a project
func (a *App) GetProjectStatus(projectPath string) (map[string]interface{}, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	status, err := git.GetProjectStatus(context.Background(), projectPath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"branch":       status.Branch,
		"staged_files": status.StagedFiles,
		"has_staged":   status.HasStaged,
	}, nil
}

// GenerateCommit generates a commit message using AI
func (a *App) GenerateCommit(projectPath, provider, language string) error {
	logger.Info("App.GenerateCommit 被调用")
	logger.Infof("参数 - projectPath: %s, provider: %s, language: %s", projectPath, provider, language)

	if a.initError != nil {
		errMsg := fmt.Sprintf("应用未正确初始化: %v", a.initError)
		logger.Errorf(errMsg)
		return a.initError
	}

	commitService := service.NewCommitService(a.ctx)
	logger.Info("CommitService 创建成功，开始生成...")
	err := commitService.GenerateCommit(projectPath, provider, language)
	if err != nil {
		logger.Errorf("CommitService.GenerateCommit 返回错误: %v", err)
	} else {
		logger.Info("CommitService.GenerateCommit 执行完成（已启动异步生成）")
	}
	return err
}

// CommitLocally commits changes to local git repository
func (a *App) CommitLocally(projectPath, message string) error {
	logger.Infof("CommitLocally 被调用 - projectPath: %s, message: %s", projectPath, message)

	if a.initError != nil {
		logger.Errorf("数据库初始化错误: %v", a.initError)
		return a.initError
	}

	if message == "" {
		err := fmt.Errorf("commit 消息不能为空")
		logger.Errorf("提交失败: %v", err)
		return err
	}

	// Save current directory and change to project path
	originalDir, err := os.Getwd()
	if err != nil {
		err := fmt.Errorf("failed to get current directory: %w", err)
		logger.Errorf("获取当前目录失败: %v", err)
		return err
	}

	if err := os.Chdir(projectPath); err != nil {
		err := fmt.Errorf("failed to change directory: %w", err)
		logger.Errorf("切换到项目目录失败: %v", err)
		return err
	}
	defer os.Chdir(originalDir)

	logger.Infof("准备提交 - 目录: %s", projectPath)

	// Use the existing CommitChanges function from git package
	if err := git.CommitChanges(context.Background(), message); err != nil {
		logger.Errorf("CommitChanges 失败: %v", err)
		return err
	}

	logger.Infof("提交成功 - 目录: %s", projectPath)

	// 发送项目状态变更事件，触发前端刷新
	runtime.EventsEmit(a.ctx, "project-status-changed", map[string]interface{}{
		"projectPath": projectPath,
		"changeType":  "commit",
		"timestamp":   time.Now(),
	})

	return nil
}

// SaveCommitHistory saves a generated commit message to history
func (a *App) SaveCommitHistory(projectID uint, message, provider, language string) error {
	if a.initError != nil {
		return a.initError
	}

	history := &models.CommitHistory{
		ProjectID: projectID,
		Message:   message,
		Provider:  provider,
		Language:  language,
	}

	if err := a.commitHistoryRepo.Create(history); err != nil {
		return fmt.Errorf("保存历史记录失败: %w", err)
	}
	return nil
}

// GetProjectHistory retrieves commit history for a project
func (a *App) GetProjectHistory(projectID uint) ([]models.CommitHistory, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	histories, err := a.commitHistoryRepo.GetByProjectID(projectID, 10)
	if err != nil {
		return nil, fmt.Errorf("获取历史记录失败: %w", err)
	}
	return histories, nil
}

// GetProjectAIConfig 获取项目的 AI 配置
func (a *App) GetProjectAIConfig(projectID int) (*service.ProjectAIConfig, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	config, err := a.projectConfigService.GetProjectAIConfig(uint(projectID))
	if err != nil {
		return nil, fmt.Errorf("获取项目 AI 配置失败: %w", err)
	}

	return config, nil
}

// UpdateProjectAIConfig 更新项目的 AI 配置
func (a *App) UpdateProjectAIConfig(projectID int, provider, language, model string, useDefault bool) error {
	if a.initError != nil {
		return a.initError
	}

	project, err := a.gitProjectRepo.GetByID(uint(projectID))
	if err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	project.UseDefault = useDefault

	if useDefault {
		project.Provider = nil
		project.Language = nil
		project.Model = nil
	} else {
		if provider != "" {
			project.Provider = &provider
		}
		if language != "" {
			project.Language = &language
		}
		if model != "" {
			project.Model = &model
		}
	}

	if err := a.gitProjectRepo.Update(project); err != nil {
		return fmt.Errorf("更新项目配置失败: %w", err)
	}

	return nil
}

// ValidateProjectConfig 验证项目配置
func (a *App) ValidateProjectConfig(projectID int) (valid bool, resetFields []string, suggestedConfig map[string]interface{}, err error) {
	if a.initError != nil {
		return false, nil, nil, a.initError
	}

	valid, fields, config, err := a.projectConfigService.ValidateProjectConfig(uint(projectID))
	if err != nil {
		return false, nil, nil, fmt.Errorf("验证项目配置失败: %w", err)
	}

	if config != nil {
		suggestedConfig = map[string]interface{}{
			"provider":  config.Provider,
			"language":  config.Language,
			"isDefault": config.IsDefault,
		}
	}

	return valid, fields, suggestedConfig, nil
}

// ConfirmResetProjectConfig 确认并重置项目配置
func (a *App) ConfirmResetProjectConfig(projectID int) error {
	if a.initError != nil {
		return a.initError
	}

	if err := a.projectConfigService.ResetProjectToDefaults(uint(projectID)); err != nil {
		return fmt.Errorf("重置项目配置失败: %w", err)
	}

	return nil
}

// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (a *App) GetConfiguredProviders() ([]models.ProviderInfo, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	cfg, err := a.configService.LoadConfig(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	providers := a.configService.GetConfiguredProviders(cfg)
	return providers, nil
}

// GetPushoverHookStatus 获取项目的 Pushover Hook 状态
func (a *App) GetPushoverHookStatus(projectPath string) (*pushover.HookStatus, error) {
	if a.initError != nil {
		return nil, a.initError
	}
	if a.pushoverService == nil {
		return nil, fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.GetHookStatus(projectPath)
}

// GetPushStatus 获取项目的推送状态
func (a *App) GetPushStatus(projectPath string) *git.PushStatus {
	if a.initError != nil {
		return &git.PushStatus{
			CanPush:      false,
			AheadCount:   0,
			BehindCount:  0,
			RemoteBranch: "",
			Error:        a.initError.Error(),
		}
	}
	status, err := git.GetPushStatus(projectPath)
	if err != nil {
		return &git.PushStatus{
			CanPush:      false,
			AheadCount:   0,
			BehindCount:  0,
			RemoteBranch: "",
			Error:        err.Error(),
		}
	}
	return status
}

// InstallPushoverHook 为项目安装 Pushover Hook
func (a *App) InstallPushoverHook(projectPath string, force bool) (*pushover.InstallResult, error) {
	if a.initError != nil {
		return &pushover.InstallResult{Success: false, Message: a.initError.Error()}, nil
	}
	if a.pushoverService == nil {
		return &pushover.InstallResult{Success: false, Message: "pushover service 未初始化"}, nil
	}

	// 调用 Service 层安装
	result, err := a.pushoverService.InstallHook(projectPath, force)
	if err != nil {
		return result, err
	}

	// 安装成功后同步数据库状态
	if result.Success {
		if syncErr := a.syncProjectHookStatusByPath(projectPath); syncErr != nil {
			logger.Warnf("同步 Hook 状态失败: %v", syncErr)
			// 不影响安装结果，只记录错误
		}
	}

	return result, nil
}

// UninstallPushoverHook 卸载项目的 Pushover Hook
func (a *App) UninstallPushoverHook(projectPath string) error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}

	// 调用 Service 层卸载
	if err := a.pushoverService.UninstallHook(projectPath); err != nil {
		return err
	}

	// 卸载成功后同步数据库状态
	if syncErr := a.syncProjectHookStatusByPath(projectPath); syncErr != nil {
		logger.Warnf("同步 Hook 状态失败: %v", syncErr)
		// 不影响卸载结果，只记录错误
	}

	return nil
}

// UpdatePushoverHook 更新项目的 Pushover Hook
func (a *App) UpdatePushoverHook(projectPath string) (*pushover.InstallResult, error) {
	if a.initError != nil {
		return &pushover.InstallResult{Success: false, Message: a.initError.Error()}, nil
	}
	if a.pushoverService == nil {
		return &pushover.InstallResult{Success: false, Message: "pushover service 未初始化"}, nil
	}

	// 调用 Service 层更新
	result, err := a.pushoverService.UpdateHook(projectPath)
	if err != nil {
		return &pushover.InstallResult{Success: false, Message: err.Error()}, nil
	}

	// 更新成功后同步数据库状态
	if result.Success {
		if syncErr := a.syncProjectHookStatusByPath(projectPath); syncErr != nil {
			logger.Warnf("同步 Hook 状态失败: %v", syncErr)
			// 不影响更新结果，只记录错误
		}
	}

	return result, nil
}

// ReinstallPushoverHook 重装项目的 Pushover Hook
func (a *App) ReinstallPushoverHook(projectPath string) (*pushover.InstallResult, error) {
	if a.initError != nil {
		return &pushover.InstallResult{Success: false, Message: a.initError.Error()}, nil
	}
	if a.pushoverService == nil {
		return &pushover.InstallResult{Success: false, Message: "pushover service 未初始化"}, nil
	}

	result, err := a.pushoverService.ReinstallHook(projectPath)
	if err != nil {
		return &pushover.InstallResult{Success: false, Message: err.Error()}, nil
	}

	// 重装成功后同步数据库状态
	if result.Success {
		if syncErr := a.syncProjectHookStatusByPath(projectPath); syncErr != nil {
			logger.Warnf("同步 Hook 状态失败: %v", syncErr)
			// 不影响重装结果，只记录错误
		}
	}

	return result, nil
}

// SetPushoverNotificationMode 设置项目的通知模式
func (a *App) SetPushoverNotificationMode(projectPath string, mode string) error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.SetNotificationMode(projectPath, pushover.NotificationMode(mode))
}

// ToggleNotification 切换指定项目的通知类型
// 通过创建或删除 .no-pushover 或 .no-windows 文件来实现
func (a *App) ToggleNotification(projectPath string, notificationType string) error {
	logger.Infof("切换通知状态: 项目=%s, 类型=%s", projectPath, notificationType)

	// 检查初始化错误
	if a.initError != nil {
		return fmt.Errorf("应用未正确初始化: %w", a.initError)
	}

	// 验证项目路径
	if projectPath == "" {
		return fmt.Errorf("项目路径不能为空")
	}

	// 验证通知类型
	var fileName string
	switch notificationType {
	case "pushover":
		fileName = ".no-pushover"
	case "windows":
		fileName = ".no-windows"
	default:
		return fmt.Errorf("不支持的通知类型: %s", notificationType)
	}

	// 文件直接放在项目根目录下，与 Python hook 的路径一致
	filePath := filepath.Join(projectPath, fileName)

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建文件来禁用通知
			file, err := os.Create(filePath)
			if err != nil {
				logger.Errorf("创建禁用文件失败: %v", err)
				return fmt.Errorf("创建禁用文件失败: %w", err)
			}
			file.Close()
			logger.Infof("已禁用 %s 通知: 创建 %s", notificationType, fileName)
			return nil
		}
		// 其他错误
		logger.Errorf("检查文件失败: %v", err)
		return fmt.Errorf("检查文件失败: %w", err)
	}

	// 文件存在，删除文件来启用通知
	if fileInfo.IsDir() {
		return fmt.Errorf("%s 是目录，不是文件", fileName)
	}

	if err := os.Remove(filePath); err != nil {
		logger.Errorf("删除禁用文件失败: %v", err)
		return fmt.Errorf("删除禁用文件失败: %w", err)
	}

	logger.Infof("已启用 %s 通知: 删除 %s", notificationType, fileName)
	return nil
}

// CheckPushoverConfig 检查 Pushover 环境变量是否已配置
// 返回配置状态，用于应用启动时的检查
func (a *App) CheckPushoverConfig() map[string]interface{} {
	token := os.Getenv("PUSHOVER_TOKEN")
	user := os.Getenv("PUSHOVER_USER")

	tokenSet := token != ""
	userSet := user != ""
	valid := tokenSet && userSet

	result := map[string]interface{}{
		"valid":     valid,
		"token_set": tokenSet,
		"user_set":  userSet,
	}

	if valid {
		logger.Info("Pushover 配置检查: 已配置")
	} else {
		logger.Warn("Pushover 配置检查: 未配置 (TOKEN=%t, USER=%t)", tokenSet, userSet)
	}

	return result
}

// GetPushoverExtensionInfo 获取 cc-pushover-hook 扩展信息
func (a *App) GetPushoverExtensionInfo() (*pushover.ExtensionInfo, error) {
	if a.initError != nil {
		return nil, a.initError
	}
	if a.pushoverService == nil {
		return nil, fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.GetExtensionInfo()
}

// CheckPushoverExtensionUpdates 检查 cc-pushover-hook 扩展更新
func (a *App) CheckPushoverExtensionUpdates() (map[string]interface{}, error) {
	if a.initError != nil {
		return nil, a.initError
	}
	if a.pushoverService == nil {
		return nil, fmt.Errorf("pushover service 未初始化")
	}

	needsUpdate, currentVersion, latestVersion, err := a.pushoverService.CheckForUpdates()
	if err != nil {
		return nil, fmt.Errorf("检查扩展更新失败: %w", err)
	}

	return map[string]interface{}{
		"needs_update":    needsUpdate,
		"current_version": currentVersion,
		"latest_version":  latestVersion,
	}, nil
}

// ClonePushoverExtension 克隆 cc-pushover-hook 扩展
func (a *App) ClonePushoverExtension() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.CloneExtension()
}

// UpdatePushoverExtension 更新 cc-pushover-hook 扩展
func (a *App) UpdatePushoverExtension() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.UpdateExtension()
}

// ReclonePushoverExtension 重新下载 cc-pushover-hook 扩展
func (a *App) ReclonePushoverExtension() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.RecloneExtension()
}

// CheckPushoverUpdates 检查项目的 Pushover Hook 更新
func (a *App) CheckPushoverUpdates(projectPath string) (map[string]interface{}, error) {
	if a.initError != nil {
		return nil, a.initError
	}
	if a.pushoverService == nil {
		return nil, fmt.Errorf("pushover service 未初始化")
	}

	// 获取扩展版本
	latestVersion, err := a.pushoverService.GetExtensionVersion()
	if err != nil {
		return nil, fmt.Errorf("获取扩展版本失败: %w", err)
	}

	// 获取项目中的 Hook 状态
	checker := pushover.NewStatusChecker(projectPath)
	status, err := checker.GetStatus(latestVersion)
	if err != nil {
		return nil, fmt.Errorf("获取 Hook 状态失败: %w", err)
	}

	if !status.Installed {
		return map[string]interface{}{
			"update_available": false,
			"current_version":  status.Version,
			"latest_version":   latestVersion,
			"installed":        false,
		}, nil
	}

	// 使用 status.UpdateAvailable（已在 GetStatus 中计算）
	return map[string]interface{}{
		"update_available": status.UpdateAvailable,
		"current_version":  status.Version,
		"latest_version":   latestVersion,
		"installed":        true,
	}, nil
}

// syncAllProjectsHookStatus 同步所有项目的 Pushover Hook 状态
func (a *App) syncAllProjectsHookStatus() {
	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		logger.Errorf("获取项目列表失败: %v", err)
		return
	}

	logger.Infof("开始同步 %d 个项目的 Hook 状态...", len(projects))

	for _, project := range projects {
		if err := a.syncProjectHookStatus(&project); err != nil {
			logger.Warnf("同步项目 %s Hook 状态失败: %v", project.Name, err)
		}
	}

	logger.Info("Hook 状态同步完成")
}

// syncProjectHookStatus 同步单个项目的 Hook 状态
func (a *App) syncProjectHookStatus(project *models.GitProject) error {
	logger.Debugf("正在检查项目 %s (路径: %s) 的 Hook 状态...", project.Name, project.Path)
	status, err := a.pushoverService.GetHookStatus(project.Path)
	if err != nil {
		return fmt.Errorf("获取 Hook 状态失败: %w", err)
	}

	logger.Debugf("项目 %s Hook 状态: installed=%v, mode=%s", project.Name, status.Installed, status.Mode)
	logger.Debugf("数据库中状态: installed=%v, mode=%s", project.HookInstalled, project.NotificationMode)

	// 只在状态发生变化时更新数据库
	needsUpdate := project.HookInstalled != status.Installed ||
		(status.Installed && project.NotificationMode != string(status.Mode))

	if !needsUpdate {
		logger.Debugf("项目 %s 状态无需更新", project.Name)
		return nil
	}

	project.HookInstalled = status.Installed
	project.NotificationMode = string(status.Mode)
	project.HookVersion = status.Version

	if status.Installed && status.InstalledAt != nil {
		project.HookInstalledAt = status.InstalledAt
	} else {
		project.HookInstalledAt = nil
	}

	if err := a.gitProjectRepo.Update(project); err != nil {
		return fmt.Errorf("更新数据库失败: %w", err)
	}

	logger.Infof("已更新项目 %s 的 Hook 状态: installed=%v, mode=%s",
		project.Name, status.Installed, status.Mode)

	return nil
}

// syncProjectHookStatusByPath 根据路径同步项目状态
func (a *App) syncProjectHookStatusByPath(projectPath string) error {
	// 根据 path 获取项目
	project, err := a.gitProjectRepo.GetByPath(projectPath)
	if err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	return a.syncProjectHookStatus(project)
}

// SyncProjectHookStatus 同步单个项目的 Hook 状态
func (a *App) SyncProjectHookStatus(projectPath string) error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}

	return a.syncProjectHookStatusByPath(projectPath)
}

// SyncAllProjectsHookStatus 手动同步所有项目的 Hook 状态
func (a *App) SyncAllProjectsHookStatus() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}

	a.syncAllProjectsHookStatus()
	return nil
}

// DebugHookStatus 调试方法：返回所有项目的 Hook 状态
func (a *App) DebugHookStatus() map[string]interface{} {
	result := make(map[string]interface{})

	if a.initError != nil {
		result["error"] = a.initError.Error()
		return result
	}

	if a.pushoverService == nil {
		result["error"] = "pushover service 未初始化"
		return result
	}

	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		result["error"] = fmt.Sprintf("获取项目失败: %v", err)
		return result
	}

	projectStatus := make([]map[string]interface{}, 0, len(projects))
	for _, project := range projects {
		status, err := a.pushoverService.GetHookStatus(project.Path)
		statusInfo := map[string]interface{}{
			"name":                 project.Name,
			"path":                 project.Path,
			"db_hook_installed":    project.HookInstalled,
			"db_notification_mode": project.NotificationMode,
			"db_hook_version":      project.HookVersion,
		}

		if err != nil {
			statusInfo["api_error"] = err.Error()
		} else {
			statusInfo["api_installed"] = status.Installed
			statusInfo["api_mode"] = status.Mode
			statusInfo["api_version"] = status.Version
			statusInfo["match"] = project.HookInstalled == status.Installed
		}

		projectStatus = append(projectStatus, statusInfo)
	}

	result["projects"] = projectStatus
	result["total"] = len(projects)
	return result
}

// PushToRemote 推送项目到远程仓库
func (a *App) PushToRemote(projectPath string) error {
	logger.Infof("PushToRemote 被调用 - projectPath: %s", projectPath)

	if a.initError != nil {
		logger.Errorf("数据库初始化错误: %v", a.initError)
		return a.initError
	}

	// 保存当前目录并切换到项目路径
	originalDir, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to get current directory: %w", err)
		logger.Errorf("获取当前目录失败: %v", err)
		return err
	}

	if err := os.Chdir(projectPath); err != nil {
		err = fmt.Errorf("failed to change directory: %w", err)
		logger.Errorf("切换到项目目录失败: %v", err)
		return err
	}
	defer os.Chdir(originalDir)

	logger.Infof("准备推送 - 目录: %s", projectPath)

	// 调用 git 包执行推送
	if err := git.PushToRemote(context.Background()); err != nil {
		logger.Errorf("PushToRemote 失败: %v", err)
		return err
	}

	logger.Infof("推送成功 - 目录: %s", projectPath)
	return nil
}

// GetStagingStatus 获取项目的暂存区状态
func (a *App) GetStagingStatus(projectPath string) (*git.StagingStatus, error) {
	logger.Infof("[App.GetStagingStatus] 开始获取暂存状态: %s", projectPath)
	if a.initError != nil {
		return nil, a.initError
	}
	status, err := git.GetStagingStatus(projectPath)
	if err != nil {
		logger.Errorf("[App.GetStagingStatus] 获取失败: %v", err)
		return nil, err
	}
	logger.Infof("[App.GetStagingStatus] 获取成功 - staged: %d, unstaged: %d", len(status.Staged), len(status.Unstaged))
	return status, nil
}

// GetFileDiff 获取文件 diff
func (a *App) GetFileDiff(projectPath, filePath string, staged bool) (string, error) {
	if a.initError != nil {
		return "", a.initError
	}
	return git.GetFileDiff(projectPath, filePath, staged)
}

// GetUntrackedFileContent 获取未跟踪文件内容
func (a *App) GetUntrackedFileContent(projectPath, filePath string) (git.FileContentResult, error) {
	logger.Infof("[App.GetUntrackedFileContent] 开始读取未跟踪文件: %s in %s", filePath, projectPath)
	if a.initError != nil {
		logger.Errorf("[App.GetUntrackedFileContent] 初始化错误: %v", a.initError)
		return git.FileContentResult{}, a.initError
	}
	result, err := git.ReadFileContent(projectPath, filePath)
	if err != nil {
		logger.Errorf("[App.GetUntrackedFileContent] 读取失败: %v", err)
	} else {
		logger.Infof("[App.GetUntrackedFileContent] 读取成功, IsBinary: %v, Content长度: %d", result.IsBinary, len(result.Content))
	}
	return result, err
}

// StageFile 暂存文件
func (a *App) StageFile(projectPath, filePath string) error {
	logger.Infof("[App.StageFile] 开始暂存文件: %s in %s", filePath, projectPath)
	if a.initError != nil {
		return a.initError
	}
	err := git.StageFile(projectPath, filePath)
	if err != nil {
		logger.Errorf("[App.StageFile] 暂存失败: %v", err)
	} else {
		logger.Infof("[App.StageFile] 暂存成功: %s", filePath)
	}
	return err
}

// StageAllFiles 暂存所有文件
func (a *App) StageAllFiles(projectPath string) error {
	if a.initError != nil {
		return a.initError
	}
	return git.StageAllFiles(projectPath)
}

// UnstageFile 取消暂存文件
func (a *App) UnstageFile(projectPath, filePath string) error {
	if a.initError != nil {
		return a.initError
	}
	return git.UnstageFile(projectPath, filePath)
}

// UnstageAllFiles 取消暂存所有文件
func (a *App) UnstageAllFiles(projectPath string) error {
	if a.initError != nil {
		return a.initError
	}
	return git.UnstageAllFiles(projectPath)
}

// DiscardFileChanges 还原工作区文件的更改
func (a *App) DiscardFileChanges(projectPath, filePath string) error {
	if a.initError != nil {
		return a.initError
	}
	return git.DiscardFileChanges(projectPath, filePath)
}

// GetUntrackedFiles 获取未跟踪文件列表
func (a *App) GetUntrackedFiles(projectPath string) ([]git.UntrackedFile, error) {
	logger.Infof("[App.GetUntrackedFiles] 开始获取未跟踪文件: %s", projectPath)
	if a.initError != nil {
		return nil, a.initError
	}
	files, err := git.GetUntrackedFiles(projectPath)
	if err != nil {
		logger.Errorf("[App.GetUntrackedFiles] 获取失败: %v", err)
	} else {
		logger.Infof("[App.GetUntrackedFiles] 获取成功，共 %d 个文件", len(files))
	}
	return files, err
}

// StageFiles 添加文件到暂存区
func (a *App) StageFiles(projectPath string, files []string) error {
	logger.Infof("[App.StageFiles] 开始暂存文件: %d 个文件 in %s", len(files), projectPath)
	if a.initError != nil {
		return a.initError
	}
	if len(files) == 0 {
		return fmt.Errorf("文件列表为空")
	}

	args := append([]string{"add"}, files...)
	cmd := git.Command("git", args...)
	cmd.Dir = projectPath

	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Errorf("[App.StageFiles] 暂存失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("添加到暂存区失败: %s\n%w", string(output), err)
	}

	logger.Infof("[App.StageFiles] 暂存成功")
	return nil
}

// AddToGitIgnore 添加到 .gitignore
func (a *App) AddToGitIgnore(projectPath, pattern, mode string) error {
	logger.Infof("[App.AddToGitIgnore] 添加到排除列表: pattern=%s, mode=%s", pattern, mode)
	if a.initError != nil {
		return a.initError
	}

	gitMode := git.ExcludeMode(mode)

	// 目录模式：pattern 已经是用户选择的最终目录路径，直接使用
	// 其他模式：需要根据文件路径生成对应的 gitignore 规则
	var finalPattern string
	var err error

	if gitMode == git.ExcludeModeDirectory {
		// 用户在下拉框中选择的目录已经是最终 pattern
		finalPattern = pattern
	} else {
		finalPattern, err = git.GenerateGitIgnorePattern(pattern, gitMode)
		if err != nil {
			logger.Errorf("[App.AddToGitIgnore] 生成规则失败: %v", err)
			return fmt.Errorf("生成规则失败: %w", err)
		}
	}

	err = git.AddToGitIgnoreFile(projectPath, finalPattern)
	if err != nil {
		logger.Errorf("[App.AddToGitIgnore] 添加失败: %v", err)
	} else {
		logger.Infof("[App.AddToGitIgnore] 添加成功: pattern=%s", finalPattern)
	}
	return err
}

// GetDirectoryOptions 获取目录层级选项
func (a *App) GetDirectoryOptions(filePath string) ([]git.DirectoryOption, error) {
	logger.Infof("[App.GetDirectoryOptions] 获取目录选项: %s", filePath)
	if a.initError != nil {
		return nil, a.initError
	}
	opts := git.GetDirectoryOptions(filePath)
	logger.Infof("[App.GetDirectoryOptions] 获取成功，共 %d 个选项", len(opts))
	return opts, nil
}

// ProjectFullStatus 项目完整状态
type ProjectFullStatus struct {
	GitStatus      *git.ProjectStatus   `json:"gitStatus"`
	StagingStatus  *git.StagingStatus   `json:"stagingStatus"`
	UntrackedCount int                  `json:"untrackedCount"`
	PushoverStatus *pushover.HookStatus `json:"pushoverStatus"`
	PushStatus     *git.PushStatus      `json:"pushStatus"`
	LastUpdated    time.Time            `json:"lastUpdated"`
}

// GetAllProjectStatuses 批量获取多个项目的完整状态
// 使用并发控制，最多同时查询 10 个项目
func (a *App) GetAllProjectStatuses(projectPaths []string) (map[string]*ProjectFullStatus, error) {
	if a.initError != nil {
		return nil, a.initError
	}

	const maxConcurrent = 10

	type result struct {
		path   string
		status *ProjectFullStatus
	}

	sem := make(chan struct{}, maxConcurrent)
	results := make(chan result, len(projectPaths))

	for _, path := range projectPaths {
		sem <- struct{}{}
		go func(p string) {
			defer func() { <-sem }()

			status := &ProjectFullStatus{
				LastUpdated: time.Now(),
			}

			// 获取 Git 状态
			gitStatus, _ := git.GetProjectStatus(context.Background(), p)
			status.GitStatus = gitStatus

			// 获取暂存区状态
			staging, _ := git.GetStagingStatus(p)
			status.StagingStatus = staging

			// 获取未跟踪文件数量
			untracked, _ := git.GetUntrackedFiles(p)
			status.UntrackedCount = len(untracked)

			// 获取 Pushover Hook 状态
			if a.pushoverService != nil {
				pushover, _ := a.pushoverService.GetHookStatus(p)
				status.PushoverStatus = pushover
			}

			// 获取推送状态
			pushStatus, _ := git.GetPushStatus(p)
			status.PushStatus = pushStatus

			results <- result{
				path:   p,
				status: status,
			}
		}(path)
	}

	// 收集所有结果
	statuses := make(map[string]*ProjectFullStatus)
	for i := 0; i < len(projectPaths); i++ {
		r := <-results
		statuses[r.path] = r.status
	}

	return statuses, nil
}

// LogFrontendError 记录前端错误到后端日志
// 接收 JSON 字符串，解析后记录到日志文件
func (a *App) LogFrontendError(errJSON string) error {
	// 检查初始化状态
	if a.initError != nil {
		return fmt.Errorf("app not initialized: %w", a.initError)
	}

	// 检查 errorService 是否已初始化
	if a.errorService == nil {
		return fmt.Errorf("error service not initialized")
	}

	// 调用 ErrorService 的 LogErrorFromJSON 方法
	return a.errorService.LogErrorFromJSON(errJSON)
}
