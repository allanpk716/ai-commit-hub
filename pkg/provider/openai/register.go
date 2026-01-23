package openai

import (
	"context"

	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/provider/registry"
)

const ProviderName = "openai"

func factory(ctx context.Context, name string, ps config.ProviderSettings) (ai.AIClient, error) {
    // No ctx usage needed for OpenAI client construction.
    return NewOpenAIClient(name, ps.APIKey, ps.Model, ps.BaseURL), nil
}

func init() {
    registry.Register(ProviderName, factory)
    registry.RegisterDefaults(ProviderName, config.ProviderSettings{Model: "chatgpt-4o-latest", BaseURL: "https://api.openai.com/v1"})
    registry.SetRequiresAPIKey(ProviderName, true)
}
