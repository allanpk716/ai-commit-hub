package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogCapture(t *testing.T) {
	capture := NewLogCapture(t)

	assert.NotNil(t, capture)
	assert.NotNil(t, capture.buffer)
	assert.NotNil(t, capture.logger)
}

func TestLogCapture_Contains(t *testing.T) {
	capture := NewLogCapture(t)

	capture.Info("test message")

	assert.True(t, capture.Contains("test message"))
	assert.False(t, capture.Contains("not found"))
}

func TestLogCapture_ContainsError(t *testing.T) {
	capture := NewLogCapture(t)

	capture.Debug("debug message")
	assert.False(t, capture.ContainsError())

	capture.Error("error message")
	assert.True(t, capture.ContainsError())
}

func TestLogCapture_GetLogsByLevel(t *testing.T) {
	capture := NewLogCapture(t)

	capture.Debug("debug msg")
	capture.Info("info msg")
	capture.Error("error msg")

	logs := capture.GetLogs()
	assert.Contains(t, logs, "[DEBUG]")
	assert.Contains(t, logs, "[INFO]")
	assert.Contains(t, logs, "[ERROR]")
}

func TestLogCapture_LogLevels(t *testing.T) {
	capture := NewLogCapture(t)

	capture.Debug("debug")
	capture.Info("info")
	capture.Warn("warning")
	capture.Error("error")

	logs := capture.GetLogs()
	assert.Contains(t, logs, "[DEBUG] debug")
	assert.Contains(t, logs, "[INFO] info")
	assert.Contains(t, logs, "[WARN] warning")
	assert.Contains(t, logs, "[ERROR] error")
}

func TestLogCapture_GetLogger(t *testing.T) {
	capture := NewLogCapture(t)

	logger := capture.GetLogger()
	assert.NotNil(t, logger)

	// 测试写入
	logger.Println("direct log")
	assert.True(t, capture.Contains("direct log"))
}
