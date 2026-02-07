package models

import "time"

// UpdateInfo 更新信息
type UpdateInfo struct {
	HasUpdate      bool      `json:"hasUpdate"`      // 是否有更新
	LatestVersion  string    `json:"latestVersion"`  // 最新版本号
	CurrentVersion string    `json:"currentVersion"` // 当前版本号
	ReleaseNotes   string    `json:"releaseNotes"`   // Release notes
	PublishedAt    time.Time `json:"publishedAt"`    // 发布时间
	DownloadURL    string    `json:"downloadURL"`    // 下载链接
	AssetName      string    `json:"assetName"`      // 资源文件名
	Size           int64     `json:"size"`           // 文件大小
	IsPrerelease   bool      `json:"isPrerelease"`   // 是否为预发布版本
	PrereleaseType string    `json:"prereleaseType"` // 预发布类型 (alpha/beta/rc)
}
