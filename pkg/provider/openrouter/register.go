package openrouter

import (
	"context"

	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	compat "github.com/allanpk716/ai-commit-hub/pkg/aicommit/provider/openai_compat"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
)

const ProviderName = "openrouter"

func factory(ctx context.Context, name string, ps config.ProviderSettings) (ai.AIClient, error) {
	// OpenRouter is OpenAI-compatible; reuse the compat client.
	return compat.NewCompatClient(name, ps.APIKey, ps.Model, ps.BaseURL), nil
}

func init() {
	registry.Register(ProviderName, factory)
	registry.RegisterDefaults(ProviderName, config.ProviderSettings{Model: "openrouter/auto", BaseURL: "https://openrouter.ai/api/v1"})
	registry.SetRequiresAPIKey(ProviderName, true)
}
