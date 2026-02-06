# Project Research Summary

**Project:** AI Commit Hub
**Domain:** Wails Windows Desktop Application (Single-Instance, Auto-Update, System Tray, CI/CD)
**Researched:** 2026-02-06
**Confidence:** HIGH

## Executive Summary

AI Commit Hub is a lightweight Windows desktop application built with Wails v2 (Go + Vue3) that generates AI-powered commit messages for Git repositories. Research indicates that expert-built Windows desktop applications in this domain prioritize three core infrastructure elements: single-instance enforcement, system tray integration with proper lifecycle management, and automatic updates. The recommended approach uses Wails' built-in `SingleInstanceLock` for process isolation, an external updater process pattern for clean file replacement during updates, and event-driven architecture (Wails Events) for non-blocking update checks.

Key risks identified include single-instance deadlock from improper mutex handling, auto-update failures due to Windows file locking, and system tray race conditions during shutdown. Mitigation strategies include using Wails' native single-instance abstraction (not manual Windows mutex), implementing a separate `updater.exe` process that waits for the main application to exit before file replacement, and using `sync.Once` with `atomic.Bool` flags to coordinate systray cleanup. The project already has some infrastructure in place (system tray with right-click menu, `quitting` flag pattern), but needs upgrades for double-click support and single-instance locking.

## Key Findings

### Recommended Stack

The research recommends Wails v2.11.0 with built-in `SingleInstanceLock` for process management. For system tray, upgrade from `getlantern/systray` to `lutischan-ferenc/systray` v1.3.0+ to enable double-click support (`SetOnDClick`). Auto-update should use `creativeprojects/go-selfupdate` v1.5.2 (actively maintained Dec 2025) instead of unmaintained alternatives like `rhysd/go-github-selfupdate` (last updated 2017) or `marcus-crane/wails-autoupdater` (unmaintained since Jan 2023). CI/CD should leverage `dAppServer/wails-build-action` for cross-platform builds with `softprops/action-gh-release` for artifact publishing.

**Core technologies:**
- **Wails v2.11.0** - Desktop framework with built-in single-instance lock, mature ecosystem, actively maintained
- **Go 1.21+** - Backend language with Windows API access via `golang.org/x/sys/windows`
- **Vue 3** - Frontend framework (already in use, proven Wails integration)
- **github.com/creativeprojects/go-selfupdate v1.5.2** - Auto-update library with GitHub Releases integration, active maintenance
- **github.com/lutischan-ferenc/systray v1.3.0** - System tray with double-click support (upgrade required)
- **GitHub Actions** - CI/CD platform with official Wails build action

### Expected Features

Research identified three tiers of features for Windows desktop applications: table stakes (users expect these), differentiators (competitive advantage), and anti-features (commonly requested but problematic).

**Must have (table stakes):**
- **Single-Instance Enforcement** - Prevents duplicate instances, uses Mutex + IPC for argument forwarding, activates existing window
- **Minimize to System Tray** - Standard Windows behavior since Windows 95, intercept close event to hide instead of quit
- **Tray Icon Right-Click Menu** - Universal Windows UI convention with "Show" and "Exit" options
- **Basic Auto-Update Checking** - Check for updates on startup, show dialog if new version available
- **Settings/Data Persistence** - SQLite + GORM (already implemented), store in user home directory
- **Window State Memory** - Remember window position and size between sessions
- **Clean Application Exit** - Proper cleanup with `quitting` flag pattern (already implemented)

**Should have (competitive):**
- **Double-Click Tray Icon to Show Window** - Faster than right-click menu, requires `lutischan-ferenc/systray` upgrade
- **Toast Notification for Updates** - Non-intrusive Windows Action Center notifications instead of modal dialog
- **Commit Message Style Templates** - Conventional Commits, emoji-based styles, differentiator vs generic tools
- **Project-Specific AI Provider** - Different projects can use different AI models (already implemented)
- **One-Click Installer Update** - Auto-download and prompt to install, still requires user approval

