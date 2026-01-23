package service

import (
	"context"
)

// MockAIClient 是用于测试的 Mock AI Client
type MockAIClient struct {
	Response string
	Error    error
	Deltas   []string
}

func (m *MockAIClient) GetCommitMessage(ctx context.Context, prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

func (m *MockAIClient) StreamCommitMessage(ctx context.Context, prompt string, deltaFunc func(string)) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}

	full := ""
	for _, delta := range m.Deltas {
		deltaFunc(delta)
		full += delta
	}
	return full, nil
}

// NewMockAIClient 创建新的 Mock AI Client
func NewMockAIClient(response string, deltas []string) *MockAIClient {
	return &MockAIClient{
		Response: response,
		Deltas:   deltas,
	}
}

// NewMockAIClientWithError 创建返回错误的 Mock AI Client
func NewMockAIClientWithError(err error) *MockAIClient {
	return &MockAIClient{Error: err}
}
