package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorService_LogError(t *testing.T) {
	service := NewErrorService()

	t.Run("记录错误", func(t *testing.T) {
		err := FrontendError{
			Type:      "error",
			Message:   "测试错误",
			Details:   "这是错误详情",
			Source:    "TestComponent",
			Timestamp: time.Now(),
		}

		// 不应该返回错误
		assert.NoError(t, service.LogError(err))
	})

	t.Run("记录警告", func(t *testing.T) {
		err := FrontendError{
			Type:      "warning",
			Message:   "测试警告",
			Details:   "这是警告详情",
			Source:    "TestComponent",
			Timestamp: time.Now(),
		}

		// 不应该返回错误
		assert.NoError(t, service.LogError(err))
	})

	t.Run("记录其他类型", func(t *testing.T) {
		err := FrontendError{
			Type:      "info",
			Message:   "测试信息",
			Details:   "这是信息详情",
			Source:    "TestComponent",
			Timestamp: time.Now(),
		}

		// 不应该返回错误
		assert.NoError(t, service.LogError(err))
	})
}

func TestErrorService_LogErrorFromJSON(t *testing.T) {
	service := NewErrorService()

	t.Run("有效的 JSON", func(t *testing.T) {
		fe := FrontendError{
			Type:      "error",
			Message:   "测试错误",
			Details:   "这是错误详情",
			Source:    "TestComponent",
			Timestamp: time.Now(),
		}

		// 序列化为 JSON
		errJSON, marshalErr := json.Marshal(fe)
		assert.NoError(t, marshalErr)

		// 从 JSON 记录
		assert.NoError(t, service.LogErrorFromJSON(string(errJSON)))
	})

	t.Run("无效的 JSON", func(t *testing.T) {
		// 无效的 JSON 字符串
		errJSON := "{invalid json"

		// 应该返回错误
		assert.Error(t, service.LogErrorFromJSON(errJSON))
	})

	t.Run("空 JSON", func(t *testing.T) {
		// 空的 JSON 字符串
		errJSON := "{}"

		// 应该不返回错误（虽然有默认值）
		assert.NoError(t, service.LogErrorFromJSON(errJSON))
	})
}

func TestFrontendError_JSON(t *testing.T) {
	fe := FrontendError{
		Type:      "error",
		Message:   "测试错误",
		Details:   "这是错误详情",
		Source:    "TestComponent",
		Timestamp: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
	}

	// 测试序列化
	data, err := json.Marshal(fe)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "测试错误")
	assert.Contains(t, string(data), "TestComponent")

	// 测试反序列化
	var decoded FrontendError
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, fe.Message, decoded.Message)
	assert.Equal(t, fe.Source, decoded.Source)
	assert.Equal(t, fe.Type, decoded.Type)
}