**Defer (v2+):**
- **Silent Auto-Update** - Background download + install, requires Squirrel framework and code signing certificate
- **Keyboard Shortcuts for Power Users** - Global hotkey registration, requires IPC for single-instance coordination
- **Commit History Analytics** - Nice-to-have analytics, not core value proposition
- **Full Git GUI** - Anti-feature, scope creep, use TortoiseGit integration instead

### Architecture Approach

The recommended architecture follows Wails event-driven patterns with clear separation between UI (Vue3 frontend), application logic (Go backend), and external processes (updater.exe). Major components include the App lifecycle manager (`app.go`), UpdateService for GitHub release checking, external updater process for clean file replacement, and systray integration with proper goroutine management. The architecture uses Wails Events for async communication between Go and TypeScript, enabling non-blocking update checks and real-time progress streaming.

**Major components:**
1. **App (app.go)** - Wails application lifecycle, window state, event routing, systray integration
2. **Single Instance Lock** - Wails built-in `SingleInstanceLock` with `OnSecondInstanceLaunch` callback for window activation
3. **UpdateService** - GitHub Releases API integration, version comparison, release metadata caching
4. **External Updater Process** - Separate `cmd/updater/main.go` program that waits for main app PID to exit, extracts update zip, relaunches application
5. **Systray Integration** - System tray with click/double-click handlers, menu items, proper cleanup coordination
6. **Vue3 Frontend** - Pinia stores for update state, Wails event listeners for async notifications

### Critical Pitfalls

Research identified five critical pitfalls that commonly break Windows desktop applications in this domain, along with prevention strategies.

