package ollama

import (
	"context"

	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/provider/registry"
)

const ProviderName = "ollama"

func factory(ctx context.Context, name string, ps config.ProviderSettings) (ai.AIClient, error) {
    return NewOllamaClient(name, ps.BaseURL, ps.Model)
}

func init() {
    registry.Register(ProviderName, factory)
    registry.RegisterDefaults(ProviderName, config.ProviderSettings{Model: "llama2", BaseURL: "http://localhost:11434"})
    registry.SetRequiresAPIKey(ProviderName, false)
}
