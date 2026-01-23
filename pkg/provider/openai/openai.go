package openai

import (
    openaic "github.com/allanpk716/ai-commit-hub/pkg/aicommit/provider/openai_compat"
)

// NewOpenAIClient returns an OpenAI-compatible client powered by the official SDK.
// It reuses the generic compat client to avoid duplication.
func NewOpenAIClient(provider, key, model, baseURL string) *openaic.Client {
    return openaic.NewCompatClient(provider, key, model, baseURL)
}
