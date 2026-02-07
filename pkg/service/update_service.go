package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/WQGroup/logger"
	"github.com/allanpk716/ai-commit-hub/pkg/models"
	"github.com/allanpk716/ai-commit-hub/pkg/version"
	"golang.org/x/mod/semver"
)

// UpdateService 更新检查服务
type UpdateService struct {
	repo         string
	httpClient   *http.Client
	mu           sync.RWMutex
	lastCheck    time.Time
	cachedResult *models.UpdateInfo
}

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string `json:"published_at"`
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

	// 检查缓存（24小时内）
	s.mu.RLock()
	if s.cachedResult != nil && time.Since(s.lastCheck) < 24*time.Hour {
		logger.Info("使用缓存的更新检查结果")
		cached := s.cachedResult
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// 获取所有 Releases
	releases, err := s.fetchAllReleases()
	if err != nil {
		// 如果是速率限制错误，尝试返回缓存
		if s.isRateLimitError(err) && s.cachedResult != nil {
			logger.Warnf("遇到速率限制错误，返回缓存结果: %v", err)
			s.mu.RLock()
			cached := s.cachedResult
			s.mu.RUnlock()
			return cached, nil
		}
		logger.Warnf("获取 Releases 失败: %v", err)
		return nil, err
	}

	// 找到最新版本
	latestRelease := s.findLatestVersion(releases)
	if latestRelease == nil {
		logger.Warn("未找到有效的 Release")
		return nil, fmt.Errorf("未找到有效的 Release")
	}

	currentVersion := version.GetVersion()
	latestVersion := latestRelease.TagName

	logger.Info("版本信息", "current", currentVersion, "latest", latestVersion)

	// 比较版本
	hasUpdate := s.compareVersions(latestVersion, currentVersion)

	// 找到对应平台的资源
	assetName, downloadURL := s.findPlatformAsset(latestRelease.Assets)

	info := &models.UpdateInfo{
		HasUpdate:      hasUpdate,
		LatestVersion:  latestVersion,
		CurrentVersion: currentVersion,
		ReleaseNotes:   latestRelease.Body,
		PublishedAt:    s.parseTime(latestRelease.PublishedAt),
		DownloadURL:    downloadURL,
		AssetName:      assetName,
		Size:           s.getAssetSize(latestRelease.Assets, assetName),
		IsPrerelease:   latestRelease.Prerelease,
		PrereleaseType: s.extractPrereleaseType(latestRelease.TagName),
	}

	// 更新缓存
	s.mu.Lock()
	s.cachedResult = info
	s.lastCheck = time.Now()
	s.mu.Unlock()

	logger.Infof("更新检查完成: hasUpdate=%v", info.HasUpdate)
	return info, nil
}

// fetchAllReleases 获取所有 Releases
func (s *UpdateService) fetchAllReleases() ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", s.repo)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误: %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 过滤掉 draft releases
	var filtered []GitHubRelease
	for _, r := range releases {
		if !r.Draft {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

// findLatestVersion 从所有 releases 中找到最新版本
func (s *UpdateService) findLatestVersion(releases []GitHubRelease) *GitHubRelease {
	if len(releases) == 0 {
		return nil
	}

	// 使用 semver.Sort 排序
	var versions []string
	releaseMap := make(map[string]*GitHubRelease)

	for i := range releases {
		tagName := releases[i].TagName
		normalized := version.NormalizeVersion(tagName)
		if semver.IsValid(normalized) {
			versions = append(versions, normalized)
			releaseMap[normalized] = &releases[i]
		}
	}

	if len(versions) == 0 {
		// 如果没有有效的 semver 版本，返回第一个非 draft release
		for i := range releases {
			if !releases[i].Draft {
				return &releases[i]
			}
		}
		return nil
	}

	// 降序排序（最新的在前）
	semver.Sort(versions)
	if len(versions) > 0 {
		latest := versions[len(versions)-1]
		return releaseMap[latest]
	}

	return nil
}

// compareVersions 比较版本号
func (s *UpdateService) compareVersions(latest, current string) bool {
	result := version.CompareVersions(latest, current)
	return result > 0
}

// isRateLimitError 检查是否为速率限制错误
func (s *UpdateService) isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	// 检查是否包含速率限制相关信息
	errStr := err.Error()
	return strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "timeout")
}

// extractPrereleaseType 从版本号中提取预发布类型
func (s *UpdateService) extractPrereleaseType(tagName string) string {
	if strings.Contains(tagName, "alpha") {
		return "alpha"
	}
	if strings.Contains(tagName, "beta") {
		return "beta"
	}
	if strings.Contains(tagName, "rc") {
		return "rc"
	}
	return ""
}

// findPlatformAsset 找到对应平台的资源（精确匹配 windows-amd64）
func (s *UpdateService) findPlatformAsset(assets []Asset) (name, url string) {
	// 精确匹配 windows-amd64
	const targetPlatform = "windows-amd64"

	for _, asset := range assets {
		if strings.Contains(asset.Name, targetPlatform) {
			return asset.Name, asset.URL
		}
	}

	logger.Warnf("未找到 %s 平台的资源", targetPlatform)
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

// StartBackgroundCheck 启动后台定时检查（每24小时）
func (s *UpdateService) StartBackgroundCheck() {
	logger.Info("启动后台更新检查服务（24小时间隔）")

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		// 立即执行一次检查
		s.performBackgroundCheck()

		// 然后定时执行
		for range ticker.C {
			s.performBackgroundCheck()
		}
	}()
}

// performBackgroundCheck 执行后台检查（不返回错误，只记录日志）
func (s *UpdateService) performBackgroundCheck() {
	logger.Info("执行后台更新检查")
	_, err := s.CheckForUpdates()
	if err != nil {
		logger.Warnf("后台更新检查失败: %v", err)
	}
}
