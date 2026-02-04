package constants

import "time"

const (
	// Systram 相关延迟
	SystrayInitDelay     = 300 * time.Millisecond
	IconSettleDelay      = 150 * time.Millisecond
	IconRetryDelay       = 200 * time.Millisecond
	MaxIconRetryAttempts = 5

	// Git 操作并发限制
	DefaultMaxConcurrentOps = 10
	LowCPUMaxConcurrentOps  = 5

	// 状态缓存
	StatusCacheTTLSec = 30

	// 窗口操作延迟
	WindowShowDelay = 100 * time.Millisecond
)
