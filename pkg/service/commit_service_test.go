package service

import (
	"context"
	"errors"
	"testing"

	"github.com/allanpk716/ai-commit-hub/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestMockAIClient_Basic(t *testing.T) {
	client := NewMockAIClient("test response", nil)

	response, err := client.GetCommitMessage(context.Background(), "test prompt")

	assert.NoError(t, err)
	assert.Equal(t, "test response", response)
}

func TestMockAIClient_WithError(t *testing.T) {
	expectedErr := errors.New("API error")
	client := NewMockAIClientWithError(expectedErr)

	response, err := client.GetCommitMessage(context.Background(), "test prompt")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, response)
}

func TestMockAIClient_Streaming(t *testing.T) {
	deltas := []string{"Hello", " ", "World"}
	client := NewMockAIClient("", deltas)

	var received []string
	full, err := client.StreamCommitMessage(context.Background(), "prompt", func(delta string) {
		received = append(received, delta)
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", full) // 返回完整消息
	assert.Equal(t, deltas, received)
}

func TestCommitService_GenerateCommit_EmptyDiff(t *testing.T) {
	// 跳过：需要注册 mock provider 到 registry
	// CommitService 直接使用 registry.Get() 获取 provider
	// 需要修改架构支持注入 mock 或修改 registry 支持测试模式
	t.Skip("需要 mock AI provider registry - 暂时跳过")

	repo := helpers.SetupTestRepo(t)
	service := NewCommitService(context.Background())

	// 没有暂存变更
	err := service.GenerateCommit(repo.Path, "mock", "zh")

	// 应该返回 nil（空 diff 不算错误）
	assert.NoError(t, err)
}

func TestCommitService_GenerateCommit_RealDiff(t *testing.T) {
	repo := helpers.SetupTestRepo(t)
	repo.CreateStagedChange(t, "test.txt", "test content")

	_ = NewCommitService(context.Background())

	// 注意：这个测试会尝试连接真实的 AI Provider
	// 如果没有配置 API Key，会失败
	// 我们可以标记为跳过，或者只验证函数调用不出错
	t.Skip("需要真实配置 AI Provider 或 mock registry")
}

func TestCommitService_GenerateCommit_WithMock(t *testing.T) {
	// 这个测试需要 mock AI provider registry
	// 由于当前的 CommitService 直接使用 registry.Get()
	// 我们需要修改架构来支持注入 mock，或者跳过这个测试

	t.Skip("需要 mock AI provider registry - 暂时跳过")
}
