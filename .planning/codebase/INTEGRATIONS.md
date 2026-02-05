# External Integrations

**Analysis Date:** 2026-02-05

## APIs & External Services

**AI Provider APIs:**
- **OpenAI** - GPT-3.5/GPT-4 models for commit message generation
  - SDK: github.com/openai/openai-go/v2 v2.7.1
  - Auth: API key via config.yaml
  - Endpoint: https://api.openai.com/v1

- **Anthropic Claude** - Claude 3 models for commit analysis
  - SDK: github.com/anthropics/anthropic-sdk-go v1.19.0
  - Auth: API key via config.yaml
  - Endpoint: https://api.anthropic.com

- **Google AI** - Gemini models for commit generation
  - SDK: github.com/google/generative-ai-go v0.20.1
  - Auth: API key via config.yaml

- **DeepSeek** - DeepSeek Chat for commit analysis
  - SDK: Custom implementation
  - Auth: API key via config.yaml
  - Endpoint: https://api.deepseek.com

- **Ollama** - Local AI model hosting
  - SDK: github.com/ollama/ollama v0.14.3
  - Auth: Local server (no auth required)
  - Endpoint: http://localhost:11434

- **Phind** - AI code assistant
  - SDK: Custom implementation
  - Auth: API key via config.yaml

- **Custom Provider Registry** - Dynamic provider loading and management
  - Location: `pkg/provider/registry/`
  - Supports runtime provider switching

## Data Storage

**Databases:**
- **SQLite** - Local database for projects and commit history
  - ORM: GORM v1.31.1
  - Driver: gorm.io/driver/sqlite v1.6.0
  - File: `~/.ai-commit-hub/ai-commit-hub.db`
  - Models: GitProject, CommitHistory

**File Storage:**
- **Local filesystem** - Git repositories, custom prompts, configuration
- **Embedded assets** - Frontend build files, application icons
- **Temporary files** - `tmp/` directory for testing and temporary data

**Caching:**
- **In-memory cache** - Project status caching (StatusCache)
- **Cache TTL** - 30 seconds with background refresh
- **Optimistic updates** - Immediate UI updates with async verification

## Authentication & Identity

**Git Authentication:**
- **SSH Keys** - Standard Git SSH key authentication
- **HTTP Basic Auth** - Username/password for remote repositories
- **Credential Helpers** - System credential manager integration
- **Token-based** - Personal access tokens for GitHub/GitLab

**AI Provider Authentication:**
- **API Keys** - Direct API key authentication for all AI providers
- **Environment Variables** - Optional alternative to config.yaml
- **Secure Storage** - API keys stored in encrypted config.yaml

## Pushover Integration

**Pushover Hook System:**
- **Custom Python Hook** - Pushover notification webhook
- **Installation Management** - Automatic download and installation
- **Version Management** - Update checking and reinstallation
- **Notification Modes** - Full, Pushover-only, Windows-only, Disabled

**Features:**
- **Environment Variables** - Pushover API token and user key
- **Local Installation** - Per-project hook installation
- **Configuration Persistence** - `.no-pushover`, `.no-windows` files
- **Status Tracking** - Real-time hook status and version info

## Git Integration

**Repository Management:**
- **Go-git v5.16.4** - Git operations implementation
- **Status Monitoring** - Real-time git status and staging area
- **Commit Operations** - Local commit creation with AI messages
- **Push Operations** - Remote repository synchronization
- **Branch Management** - Current branch detection and switching

**Features:**
- **Diff Generation** - Unified diff for AI analysis
- **Staging Area** - Interactive file staging
- **Untracked Files** - New file detection and exclusion
- **Gitignore Support** - Standard gitignore processing

## System Integration

**Desktop Integration:**
- **System Tray** - Windows/Linux tray functionality
- **Window Management** - Minimize to tray, restore from tray
- **File Associations** - Optional Git repository associations
- **Auto-start** - Optional system startup integration

**Build Integration:**
- **Wails Build** - Cross-platform desktop builds
- **Code Signing** - Windows certificate signing support
- **Auto-update** - Built-in update mechanism
- **Icon Management** - Multi-resolution system tray icons

## Monitoring & Observability

**Logging:**
- **Structured Logging** - JSON and text log formats
- **Log Rotation** - Automatic file rotation and cleanup
- **Log Levels** - Debug, Info, Warn, Error levels
- **Performance Tracking** - Response times and error rates

**Error Tracking:**
- **Custom Error Service** - Centralized error handling
- **Error Display** - Toast notifications for user feedback
- **Error Recovery** - Graceful degradation and retry logic

## CI/CD & Deployment

**GitHub Actions:**
- **Release Workflow** - Automated builds and releases
- **Node.js Version** - v20 for frontend builds
- **Windows Builds** - Primary platform support
- **Version Management** - Automatic version injection

**Deployment:**
- **GitHub Releases** - Primary distribution channel
- **Binary Distribution** - Pre-built executables for Windows
- **Homebrew** - Optional macOS installation via Homebrew
- **Scoop** - Windows package manager support

## Environment Configuration

**Required env vars:**
- `PUSHOVER_API_TOKEN` - Pushover API token (for notifications)
- `PUSHOVER_USER_KEY` - Pushover user key (for notifications)

**Secrets location:**
- Configuration file: `~/.ai-commit-hub/config.yaml`
- Environment variables (alternative to config file)
- System credential manager (for Git authentication)

## Webhooks & Callbacks

**Incoming:**
- **Pushover Webhook** - HTTP endpoint for Git commit notifications
- **Status Webhook** - Repository status change notifications
- **Custom Hooks** - Extensible webhook system architecture

**Outgoing:**
- **Git Push** - Remote repository synchronization
- **Commit History** - Local database persistence
- **Status Updates** - Real-time UI updates via Wails Events

---

*Integration audit: 2026-02-05*