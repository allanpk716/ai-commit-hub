package phind

import (
	"context"

	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/ai"
	"github.com/allanpk716/ai-commit-hub/pkg/aicommit/config"
	"github.com/allanpk716/ai-commit-hub/pkg/provider/registry"
)

const ProviderName = "phind"

func factory(ctx context.Context, name string, ps config.ProviderSettings) (ai.AIClient, error) {
    return NewPhindClient(name, ps.APIKey, ps.Model, ps.BaseURL)
}

func init() {
    registry.Register(ProviderName, factory)
    registry.RegisterDefaults(ProviderName, config.ProviderSettings{Model: "Phind-70B", BaseURL: "https://https.extension.phind.com/agent/"})
    registry.SetRequiresAPIKey(ProviderName, false)
}
