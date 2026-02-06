package models

import "gorm.io/gorm"

// WindowState 窗口状态模型
type WindowState struct {
	gorm.Model
	Key       string `gorm:"uniqueIndex;not null;size:100"` // 配置键(如 "window.main")
	X         int    `gorm:"not null"`                       // 窗口 X 坐标
	Y         int    `gorm:"not null"`                       // 窗口 Y 坐标
	Width     int    `gorm:"not null"`                       // 窗口宽度
	Height    int    `gorm:"not null"`                       // 窗口高度
	Maximized bool   `gorm:"default:false"`                  // 是否最大化
	MonitorID string `gorm:"size:50"`                        // 显示器编号(可选)
}

// TableName 指定表名
func (WindowState) TableName() string {
	return "window_states"
}
