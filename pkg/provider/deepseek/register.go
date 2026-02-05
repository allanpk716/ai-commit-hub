package deepseek

import (
	"context"

	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
)

const ProviderName = "deepseek"

func factory(ctx context.Context, name string, ps config.ProviderSettings) (ai.AIClient, error) {
	return NewDeepseekClient(name, ps.APIKey, ps.Model, ps.BaseURL)
}

func init() {
	registry.Register(ProviderName, factory)
	registry.RegisterDefaults(ProviderName, config.ProviderSettings{Model: "deepseek-chat", BaseURL: "https://api.deepseek.com/v1"})
	registry.SetRequiresAPIKey(ProviderName, true)
}
