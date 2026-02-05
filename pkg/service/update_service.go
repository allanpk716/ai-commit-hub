package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/version"
)

// UpdateService 更新检查服务
type UpdateService struct {
	repo       string
	httpClient *http.Client
}

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset Release 资源
type Asset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	URL  string `json:"browser_download_url"`
}

// NewUpdateService 创建更新检查服务
func NewUpdateService(repo string) *UpdateService {
	return &UpdateService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckForUpdates 检查更新
func (s *UpdateService) CheckForUpdates() (*models.UpdateInfo, error) {
	logger.Info("检查更新", "repo", s.repo)

	// 获取最新 Release
	release, err := s.fetchLatestRelease()
	if err != nil {
		logger.Warnf("获取 Release 失败: %v", err)
		return nil, err
	}

	currentVersion := version.GetVersion()
	latestVersion := release.TagName

	logger.Info("版本信息", "current", currentVersion, "latest", latestVersion)

	// 比较版本
	hasUpdate := s.compareVersions(latestVersion, currentVersion)

	// 找到对应平台的资源
	assetName, downloadURL := s.findPlatformAsset(release.Assets)

	info := &models.UpdateInfo{
		HasUpdate:      hasUpdate,
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		ReleaseNotes:   release.Body,
		PublishedAt:    s.parseTime(release.PublishedAt),
		DownloadURL:    downloadURL,
		AssetName:      assetName,
		Size:           s.getAssetSize(release.Assets, assetName),
	}

	logger.Infof("更新检查完成: hasUpdate=%v", info.HasUpdate)
	return info, nil
}

// fetchLatestRelease 获取最新 Release
func (s *UpdateService) fetchLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", s.repo)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return &release, nil
}

// compareVersions 比较版本号
func (s *UpdateService) compareVersions(latest, current string) bool {
	result := version.CompareVersions(latest, current)
	return result > 0
}

// findPlatformAsset 找到对应平台的资源
func (s *UpdateService) findPlatformAsset(assets []Asset) (name, url string) {
	// 仅支持 Windows 平台
	const targetOS = "windows"

	for _, asset := range assets {
		if strings.Contains(asset.Name, targetOS) {
			return asset.Name, asset.URL
		}
	}

	return "", ""
}

// getAssetSize 获取资源大小
func (s *UpdateService) getAssetSize(assets []Asset, assetName string) int64 {
	for _, asset := range assets {
		if asset.Name == assetName {
			return asset.Size
		}
	}
	return 0
}

// parseTime 解析时间
func (s *UpdateService) parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		logger.Warnf("解析时间失败: %v", err)
		return time.Time{}
	}
	return t
}
