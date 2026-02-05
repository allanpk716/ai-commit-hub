# Technology Stack

**Analysis Date:** 2026-02-05

## Languages

**Primary:**
- Go 1.24.1 - Backend API and core functionality
- TypeScript 5.9.3 - Frontend development

**Secondary:**
- Vue 3.5.24 - UI framework
- JavaScript - Frontend logic

## Runtime

**Backend:**
- Go runtime 1.24.11
- Wails v2.11.0 - Desktop application framework

**Frontend:**
- Vite 7.2.4 - Build tool and dev server
- Node.js - Runtime environment

**Package Manager:**
- Go mod - Go dependency management
- npm - Frontend dependency management
- Package-lock.json - Frontend dependency locking

## Frameworks

**Backend Core:**
- Wails v2.11.0 - Desktop application framework
- GORM v1.31.1 - ORM for database operations
- SQLite - Local database storage

**Frontend:**
- Vue 3.5.24 - Progressive JavaScript framework
- Pinia 3.0.4 - State management
- Vite 7.2.4 - Build tool and dev server

**Testing:**
- Vitest 4.0.18 - Frontend testing framework
- Go testing - Backend testing
- @testing-library/vue 8.1.0 - Vue component testing

## Key Dependencies

**Critical:**
- github.com/WQGroup/logger v0.0.16 - Logging framework
- github.com/wailsapp/wails/v2 v2.11.0 - Desktop application framework
- gorm.io/gorm v1.31.1 - Database ORM
- gorm.io/driver/sqlite v1.6.0 - SQLite driver
- github.com/go-git/go-git/v5 v5.16.4 - Git operations

**AI Providers:**
- github.com/anthropics/anthropic-sdk-go v1.19.0 - Anthropic Claude
- github.com/openai/openai-go/v2 v2.7.1 - OpenAI API
- github.com/google/generative-ai-go v0.20.1 - Google Gemini
- github.com/ollama/ollama v0.14.3 - Ollama local AI
- github.com/allanpk716/ai-commit-hub/pkg/provider/* - Custom provider implementations

**Infrastructure:**
- github.com/getlantern/systray v1.2.2 - System tray functionality
- github.com/sergi/go-diff v1.4.0 - Text diff computation
- github.com/go-playground/validator/v10 v10.30.1 - Data validation
- gopkg.in/yaml.v3 v3.0.1 - YAML configuration parsing

## Configuration

**Environment:**
- Windows primary development platform
- Config directory: `~/.config/ai-commit-hub/` (Linux/macOS) or `C:\Users\<username>\.ai-commit-hub\` (Windows)
- Configuration file: `config.yaml`

**Build:**
- Wails build configuration in `vite.config.ts`
- TypeScript configuration in `tsconfig.app.json`
- Vue 3 + Vite setup in `frontend/`

## Platform Requirements

**Development:**
- Go 1.24+
- Node.js 18+
- npm or yarn

**Production:**
- Windows, macOS, Linux (cross-platform via Wails)
- System tray support
- SQLite database

---

*Stack analysis: 2026-02-05*