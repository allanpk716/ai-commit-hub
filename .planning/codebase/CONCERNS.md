# Codebase Concerns

**Analysis Date:** 2025-02-05

## Tech Debt

**Large Monolithic App Structure:**
- Issue: Main app.go file is 1942 lines, violating single responsibility principle
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go`
- Impact: Hard to maintain, test, and understand
- Fix approach: Extract logical modules into separate services (e.g., project management, AI generation, UI management)

**Vue Component Bloat:**
- Issue: Components like CommitPanel.vue are too large (1895 lines)
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\components\CommitPanel.vue`
- Impact: Difficult to maintain, violates component design principles
- Fix approach: Split into smaller, focused components (e.g., commit generation, file staging, settings panels)

**Mixed Responsibilities in Store:**
- Issue: commitStore.ts handles multiple concerns (project status, config, AI generation)
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\frontend\src\stores\commitStore.ts`
- Impact: Hard to test and maintain
- Fix approach: Split into projectStore, aiConfigStore, and generationStore

**Multiple return patterns:**
- Issue: Functions inconsistently return null, empty structs, or errors
- Files: Multiple files including `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\git\status.go`, `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\git\staging.go`
- Impact: Inconsistent error handling and null checks throughout the codebase
- Fix approach: Establish consistent return patterns with proper error types

## Known Bugs

**Wails Bindings Generation Issues:**
- Issue: Windows binding generation failures causing "not a valid Win32 application" errors
- Files: Referenced in multiple docs
- Trigger: Running `wails dev` on Windows
- Workaround: Delete wbindings files or use existing bindings
- Priority: High - blocks development on Windows

**Systray State Management:**
- Issue: Complex atomic flags and mutexes for systray lifecycle
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` (lines 69-70, 294-305)
- Trigger: Multiple systray operations
- Impact: Potential race conditions in window visibility management
- Priority: Medium - affects user experience

## Security Considerations

**Command Injection Risk:**
- Area: Shell commands for terminal/file explorer
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` (OpenInTerminal, OpenInFileExplorer)
- Current mitigation: Platform-specific commands with parameterized paths
- Recommendations: Validate all path inputs, use Go filepath operations, implement command sandboxing

**API Key Storage:**
- Area: AI provider credentials
- Risk: Potential exposure in configuration files
- Current mitigation: Local file storage with no encryption
- Recommendations: Implement secure credential storage or OS keyring integration

## Performance Bottlenecks

**Synchronous Hook Initialization:**
- Issue: Blocking initialization of all project hooks during startup
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` (line 173)
- Problem: Delays app startup with many projects
- Improvement path: Async initialization with lazy loading

**Git Operations Frequency:**
- Issue: Multiple git status calls per project
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` (GetSingleProjectStatus)
- Problem: Redundant git invocations
- Improvement path: Implement status cache with proper invalidation

## Fragile Areas

**Cross-Platform UI Code:**
- Files: `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` (OpenInTerminal, OpenInFileExplorer)
- Why fragile: Platform-specific logic mixed in main app
- Safe modification: Extract to dedicated platform services
- Test coverage: Partial integration tests exist

**Error Handling Patterns:**
- Files: Throughout pkg/ directory, especially git operations
- Why fragile: Mixed error handling strategies
- Safe modification: Centralize error definitions and handling
- Test coverage: Limited unit tests for error paths

## Dependencies at Risk

**External Git Commands:**
- Package: Git command-line tools
- Risk: Dependencies on external git installation and version
- Impact: Cross-platform compatibility issues
- Migration plan: Implement gitlib-go for embedded git functionality

**Wails v2 Framework:**
- Risk: Framework version dependencies and compatibility
- Impact: Potential breaking changes in future updates
- Migration plan: Monitor framework roadmap, consider gradual migration to v3 if available

## Missing Critical Features

**Project Configuration Validation:**
- Problem: Limited validation of AI provider configurations
- Blocks: Users cannot easily identify misconfigured providers
- Priority: Medium - affects user onboarding

**Batch Operations:**
- Problem: No batch staging or unstaging of files
- Blocks: Efficiency with many modified files
- Priority: Medium - affects productivity

## Test Coverage Gaps

**Error Path Testing:**
- What's not tested: Network failures, Git command errors, Configuration validation
- Files: Most pkg/ services have minimal error test coverage
- Risk: Runtime errors could crash the application
- Priority: High - stability concern

**Vue Component Testing:**
- What's not tested: Component interactions, edge cases, error states
- Files: Frontend components have limited test files
- Risk: UI bugs may go unnoticed
- Priority: Medium - user experience concern

**Integration Testing:**
- What's not tested: End-to-end workflows with real repositories
- Files: Limited integration test coverage
- Risk: Component interactions may fail in production
- Priority: High - reliability concern

---

*Concerns audit: 2025-02-05*