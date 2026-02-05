# Codebase Structure

**Analysis Date:** 2026-02-05

## Directory Layout

```
ai-commit-hub/
├── app.go                     # Wails application entry point
├── main.go                    # Wails application initialization
├── wails.json                 # Wails configuration
├── go.mod                     # Go module dependencies
├── go.sum                     # Go dependency checksums
├── frontend/                  # Vue3 frontend application
│   ├── src/
│   │   ├── App.vue           # Main Vue application component
│   │   ├── main.ts           # Vue application entry point
│   │   ├── components/       # Vue components
│   │   │   ├── CommitPanel.vue        # Commit generation UI
│   │   │   ├── ProjectList.vue        # Project management
│   │   │   ├── StagingArea.vue        # Git staging interface
│   │   │   ├── SettingsDialog.vue     # Configuration UI
│   │   │   └── ...                   # Other components
│   │   ├── stores/           # Pinia state management
│   │   │   ├── commitStore.ts         # Commit state
│   │   │   ├── projectStore.ts        # Project state
│   │   │   ├── statusCache.ts         # Status caching
│   │   │   └── ...                   # Other stores
│   │   ├── types/           # TypeScript type definitions
│   │   ├── utils/           # Utility functions
│   │   └── assets/          # Static assets
│   ├── package.json         # Node.js dependencies
│   ├── vite.config.ts      # Vite build configuration
│   └── tsconfig.json        # TypeScript configuration
├── pkg/                     # Go packages
│   ├── service/            # Business logic services
│   │   ├── commit_service.go          # Commit generation
│   │   ├── config_service.go          # Configuration management
│   │   ├── project_config_service.go   # Project configs
│   │   ├── error_service.go           # Error handling
│   │   ├── startup_service.go         # App startup
│   │   └── update_service.go          # Update management
│   ├── repository/         # Data access layer
│   │   ├── git_project_repository.go   # Project CRUD
│   │   └── Logs/                      # Log storage
│   ├── models/             # Data models
│   │   └── git_project.go             # Project model
│   ├── git/                # Git operations
│   │   ├── git.go                     # Git command wrapper
│   │   ├── status.go                  # Git status
│   │   ├── diff.go                    # Git diff
│   │   ├── commit.go                  # Git commit
│   │   ├── push.go                    # Git push
│   │   └── gitignore.go               # Git ignore
│   ├── provider/           # AI provider implementations
│   │   ├── registry/                  # Provider registry
│   │   ├── openai/                    # OpenAI implementation
│   │   ├── anthropic/                 # Anthropic implementation
│   │   ├── google/                    # Google implementation
│   │   ├── ollama/                    # Ollama implementation
│   │   └── ...                       # Other providers
│   ├── ai/                 # AI integration
│   │   └── client.go                  # AI client interface
│   ├── config/            # Configuration handling
│   │   ├── config.go                  # Config structure
│   │   └── provider.go                # Provider config
│   ├── prompt/            # Prompt construction
│   │   └── prompt.go                  # Prompt templates
│   ├── pushover/          # Pushover integration
│   │   ├── installer.go               # Extension installer
│   │   └── version.go                 # Version checking
│   ├── update/            # Update management
│   │   └── update.go                  # Update checking
│   └── version/           # Version information
│       └── version.go                 # Build version
└── docs/                  # Documentation
    ├── development/       # Development guides
    │   ├── wails-development-standards.md
    │   └── logging-standards.md
    ├── lessons-learned/   # Implementation guides
    └── plans/            # Project planning
```

## Directory Purposes

**Backend (pkg/):**
- Purpose: Go backend business logic and services
- Contains: Services, repositories, models, providers
- Key files: `pkg/service/commit_service.go`, `pkg/git/git.go`

**Frontend (frontend/):**
- Purpose: Vue3 user interface
- Contains: Components, stores, types, utilities
- Key files: `frontend/src/App.vue`, `frontend/src/stores/commitStore.ts`

**Configuration & Models:**
- Purpose: Data persistence and configuration
- Contains: GORM models, YAML config handling
- Key files: `pkg/models/git_project.go`, `pkg/config/config.go`

**AI Providers:**
- Purpose: Abstract AI service integrations
- Contains: Provider implementations and registry
- Key files: `pkg/provider/registry/registry.go`, `pkg/provider/openai/openai.go`

**Documentation:**
- Purpose: Project documentation and guides
- Contains: Development standards, implementation lessons
- Key files: `docs/development/wails-development-standards.md`

## Key File Locations

**Entry Points:**
- `main.go`: Wails application initialization
- `app.go`: Application event handlers and logic
- `frontend/src/main.ts`: Vue application entry
- `frontend/src/App.vue`: Main Vue component

**Configuration:**
- `wails.json`: Wails build configuration
- `frontend/vite.config.ts`: Frontend build configuration
- `frontend/tsconfig.json`: TypeScript configuration
- `go.mod`: Go dependencies

**Core Logic:**
- `pkg/service/commit_service.go`: Commit message generation
- `pkg/git/git.go`: Git operations wrapper
- `pkg/provider/registry/registry.go`: AI provider management
- `pkg/config/config.go`: Configuration management

**Testing:**
- `pkg/service/commit_service_test.go`: Service tests
- `pkg/git/commit_test.go`: Git operation tests
- `tests/`: Integration tests

## Naming Conventions

**Files:**
- Go: `snake_case` (e.g., `commit_service.go`)
- Vue: `PascalCase` (e.g., `CommitPanel.vue`)
- Tests: `*_test.go` (e.g., `commit_service_test.go`)

**Packages:**
- Go: `snake_case` (e.g., `pkg/service/`)
- Vue: `kebab-case` for filenames, PascalCase for imports

**Variables:**
- Go: `camelCase` (e.g., `projectPath`)
- Vue/TypeScript: `camelCase` (e.g., `projectPath`)

**Functions:**
- Go: `camelCase` with clear intent (e.g., `GenerateCommit`)
- Vue/TypeScript: `camelCase` (e.g., `generateCommit`)

## Where to Add New Code

**New AI Provider:**
- Implementation: `pkg/provider/[provider-name]/`
- Registration: `pkg/provider/registry/registry.go`
- Configuration: Update provider models

**New Git Feature:**
- Implementation: `pkg/git/[feature].go`
- Service integration: `pkg/service/commit_service.go`
- Tests: `pkg/git/[feature]_test.go`

**New UI Component:**
- Implementation: `frontend/src/components/[ComponentName].vue`
- Store integration: `frontend/src/stores/[store].ts`
- Types: `frontend/src/types/[type].ts`

**New Configuration Option:**
- Model: `pkg/models/[model].go`
- Service: `pkg/service/config_service.go`
- Frontend: `frontend/src/stores/[store].ts`

**New Business Logic:**
- Service: `pkg/service/[logic]_service.go`
- Repository: `pkg/repository/[model]_repository.go`
- Model: `pkg/models/[model].go`

## Special Directories

**.wails/**:
- Purpose: Wails build artifacts and temporary files
- Generated: Yes
- Committed: No

**frontend/dist/**:
- Purpose: Frontend build output
- Generated: Yes
- Committed: No

**frontend/wailsjs/**:
- Purpose: Generated Go bindings for Vue
- Generated: Yes
- Committed: No

**tmp/**:
- Purpose: Temporary files for development
- Generated: Yes
- Committed: No

---

*Structure analysis: 2026-02-05*