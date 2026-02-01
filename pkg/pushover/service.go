package pushover

import (
	"fmt"
	"os"
	"path/filepath"
)

// Service Pushover 服务
type Service struct {
	repoManager *RepositoryManager
	installer   *Installer
}

// NewService 创建 Pushover 服务
func NewService(appPath string) *Service {
	extensionsPath := filepath.Join(appPath, "extensions")
	repoManager := NewRepositoryManager(extensionsPath)
	extensionPath := repoManager.GetExtensionPath()

	return &Service{
		repoManager: repoManager,
		installer:   NewInstaller(extensionPath),
	}
}

// CheckHookInstalled 检查项目的 Hook 是否已安装
func (s *Service) CheckHookInstalled(projectPath string) bool {
	checker := NewStatusChecker(projectPath)
	return checker.CheckInstalled()
}

// GetHookStatus 获取项目的 Hook 状态
func (s *Service) GetHookStatus(projectPath string) (*HookStatus, error) {
	checker := NewStatusChecker(projectPath)

	// 获取扩展最新版本（用于比较）
	latestVersion := ""
	if s.repoManager.IsCloned() {
		if v, err := s.repoManager.GetVersion(); err == nil {
			latestVersion = v
		}
	}

	return checker.GetStatus(latestVersion)
}

// InstallHook 为项目安装 Hook
func (s *Service) InstallHook(projectPath string, force bool) (*InstallResult, error) {
	// 检查项目路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	// 检查扩展是否已下载
	if !s.repoManager.IsCloned() {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	return s.installer.Install(projectPath, force)
}

// UninstallHook 卸载项目的 Hook
func (s *Service) UninstallHook(projectPath string) error {
	return s.installer.Uninstall(projectPath)
}

// UpdateHook 更新项目的 Hook
func (s *Service) UpdateHook(projectPath string) (*InstallResult, error) {
	// 检查项目路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	// 检查扩展是否已下载
	if !s.repoManager.IsCloned() {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	return s.installer.Update(projectPath)
}

// ReinstallHook 重装项目的 Hook（保留用户配置）
func (s *Service) ReinstallHook(projectPath string) (*InstallResult, error) {
	// 检查项目路径是否存在
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	// 检查扩展是否已下载
	if !s.repoManager.IsCloned() {
		return nil, fmt.Errorf("cc-pushover-hook 扩展未下载，请先下载扩展")
	}

	// 检查项目是否已安装 Hook
	if !s.CheckHookInstalled(projectPath) {
		return nil, fmt.Errorf("项目未安装 Pushover Hook，请先安装")
	}

	return s.installer.Reinstall(projectPath)
}

// SetNotificationMode 设置项目的通知模式
func (s *Service) SetNotificationMode(projectPath string, mode NotificationMode) error {
	return s.installer.SetNotificationMode(projectPath, mode)
}

// CloneExtension 克隆 cc-pushover-hook 扩展
func (s *Service) CloneExtension() error {
	return s.repoManager.Clone()
}

// UpdateExtension 更新 cc-pushover-hook 扩展
func (s *Service) UpdateExtension() error {
	return s.repoManager.Update()
}

// GetExtensionInfo 获取扩展信息
func (s *Service) GetExtensionInfo() (*ExtensionInfo, error) {
	return s.repoManager.GetExtensionInfo()
}

// IsExtensionDownloaded 检查扩展是否已下载
func (s *Service) IsExtensionDownloaded() bool {
	return s.repoManager.IsCloned()
}

// GetExtensionVersion 获取扩展版本
func (s *Service) GetExtensionVersion() (string, error) {
	if !s.repoManager.IsCloned() {
		return "", fmt.Errorf("扩展未下载")
	}
	return s.repoManager.GetVersion()
}

// RecloneExtension 删除并重新下载扩展
func (s *Service) RecloneExtension() error {
	return s.repoManager.Reclone()
}

// GetExtensionPath 获取扩展目录路径
func (s *Service) GetExtensionPath() string {
	return s.repoManager.GetExtensionPath()
}

// CheckForUpdates 检查是否有可用更新
// 返回: (是否需要更新, 当前版本, 最新版本, 错误)
func (s *Service) CheckForUpdates() (bool, string, string, error) {
	// 检查扩展是否已下载
	if !s.repoManager.IsCloned() {
		return false, "", "", fmt.Errorf("扩展未下载")
	}

	// 获取当前版本和最新版本
	currentVersion, err := s.repoManager.GetVersion()
	if err != nil {
		return false, "", "", fmt.Errorf("获取当前版本失败: %w", err)
	}

	latestVersion, err := s.repoManager.GetLatestVersion()
	if err != nil {
		return false, "", "", fmt.Errorf("获取最新版本失败: %w", err)
	}

	// 比较版本
	needsUpdate := CompareVersions(currentVersion, latestVersion) < 0

	return needsUpdate, currentVersion, latestVersion, nil
}