1. **Single-Instance Deadlock with Wails** - Using Go `sync.Mutex` across process boundaries (doesn't work) or file-based locks that become stale after crashes. **Prevention:** Use Wails built-in `SingleInstanceLock` option with UniqueId and `OnSecondInstanceLaunch` callback, not manual Windows mutex implementation.

2. **Auto-Update Breaking Running Application** - Windows locks executable files while running, causing "access denied" errors when trying to replace in-use files. **Prevention:** Use external updater process pattern (`cmd/updater/main.go`) that waits for main app PID to exit before file replacement, implement backup mechanism (`.bak` files) for rollback.

3. **System Tray Event Handling Race Conditions** - Multiple goroutines accessing tray state without synchronization, calling `systray.Quit()` from within event handlers without coordination, timing issues between systray and Wails initialization. **Prevention:** Use `sync.RWMutex` for window state, `atomic.Bool` for `quitting` flag, `sync.Once` for one-time cleanup, add 300ms delay between systray and Wails initialization.

4. **GitHub Actions Build Failures** - Go version incompatibilities (e.g., Go 1.25.x with Wails 2.10.2), corrupted `wailsbindings` executables, missing WebView2 runtime in CI. **Prevention:** Pin Go version to 1.21+ in workflows, use `wails build -clean` to clear cache, install WebView2 runtime explicitly on Windows runners.

5. **Windows-Specific Issues (UAC, File Locking, Permissions)** - Writing to Program Files instead of user profile, not checking WebView2 availability, unexpected UAC prompts after Windows updates. **Prevention:** Write to `%USERPROFILE%\.ai-commit-hub`, embed Windows manifest requesting `asInvoker`, check WebView2 at startup with download link fallback.

## Implications for Roadmap

Based on research synthesis, suggested phase structure prioritizes infrastructure stability (single-instance, systray) before update functionality, with CI/CD automation enabling rapid iteration.

### Phase 1: Foundation & Stability
**Rationale:** Single-instance enforcement and systray fixes are interdependent with window management. Establishing these patterns first prevents architectural debt. The project already has most systray code; this phase focuses on upgrades and fixes.

**Delivers:**
- Single-instance lock with second instance forwarding
- System tray double-click support (library upgrade)
- Window state memory (position/size persistence)
- Clean application exit validation

**Addresses:**
- Single-Instance Enforcement (FEATURES.md table stakes)
- Double-Click Tray Icon (FEATURES.md differentiator)
- Window State Memory (FEATURES.md table stakes)
- Clean Application Exit (FEATURES.md table stakes)

**Avoids:**
- Single-Instance Deadlock (PITFALLS.md #1)
- System Tray Race Conditions (PITFALLS.md #3)
- Systray Shutdown Deadlock (ARCHITECTURE.md anti-pattern #3, #4)

### Phase 2: Update Infrastructure
**Rationale:** Update functionality requires stable foundation (single-instance prevents update conflicts). This phase builds backend services before adding UI. External updater process must be implemented before frontend integration.

**Delivers:**
- UpdateService with GitHub Releases API integration
- Downloader with progress tracking
- External updater process (cmd/updater/main.go)
- Installer launching logic with PID waiting
- Basic update UI dialog

**Addresses:**
- Basic Auto-Update Checking (FEATURES.md table stakes)
- One-Click Installer Update (FEATURES.md differentiator)

**Uses:**
- creativeprojects/go-selfupdate v1.5.2 (STACK.md)
- External Updater Process Pattern (ARCHITECTURE.md pattern #2)

**Implements:**
- UpdateService, Downloader, Installer components (ARCHITECTURE.md)
- Wails Event-Driven Update Flow (ARCHITECTURE.md pattern #3)

**Avoids:**
- Auto-Update Breaking Running Application (PITFALLS.md #2)
- In-Process Update Extraction (ARCHITECTURE.md anti-pattern #2)
- Blocking Startup with Update Check (ARCHITECTURE.md anti-pattern #1)

### Phase 3: CI/CD & Polish
**Rationale:** CI/CD automation enables frequent releases with confidence. Build workflow validates code quality; release workflow automates artifact publishing. This phase also addresses remaining technical debt (compilation errors, test fixes).

**Delivers:**
- GitHub Actions build workflow (test + artifact)
- GitHub Actions release workflow (tag-triggered)
- Code signing configuration (optional but recommended)
- Toast notification for updates
- Fix compilation errors in app.go and tests
- Window state persistence implementation

**Addresses:**
- Toast Notification for Updates (FEATURES.md differentiator)
- CI/CD修复 (PROJECT.md active requirements)
- 编译错误修复 (PROJECT.md active requirements)

**Uses:**
- dAppServer/wails-build-action (STACK.md)
- softprops/action-gh-release (STACK.md)

**Avoids:**
- GitHub Actions Build Failures (PITFALLS.md #4)

### Phase 4: Validation & Launch Preparation
**Rationale:** End-to-end testing validates the complete update flow. This phase ensures all infrastructure works together before public release. Includes documentation and user guides.

**Delivers:**
- End-to-end update testing (manual)
- Update rollback testing
- Single-instance cross-session testing
- User documentation (update process, troubleshooting)
- Release notes generation from commits

**Addresses:**
- Full integration testing of all features
- Documentation for end users

**Avoids:**
- All "Looks Done But Isn't" issues (PITFALLS.md checklist)

### Phase Ordering Rationale

The phase order follows three principles: dependency management, risk mitigation, and incremental validation. Phase 1 establishes core application lifecycle patterns (single-instance, systray) that all other features depend on. Phase 2 builds update infrastructure on stable foundation, preventing update conflicts from multi-instance scenarios. Phase 3 automates release pipeline, enabling rapid iteration without manual build steps. Phase 4 validates complete flow before exposing to users.

This grouping avoids the most critical pitfalls: single-instance deadlock (addressed in Phase 1), update breaking apps (addressed in Phase 2), and CI failures (addressed in Phase 3). The architecture supports phased delivery—each phase produces working software that can be tested independently.

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 2 (Update Infrastructure):** External updater process error handling needs validation—test what happens when updater crashes mid-extraction, implement robust recovery logic.
- **Phase 3 (CI/CD):** Code signing certificate selection and costs vary ($100-500/year), research which certificate authority (DigiCert vs Sectigo) offers best value for Windows desktop apps.
- **Phase 4 (Validation):** GitHub Releases API rate limiting (60 requests/hour unauthenticated) may affect users at scale, research caching strategy and authenticated API quotas.

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation):** Wails single-instance and systray patterns are well-documented in official guides, project already has working implementations to reference.
- **Phase 3 (CI/CD):** Wails build action and release workflows have community examples, straightforward implementation following official docs.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Based on official Wails documentation, actively maintained libraries (go-selfupdate v1.5.2 updated Dec 2025), and project's existing working implementations |
| Features | HIGH | Windows UX conventions verified with Microsoft official documentation, feature priorities based on competitor analysis (SourceTree, GitKraken) and project requirements |
| Architecture | HIGH | Patterns sourced from Wails official docs (single-instance, event-driven flow), project's existing codebase analysis (app.go, systray implementation), and established Go desktop app patterns |
| Pitfalls | MEDIUM-HIGH | 5 critical pitfalls verified with official docs (Wails troubleshooting, Microsoft UAC), 3 medium pitfalls from community issues (GitHub discussions), but Windows-specific issues (August 2025 UAC bug) have limited real-world validation data |

**Overall confidence:** HIGH

All critical recommendations (Wails built-in single-instance, external updater process, lutischan-ferenc systray upgrade, go-selfupdate library) are supported by official documentation or actively maintained codebases. The project's existing implementations provide validation that the recommended approaches work in practice. Medium confidence on some Windows-specific issues (recent UAC bug) reflects limited real-world testing data, but mitigation strategies are well-defined.

### Gaps to Address

- **External Updater Error Recovery:** Research doesn't cover all failure modes for external updater process (e.g., updater crashes after main app exits but before extraction completes). **Handle during planning:** Add recovery logic to main app startup—check for partial update state on launch, prompt user to retry or download full installer.

- **Code Signing Certificate Selection:** Research identifies need for certificate but doesn't recommend specific certificate authority. **Handle during planning:** Compare DigiCert, Sectigo, and GlobalSign for Windows code signing certificates—evaluate cost, validation requirements, and Smartscreen reputation impact.

- **GitHub API Rate Limiting at Scale:** Current approach (unauthenticated GitHub Releases API) may hit rate limits with >1000 users checking updates daily. **Handle during planning:** Implement ETag/If-Modified-Since caching to reduce API calls by ~90%. Monitor usage and add authentication if rate limits become issue.

- **Double-Click High DPI Testing:** Research doesn't validate double-click behavior on high DPI displays (150%, 200% scaling). **Handle during planning:** Test double-click on Windows 11 with different DPI settings, add explicit testing scenario in Phase 1 validation.

## Sources

### Primary (HIGH confidence)
- **Wails Official Documentation** - Single instance lock guide (v2.11.0), system tray API, event-driven architecture patterns
- **Microsoft Official Documentation** - Windows UAC guidelines, file locking behavior, Notification Area UX conventions
- **Project Documentation** - `docs/lessons-learned/windows-tray-icon-implementation-guide.md`, `docs/fixes/tray-icon-doubleclick-fix.md`, `docs/fixes/systray-exit-fix.md` (verified working implementations)
- **Active Libraries** - `creativeprojects/go-selfupdate` v1.5.2 (Dec 2025), `lutischan-ferenc/systray` v1.3.0, `dAppServer/wails-build-action`

### Secondary (MEDIUM confidence)
- **Community Resources** - Wails GitHub discussions (#2441 multi-instance issues, #1178 self-updating support), Stack Overflow patterns (Go Windows single-instance, systray click handling)
- **Competitor Analysis** - SourceTree, GitKraken feature sets and UX patterns (industry standards for Git GUIs)
- **Windows Internals** - Thread management, mutex behavior, race condition detection (Google Project Zero, technical blogs)

### Tertiary (LOW confidence)
- **Unmaintained Libraries** - `rhysd/go-github-selfupdate` (last updated 2017), `marcus-crane/wails-autoupdater` (Jan 2023) - cited as anti-patterns to avoid
- **Community Tutorials** - CSDN blog posts (Dec 2025), Medium articles on Wails CI/CD - useful for implementation details but less reliable than official docs

---
*Research completed: 2026-02-06*
*Ready for roadmap: yes*
