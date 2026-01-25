package main

import (
	"context"
	"fmt"
	stdruntime "runtime"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/allanpk716/ai-commit-hub/pkg/git"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/pushover"
	"github.com/allanpk716/ai-commit-hub/pkg/repository"
	"github.com/allanpk716/ai-commit-hub/pkg/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx                  context.Context
	dbPath               string
	gitProjectRepo       *repository.GitProjectRepository
	commitHistoryRepo    *repository.CommitHistoryRepository
	configService        *service.ConfigService
	projectConfigService *service.ProjectConfigService
	pushoverService      *pushover.Service
	initError            error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("AI Commit Hub starting up...")

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get home directory:", err)
		return
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Failed to create config directory:", err)
		return
	}

	a.dbPath = filepath.Join(configDir, "ai-commit-hub.db")

	// Initialize database
	dbConfig := &repository.DatabaseConfig{Path: a.dbPath}
	if err := repository.InitializeDatabase(dbConfig); err != nil {
		a.initError = fmt.Errorf("database initialization failed: %w", err)
		fmt.Println("Failed to initialize database:", err)
		return
	}

	// Initialize repositories (only if database init succeeded)
	a.gitProjectRepo = repository.NewGitProjectRepository()
	a.commitHistoryRepo = repository.NewCommitHistoryRepository()

	// Initialize config service and ensure default config exists
	a.configService = service.NewConfigService()
	if _, err := a.configService.LoadConfig(ctx); err != nil {
		fmt.Println("Failed to initialize config:", err)
		// Continue anyway - config will be created when needed
	}

	// Initialize project config service
	cfg, _ := a.configService.LoadConfig(ctx)
	a.projectConfigService = service.NewProjectConfigService(a.gitProjectRepo, cfg)

	// Run database migrations
	db := repository.GetDB()
	if err := repository.MigrateAddProjectAIConfig(db); err != nil {
		fmt.Printf("数据库迁移失败: %v\n", err)
		// Continue anyway - migration may have already been applied
	}

	// Run Pushover Hook migration
	if err := repository.MigrateAddPushoverHookFields(db); err != nil {
		fmt.Printf("Pushover Hook 迁移失败: %v\n", err)
		// Continue anyway - migration may have already been applied
	}

	// Initialize pushover service
	// 获取可执行文件所在目录作为 appPath
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取可执行文件路径失败: %v\n", err)
	} else {
		appPath := filepath.Dir(execPath)
		a.pushoverService = pushover.NewService(appPath)
	}

	// 同步所有项目的 Hook 状态（阻塞执行，确保前端获取到最新状态）
	if a.pushoverService != nil {
		fmt.Println("准备启动 Hook 状态同步...")
		a.syncAllProjectsHookStatus()
	} else {
		fmt.Println("Pushover service 未初始化，跳过 Hook 状态同步")
	}

	fmt.Println("AI Commit Hub initialized successfully")
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("AI Commit Hub shutting down...")
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
		cmd = exec.Command("explorer", configDir)
	case "darwin":
		cmd = exec.Command("open", configDir)
	default:
		cmd = exec.Command("xdg-open", configDir)
	}

	return cmd.Start()
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
	if a.initError != nil {
		return a.initError
	}

	commitService := service.NewCommitService(a.ctx)
	return commitService.GenerateCommit(projectPath, provider, language)
}

// CommitLocally commits changes to local git repository
func (a *App) CommitLocally(projectPath, message string) error {
	if a.initError != nil {
		return a.initError
	}

	if message == "" {
		return fmt.Errorf("commit 消息不能为空")
	}

	// Save current directory and change to project path
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Use the existing CommitChanges function from git package
	if err := git.CommitChanges(context.Background(), message); err != nil {
		return err
	}

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
			fmt.Printf("同步 Hook 状态失败: %v\n", syncErr)
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
		fmt.Printf("同步 Hook 状态失败: %v\n", syncErr)
		// 不影响卸载结果，只记录错误
	}

	return nil
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

// syncAllProjectsHookStatus 同步所有项目的 Pushover Hook 状态
func (a *App) syncAllProjectsHookStatus() {
	projects, err := a.gitProjectRepo.GetAll()
	if err != nil {
		fmt.Printf("获取项目列表失败: %v\n", err)
		return
	}

	fmt.Printf("开始同步 %d 个项目的 Hook 状态...\n", len(projects))

	for _, project := range projects {
		if err := a.syncProjectHookStatus(&project); err != nil {
			fmt.Printf("同步项目 %s Hook 状态失败: %v\n", project.Name, err)
		}
	}

	fmt.Printf("Hook 状态同步完成\n")
}

// syncProjectHookStatus 同步单个项目的 Hook 状态
func (a *App) syncProjectHookStatus(project *models.GitProject) error {
	fmt.Printf("[DEBUG] 正在检查项目 %s (路径: %s) 的 Hook 状态...\n", project.Name, project.Path)
	status, err := a.pushoverService.GetHookStatus(project.Path)
	if err != nil {
		return fmt.Errorf("获取 Hook 状态失败: %w", err)
	}

	fmt.Printf("[DEBUG] 项目 %s Hook 状态: installed=%v, mode=%s\n", project.Name, status.Installed, status.Mode)
	fmt.Printf("[DEBUG] 数据库中状态: installed=%v, mode=%s\n", project.HookInstalled, project.NotificationMode)

	// 只在状态发生变化时更新数据库
	needsUpdate := project.HookInstalled != status.Installed ||
		(status.Installed && project.NotificationMode != string(status.Mode))

	if !needsUpdate {
		fmt.Printf("[DEBUG] 项目 %s 状态无需更新\n", project.Name)
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

	fmt.Printf("已更新项目 %s 的 Hook 状态: installed=%v, mode=%s\n",
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
			"name":            project.Name,
			"path":            project.Path,
			"db_hook_installed": project.HookInstalled,
			"db_notification_mode": project.NotificationMode,
			"db_hook_version":  project.HookVersion,
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
