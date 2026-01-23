package helpers

import (
	"bytes"
	"io"
	"log"
	"strings"
	"testing"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogCapture 捕获日志输出
type LogCapture struct {
	buffer *bytes.Buffer
	logger *log.Logger
}

// NewLogCapture 创建日志捕获器
func NewLogCapture(t *testing.T) *LogCapture {
	t.Helper()

	buffer := &bytes.Buffer{}

	// 配置 logger 输出到 buffer
	logger := log.New(buffer, "", log.LstdFlags)

	return &LogCapture{
		buffer: buffer,
		logger: logger,
	}
}

// GetLogs 获取所有日志
func (lc *LogCapture) GetLogs() string {
	return lc.buffer.String()
}

// GetLogsByLevel 按级别获取日志
func (lc *LogCapture) GetLogsByLevel(level string) []string {
	logs := strings.Split(lc.buffer.String(), "\n")
	var filtered []string

	for _, log := range logs {
		if strings.Contains(log, "["+level+"]") || strings.Contains(log, strings.ToUpper(level)) {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// Contains 验证日志包含特定内容
func (lc *LogCapture) Contains(substring string) bool {
	return strings.Contains(lc.buffer.String(), substring)
}

// ContainsError 验证有错误日志
func (lc *LogCapture) ContainsError() bool {
	logs := lc.buffer.String()
	return strings.Contains(logs, "[ERROR]") || strings.Contains(logs, "ERROR") ||
		strings.Contains(logs, "[WARN]") || strings.Contains(logs, "WARN")
}

// GetLogger 获取 logger 实例
func (lc *LogCapture) GetLogger() *log.Logger {
	return lc.logger
}

// Debug 记录调试日志
func (lc *LogCapture) Debug(msg string) {
	lc.logger.Printf("[DEBUG] %s", msg)
}

// Info 记录信息日志
func (lc *LogCapture) Info(msg string) {
	lc.logger.Printf("[INFO] %s", msg)
}

// Warn 记录警告日志
func (lc *LogCapture) Warn(msg string) {
	lc.logger.Printf("[WARN] %s", msg)
}

// Error 记录错误日志
func (lc *LogCapture) Error(msg string) {
	lc.logger.Printf("[ERROR] %s", msg)
}

// SetOutput 设置日志输出目标（用于测试）
func (lc *LogCapture) SetOutput(w io.Writer) {
	lc.logger.SetOutput(w)
}

// NewLogCaptureWithLogger 创建带自定义 logger 的捕获器
func NewLogCaptureWithLogger(t *testing.T, logger *log.Logger) *LogCapture {
	t.Helper()

	buffer := &bytes.Buffer{}
	logger.SetOutput(buffer)

	return &LogCapture{
		buffer: buffer,
		logger: logger,
	}
}
