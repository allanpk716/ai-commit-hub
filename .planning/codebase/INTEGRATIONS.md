# External Integrations

**Analysis Date:** 2026-02-05

## APIs & External Services

**AI Providers:**
- **OpenAI** - GPT-3.5/4 models
  - SDK: `github.com/openai/openai-go/v2`
  - Auth: API key via `APIKey` config
  - Models: Default configured in `pkg/config/config.go`

- **Anthropic Claude** - Claude 3 models
  - SDK: `github.com/anthropics/anthropic-sdk-go`
  - Auth: API key via `APIKey` config
  - Models: Custom provider in `pkg/provider/anthropic/`

- **Google Gemini** - Google AI models
  - SDK: `github.com/google/generative-ai-go`
  - Auth: API key via `APIKey` config
  - Models: Google AI provider

- **Ollama** - Local AI models
  - SDK: `github.com/ollama/ollama`
  - Auth: Local API
  - Models: Configurable local models

- **DeepSeek** - DeepSeek AI models
  - Custom provider in `pkg/provider/deepseek/`
  - Auth: API key

- **OpenRouter** - Multi-provider AI routing
  - Custom provider in `pkg/provider/openrouter/`
  - Auth: API key

- **Phind** - Programming-focused AI
  - Custom provider in `pkg/provider/phind/`
  - Auth: API key

## Data Storage

**Databases:**
- **SQLite** - Local database
  - Connection: GORM driver
  - Client: `gorm.io/driver/sqlite`
  - Location: `~/.ai-commit-hub/ai-commit-hub.db` (platform-specific)

**File Storage:**
- Local filesystem - Git repositories and config files
- No cloud file storage integration

**Caching:**
- In-memory caching via Go structures
- No external caching service

## Authentication & Identity

**Auth Provider:**
- Custom API key authentication
  - Implementation: API keys stored in `config.yaml`
  - Encryption: No encryption (plaintext storage)
  - Multi-provider: Each AI provider has separate API keys

**User Management:**
- No user accounts system
- Local configuration only

## Monitoring & Observability

**Error Tracking:**
- No external error tracking service
- Local logging via `github.com/WQGroup/logger`

**Logs:**
- Framework: `github.com/WQGroup/logger`
- Format: JSON and text
- Rotation: Automatic cleanup configured
- Location: Platform-specific log directory

## CI/CD & Deployment

**Hosting:**
- Wails cross-platform builds
- No external hosting service

**CI Pipeline:**
- GitHub Actions in `.github/`
- Manual builds via `wails build`
- No automated deployment

## Environment Configuration

**Required env vars:**
- None required (configuration via YAML)
- Platform-specific config directories

**Secrets location:**
- Configuration file: `~/.ai-commit-hub/config.yaml`
- API keys stored in `providers` section
- File permissions: Standard user permissions

## Webhooks & Callbacks

**Incoming:**
- No webhook endpoints (desktop application)

**Outgoing:**
- No outgoing webhooks
- Local system tray notifications

## System Integration

**Operating System:**
- Windows system tray integration
- macOS/Linux system tray (via systray)
- File system operations for Git repositories

**Notifications:**
- **Pushover** - Push notification service
  - SDK: Custom implementation in `pkg/pushover/`
  - Auth: API token
  - Features: Update notifications, alerts
  - Configurable modes: enabled, pushover_only, windows_only, disabled

**Git Integration:**
- **Local Git repositories** - Direct git operations
  - SDK: `github.com/go-git/go-git/v5`
  - Operations: Commit, diff, status
  - No external Git hosting integration

---

*Integration audit: 2026-02-05*