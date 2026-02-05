package main

import (
	"errors"
	"testing"
	"time"

	apperrors "github.com/allanpk716/ai-commit-hub/pkg/errors"
	"github.com/allanpk716/ai-commit-hub/pkg/constants"
)

func TestTimingConstants(t *testing.T) {
	// 测试延迟常量值是否正确
	tests := []struct {
		name     string
		constant time.Duration
		expected time.Duration
	}{
		{
			name:     "SystrayInitDelay",
			constant: constants.SystrayInitDelay,
			expected: 300 * time.Millisecond,
		},
		{
			name:     "IconSettleDelay",
			constant: constants.IconSettleDelay,
			expected: 150 * time.Millisecond,
		},
		{
			name:     "IconRetryDelay",
			constant: constants.IconRetryDelay,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "WindowShowDelay",
			constant: constants.WindowShowDelay,
			expected: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.constant)
			}
		})
	}
}

func TestConcurrentOpsConstants(t *testing.T) {
	// 测试并发操作常量
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{
			name:     "DefaultMaxConcurrentOps",
			constant: constants.DefaultMaxConcurrentOps,
			expected: 10,
		},
		{
			name:     "LowCPUMaxConcurrentOps",
			constant: constants.LowCPUMaxConcurrentOps,
			expected: 5,
		},
		{
			name:     "MaxIconRetryAttempts",
			constant: constants.MaxIconRetryAttempts,
			expected: 5,
		},
		{
			name:     "StatusCacheTTLSec",
			constant: constants.StatusCacheTTLSec,
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, tt.constant)
			}
		})
	}
}

func TestAppInitError(t *testing.T) {
	// 测试 AppInitError 错误类型
	testErr := errors.New("database connection failed")

	appErr := apperrors.NewAppInitError(testErr)

	errorMsg := appErr.Error()
	if errorMsg == "" {
		t.Error("Error message should not be empty")
	}

	// 检查错误消息包含原始错误信息
	if !contains(errorMsg, "database connection failed") {
		t.Errorf("Error message should contain original error, got: %s", errorMsg)
	}

	// 测试 Unwrap
	if appErr.Unwrap() != testErr {
		t.Error("Unwrap should return original error")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCheckInit(t *testing.T) {
	// 测试 CheckInit 辅助函数
	t.Run("无错误时返回 nil", func(t *testing.T) {
		err := apperrors.CheckInit(nil)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("有错误时返回 AppInitError", func(t *testing.T) {
		originalErr := errors.New("init failed")
		err := apperrors.CheckInit(originalErr)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		appErr, ok := err.(*apperrors.AppInitError)
		if !ok {
			t.Error("Expected AppInitError type")
		}

		if appErr.Unwrap() != originalErr {
			t.Error("Unwrap should return original error")
		}
	})
}
