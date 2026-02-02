package models

import "time"

// UpdatePreferences 用户更新偏好设置
type UpdatePreferences struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SkippedVersion string   `gorm:"index" json:"skippedVersion"` // 用户跳过的版本号
	SkipReason      string   `json:"skipReason"`                  // 跳过原因 (not_now/this_version)
	CreatedAt      time.Time `json:"createdAt"`                   // 跳过时间
	LastCheckTime  time.Time `json:"lastCheckTime"`               // 最后检查更新的时间
	AutoCheck      bool     `json:"autoCheck"`                   // 是否自动检查（默认 true）
}
