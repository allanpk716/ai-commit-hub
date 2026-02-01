package service

import (
	"encoding/json"
	"time"

	"github.com/WQGroup/logger"
)

// FrontendError 前端错误结构
type FrontendError struct {
	Type      string    `json:"type"`      // "error" | "warning"
	Message   string    `json:"message"`   // 简短消息
	Details   string    `json:"details"`   // 详细信息
	Source    string    `json:"source"`    // 来源组件
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ErrorService 错误服务
type ErrorService struct{}

// NewErrorService 创建错误服务实例
func NewErrorService() *ErrorService {
	return &ErrorService{}
}

// LogError 记录前端错误到日志文件
func (s *ErrorService) LogError(err FrontendError) error {
	// 格式化时间戳
	timestamp := err.Timestamp.Format("2006-01-02 15:04:05")

	// 根据错误类型记录日志
	if err.Type == "error" {
		// 错误级别
		logger.Errorf("[Frontend Error] %s | %s | %s\n  Message: %s\n  Details: %s",
			timestamp,
			err.Source,
			err.Type,
			err.Message,
			err.Details,
		)
	} else if err.Type == "warning" {
		// 警告级别
		logger.Warnf("[Frontend Warning] %s | %s | %s\n  Message: %s\n  Details: %s",
			timestamp,
			err.Source,
			err.Type,
			err.Message,
			err.Details,
		)
	} else {
		// 默认使用 info 级别
		logger.Infof("[Frontend Log] %s | %s | %s\n  Message: %s\n  Details: %s",
			timestamp,
			err.Source,
			err.Type,
			err.Message,
			err.Details,
		)
	}

	return nil
}

// LogErrorFromJSON 从 JSON 字符串解析并记录错误（用于 Wails 绑定）
func (s *ErrorService) LogErrorFromJSON(errJSON string) error {
	var err FrontendError
	if err := json.Unmarshal([]byte(errJSON), &err); err != nil {
		return err
	}
	return s.LogError(err)
}
