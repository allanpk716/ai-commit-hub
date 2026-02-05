# Architecture

**Analysis Date:** 2026-02-05

## Pattern Overview

**Overall:** Hybrid desktop application with Wails (Go backend + Vue3 frontend)

**Key Characteristics:**
- Event-driven architecture with Wails Events for real-time updates
- Provider pattern for AI service abstraction
- Repository pattern for data persistence
- Clean separation between Go business logic and Vue UI
- System tray integration for background operations

## Layers

**Backend (Go) Layer:**
- Purpose: Core business logic, AI integration, Git operations, data persistence
- Location: `pkg/`
- Contains: Services, repositories, models, providers
- Depends on: External AI APIs, Git CLI, SQLite database
- Used by: Wails bridge, Vue frontend via generated bindings

**Frontend (Vue3) Layer:**
- Purpose: User interface, user interactions, state management
- Location: `frontend/src/`
- Contains: Vue components, stores, types, utilities
- Depends on: Wails runtime, generated Go bindings
- Used by: End users

**Data Layer:**
- Purpose: Data persistence and model definitions
- Location: `pkg/models/`, `pkg/repository/`
- Contains: GORM models, repositories, database schema
- Depends on: SQLite, GORM ORM
- Used by: Service layer

**AI Provider Layer:**
- Purpose: Abstract AI service integration
- Location: `pkg/ai/`, `pkg/provider/`
- Contains: Provider interfaces, implementations, registry
- Depends on: External AI APIs
- Used by: Service layer for AI operations

## Data Flow

**Commit Message Generation Flow:**

1. User selects project in Vue UI
2. Frontend calls Wails binding `GetProjectStatus()`
3. Go service queries Git repository status via `git` package
4. Frontend calls `GenerateCommit()` with project path and provider config
5. Go service:
   - Loads project config via `ConfigService`
   - Extracts Git diff via `git` package
   - Constructs prompt via `prompt` package
   - Calls AI provider via `provider` package
   - Streams response back to frontend via Wails Events
6. Frontend displays streaming message and final commit message
7. User confirms commit generation
8. Frontend calls Git operations via Wails bindings

**State Management:**
- Vue stores handle UI state (Pinia)
- Go services handle business state
- Wails Events bridge state updates
- SQLite persists configuration and project data

## Key Abstractions

**AI Provider Interface:**
- Purpose: Abstract different AI service providers
- Examples: `pkg/provider/openai/openai.go`, `pkg/provider/anthropic/anthropic.go`
- Pattern: Interface with concrete implementations

**Git Operations:**
- Purpose: Wrap Git CLI commands in Go
- Examples: `pkg/git/git.go`, `pkg/git/status.go`, `pkg/git/diff.go`
- Pattern: Service layer with command execution

**Configuration Service:**
- Purpose: Manage YAML-based configuration
- Examples: `pkg/service/config_service.go`, `pkg/config/`
- Pattern: Singleton with caching

**Project Management:**
- Purpose: Track multiple Git projects and their configs
- Examples: `pkg/service/project_config_service.go`
- Pattern: Repository pattern with GORM models

## Entry Points

**Wails Application Entry:**
- Location: `main.go`
- Triggers: Application startup, asset embedding
- Responsibilities: Wails app initialization, logger setup, asset serving

**Go Application Logic:**
- Location: `app.go`
- Triggers: Wails event handlers, system tray lifecycle
- Responsibilities: Request routing, service orchestration, UI event handling

**Frontend Application:**
- Location: `frontend/src/main.ts`, `frontend/src/App.vue`
- Triggers: Vue app initialization, component mounting
- Responsibilities: UI rendering, user interactions, state management

**System Tray Integration:**
- Location: `app.go` (systray integration)
- Triggers: Application minimization, tray icon clicks
- Responsibilities: Background operations, context menu, system notifications

## Error Handling

**Strategy:** Hierarchical error handling with graceful degradation

**Patterns:**
- Service layer errors with context (`pkg/service/error_service.go`)
- Frontend error display via toast components
- Wails Events for error streaming
- Log aggregation via `github.com/WQGroup/logger`

## Cross-Cutting Concerns

**Logging:**
- Framework: `github.com/WQGroup/logger`
- Patterns: Structured logging, JSON/text output, log rotation
- Location: `pkg/service/Logs/`

**Configuration:**
- Framework: YAML config with GORM backend
- Patterns: Environment-specific configs, project-scoped configs
- Location: `pkg/config/`, `pkg/models/`

**Git Integration:**
- Framework: Git CLI wrapper
- Patterns: Command execution with output parsing
- Location: `pkg/git/`

---

*Architecture analysis: 2026-02-05*