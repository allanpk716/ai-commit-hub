# Technology Stack

**Analysis Date:** 2026-02-05

## Languages

**Primary:**
- Go 1.24.11 - Backend implementation, Wails application, AI provider integrations
- TypeScript 5.9.3 - Frontend development, type safety, component logic

**Secondary:**
- JavaScript - Frontend runtime, Vite build system
- Vue.js template syntax - UI templating

## Runtime

**Environment:**
- Wails v2.11.0 - Cross-platform desktop application framework
- Node.js 20+ - Frontend development and build environment

**Package Manager:**
- Go Module - Go dependency management (go.mod)
- npm - Frontend dependency management (package.json)

## Frameworks

**Core:**
- Wails v2.11.0 - Desktop application framework (Go + Web UI)
- Vue 3.5.24 - Frontend UI framework with Composition API
- Pinia 3.0.4 - State management for Vue 3

**Testing:**
- Vitest 4.0.18 - Frontend testing framework
- @testing-library/vue 8.1.0 - Vue component testing utilities
- Go testify 1.11.1 - Go testing framework

**Build/Dev:**
- Vite 7.2.4 - Frontend build tool and development server
- vue-tsc 3.1.4 - TypeScript type checking for Vue
- TypeScript 5.9.3 - Type checking and compilation

## Key Dependencies

**Critical:**
- Wails v2.11.0 - Core desktop application framework
- Go-git v5.16.4 - Git operations and repository management
- GORM v1.31.1 + SQLite v1.6.0 - Database ORM and driver
- WQGroup logger v0.0.16 - Structured logging with rotation

**Infrastructure:**
- gopkg.in/yaml.v3 v3.0.1 - YAML configuration parsing
- golang.org/x/sys v0.39.0 - Go system interface
- golang.org/x/oauth2 v0.30.0 - OAuth2 authentication

## AI Provider Ecosystem

**Multiple AI Provider Support:**
- OpenAI v2.7.1 - OpenAI API integration (GPT models)
- Anthropic v1.19.0 - Claude API integration
- Google v0.20.1 - Gemini AI integration
- Ollama v0.14.3 - Local Ollama models integration
- DeepSeek - DeepSeek API integration
- Phind - Phind AI integration
- Custom provider registry for dynamic provider loading

## Configuration

**Environment:**
- User home directory: `~/.ai-commit-hub/` (Windows: `C:\Users\<username>\.ai-commit-hub\`)
- Configuration file: `config.yaml`
- Database: `ai-commit-hub.db` (SQLite)
- Custom prompts: `prompts/` directory

**Build:**
- Wails configuration: `wails.json`
- Build flags: Version, commit SHA, build time injection
- Asset embedding: Frontend dist, icons

## Platform Requirements

**Development:**
- Go 1.24+ (toolchain go1.24.11)
- Node.js 20+
- npm
- Wails CLI

**Production:**
- Windows: Compiled Go binary + embedded web assets
- Cross-platform build support through Wails
- System tray functionality (Windows/Linux)

## Logging

**Framework:**
- WQGroup logger v0.0.16 - Structured logging with multiple formatters
- File-based logging with rotation (100MB max, 30 days retention)
- Log directory: `~/.ai-commit-hub/logs/`

**Features:**
- JSON, text, and structured log formats
- Automatic log rotation and cleanup
- Thread-safe concurrent logging

---

*Stack analysis: 2026-02-05*