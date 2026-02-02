package service

import (
	"testing"
)

func TestNewUpdateService(t *testing.T) {
	service := NewUpdateService("allanpk716/ai-commit-hub")

	if service == nil {
		t.Fatal("NewUpdateService() 返回 nil")
	}

	if service.repo != "allanpk716/ai-commit-hub" {
		t.Errorf("repo = %s, want allanpk716/ai-commit-hub", service.repo)
	}

	if service.httpClient == nil {
		t.Error("httpClient 为 nil")
	}
}

func TestCheckForUpdates(t *testing.T) {
	// 跳过测试如果在 CI 环境中没有网络访问
	t.Skip("需要网络访问，跳过自动化测试")

	service := NewUpdateService("allanpk716/ai-commit-hub")

	info, err := service.CheckForUpdates()
	if err != nil {
		t.Logf("CheckForUpdates 失败（可能在离线环境）: %v", err)
		return
	}

	if info == nil {
		t.Fatal("info 为 nil")
	}

	if info.CurrentVersion == "" {
		t.Error("CurrentVersion 为空")
	}

	t.Logf("更新信息: HasUpdate=%v, Current=%s, Latest=%s",
		info.HasUpdate, info.CurrentVersion, info.LatestVersion)
}

func TestCompareVersions(t *testing.T) {
	service := NewUpdateService("test/repo")

	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{"有更新", "v2.0.0", "v1.0.0", true},
		{"无更新", "v1.0.0", "v1.0.0", false},
		{"旧版本", "v1.0.0", "v2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.compareVersions(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %v, want %v",
					tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}

func TestFindPlatformAsset(t *testing.T) {
	service := NewUpdateService("test/repo")

	assets := []Asset{
		{Name: "ai-commit-hub-1.0.0-windows.zip", Size: 1000, URL: "http://example.com/windows.zip"},
		{Name: "ai-commit-hub-1.0.0-darwin.zip", Size: 2000, URL: "http://example.com/darwin.zip"},
		{Name: "ai-commit-hub-1.0.0-linux.zip", Size: 1500, URL: "http://example.com/linux.zip"},
	}

	// 测试当前平台
	name, url := service.findPlatformAsset(assets)

	if name == "" {
		t.Error("未找到对应平台的资源")
	}

	if url == "" {
		t.Error("下载 URL 为空")
	}

	t.Logf("找到资源: %s -> %s", name, url)
}

func TestGetAssetSize(t *testing.T) {
	service := NewUpdateService("test/repo")

	assets := []Asset{
		{Name: "test.zip", Size: 1024},
		{Name: "other.zip", Size: 2048},
	}

	size := service.getAssetSize(assets, "test.zip")
	if size != 1024 {
		t.Errorf("getAssetSize() = %d, want 1024", size)
	}

	// 测试不存在的资源
	size = service.getAssetSize(assets, "notfound.zip")
	if size != 0 {
		t.Errorf("getAssetSize() (不存在) = %d, want 0", size)
	}
}
