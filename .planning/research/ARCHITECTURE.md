# Architecture Research

**Domain:** Wails v2 Desktop Application Enhancements
**Researched:** 2026-02-06
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         User Interface Layer                            │
│  ┌───────────────┐  ┌──────────────┐  ┌───────────────┐               │
│  │ Main Window   │  │ Update UI    │  │ System Tray   │               │
│  │ (Vue3 Frontend)│  │ Components   │  │ Integration    │               │
│  └───────┬───────┘  └──────┬───────┘  └───────┬───────┘               │
│          │                  │                  │                        │
├──────────┼──────────────────┼──────────────────┼────────────────────────┤
│          │         Wails Events & Bindings      │                        │
│          ▼                  ▼                  ▼                        │
├─────────────────────────────────────────────────────────────────────────┤
│                      Application Logic Layer (Go)                       │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                        App (app.go)                            │    │
│  │  - Lifecycle management (startup/shutdown)                     │    │
│  │  - Window state management                                     │    │
│  │  - Event routing (Wails Events)                                │    │
│  │  - Systray integration                                         │    │
│  └────────────────────────────────────────────────────────────────┘    │
│           │                │                │                          │
│  ┌────────┴────────┐ ┌────┴──────┐ ┌─────┴────────┐                   │
│  │ Service Layer    │ │ Update     │ │ Single       │                   │
│  │                  │ │ Service    │ │ Instance     │                   │
│  │ - Config        │ │            │ │ Lock         │                   │
│  │ - Commit        │ │ - Check    │ │ - Mutex      │                   │
│  │ - Projects      │ │ - Download │ │ - IPC        │                   │
│  │ - Pushover      │ │ - Install  │ │              │                   │
│  └─────────────────┘ └────────────┘ └──────────────┘                   │
├─────────────────────────────────────────────────────────────────────────┤
│                      Data & Integration Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │ SQLite/GORM  │  │ Git          │  │ GitHub API   │                  │
│  │ Repository   │  │ Operations   │  │ (Updates)    │                  │
│  └──────────────┘  └──────────────┘  └──────────────┘                  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **App (app.go)** | Wails application lifecycle, window state, event routing, systray | Main Go struct with startup/shutdown hooks |
| **Single Instance Lock** | Prevent multiple app instances, handle second instance launch | Wails SingleInstanceLock with UniqueId + callback |
| **UpdateService** | Check GitHub releases, compare versions, find platform assets | HTTP client to GitHub Releases API |
| **Downloader** | Download update packages with progress tracking | HTTP client with io.Writer progress callbacks |
| **Installer** | Launch external updater.exe, pass arguments, exit main app | exec.Command with CREATE_NO_WINDOW flag |
| **Systray Integration** | System tray icon, menu, click/double-click handlers | github.com/getlantern/systray |
| **Vue3 Frontend** | User interface, event listeners, update notifications | Pinia stores + Wails event listeners |
| **CI/CD Pipeline** | Build, test, package, release automation | GitHub Actions workflows |

## Recommended Project Structure

```
ai-commit-hub/
├── main.go                    # Wails entry point, SingleInstanceLock config
├── app.go                     # App lifecycle, systray, window management
├── wails.json                 # Wails build configuration
│
├── pkg/
│   ├── service/
│   │   ├── update_service.go          # GitHub release checking
│   │   ├── config_service.go          # Configuration management
│   │   └── [other services].go        # Existing services
│   │
│   ├── update/
│   │   ├── downloader.go              # Download with progress
│   │   └── installer.go               # External updater launching
│   │
│   ├── singleinstance/                # NEW: Single instance management
│   │   ├── lock.go                    # Wails SingleInstanceLock wrapper
│   │   └── handler.go                 # Second instance callback logic
│   │
│   ├── systray/                       # NEW: Extract from app.go
│   │   ├── manager.go                 # Systray lifecycle management
│   │   ├── events.go                  # Click/double-click handlers
│   │   └── menu.go                    # Menu item creation/updates
│   │
│   └── [existing packages]...        # Repository, models, git, etc.
│
├── cmd/
│   └── updater/                       # NEW: External updater program
│       └── main.go                    # Zip extraction, process waiting
│
├── frontend/
│   ├── src/
│   │   ├── stores/
│   │   │   ├── updateStore.ts         # Update state management
│   │   │   └── [existing stores].ts   # Project, commit, etc.
│   │   │
│   │   ├── components/
│   │   │   ├── UpdateDialog.vue       # Update notification UI
│   │   │   └── [existing components].vue
│   │   │
│   │   └── events/                    # NEW: Wails event handling
│   │       ├── update.ts              # Update event listeners
│   │       └── window.ts              # Window event listeners
│   │
│   └── [wails build output]
│
└── .github/
    └── workflows/
        ├── build.yml                  # CI: Build and test
        └── release.yml                # CD: Automated releases
```

### Structure Rationale

- **pkg/singleinstance/**: Encapsulates single-instance logic, keeping app.go focused on lifecycle
- **pkg/systray/**: Extracts systray-specific code from app.go, improving testability and maintainability
- **cmd/updater/**: Separate program ensures clean process isolation during updates (no file locks)
- **frontend/src/events/**: Centralizes Wails event subscription logic, reduces component coupling
- **.github/workflows/**: Separates build verification from release automation

## Architectural Patterns

### Pattern 1: Single Instance Lock with Second Instance Forwarding

**What:** Prevents multiple application instances using platform-specific locking mechanisms (mutex on Windows, file lock on macOS, dbus on Linux). When a second instance attempts to launch, the first instance receives the arguments via callback.

**When to use:** Desktop applications that should consolidate user interactions into a single window, especially when handling file associations or deep links.

**Trade-offs:**
- ✅ Pros: Prevents resource waste, ensures consistent state, enables file association handling
- ❌ Cons: Adds complexity to callback handling, requires explicit window focus management

**Example:**
```go
// main.go
func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:     "AI Commit Hub",
        Width:     1280,
        Height:    800,
        OnStartup: app.startup,
        OnShutdown: app.shutdown,
        OnBeforeClose: app.onBeforeClose,
        SingleInstanceLock: &options.SingleInstanceLock{
            // Unique identifier for lock (use UUID or reverse domain notation)
            UniqueId: "com.github.allanpk716.ai-commit-hub",

            // Callback when second instance launches
            OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
        },
        Bind: []interface{}{app},
    })

    if err != nil {
        logger.Errorf("Error: %v", err)
    }
}

// app.go
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
    logger.Infof("Second instance launched from: %s", secondInstanceData.WorkingDirectory)
    logger.Infof("Arguments: %v", secondInstanceData.Args)

    // Bring existing window to front
    if a.ctx != nil {
        runtime.WindowShow(a.ctx)
        runtime.WindowUnminimise(a.ctx)

        // Emit event to frontend with second instance arguments
        runtime.EventsEmit(a.ctx, "second-instance-args", map[string]interface{}{
            "args":         secondInstanceData.Args,
            "workingDir":   secondInstanceData.WorkingDirectory,
        })
    }
}
```

### Pattern 2: External Updater Process Pattern

**What:** Main application downloads update package, launches external updater program, then exits. The updater waits for main process to terminate (releasing file locks), extracts the update, and relaunches the application.

**When to use:** Desktop applications on Windows that need to self-update without administrator privileges.

**Trade-offs:**
- ✅ Pros: Clean file lock handling, no in-process update complexity, supports full binary replacement
- ❌ Cons: Requires separate binary, more complex process management, potential updater process orphaning

**Example:**
```go
// pkg/update/installer.go
func (i *Installer) Install(updateZipPath string) error {
    // Get main app PID (for updater to wait on)
    pid := os.Getpid()

    // Get target directory (where main app is located)
    execPath, _ := os.Executable()
    targetDir := filepath.Dir(execPath)

    // Verify updater exists
    if _, err := os.Stat(i.updaterPath); os.IsNotExist(err) {
        return fmt.Errorf("updater not found: %s", i.updaterPath)
    }

    // Launch updater with arguments
    cmd := exec.Command(i.updaterPath,
        "--source", updateZipPath,
        "--target", targetDir,
        "--pid", strconv.Itoa(pid),
    )

    // Hide console window on Windows
    if runtime.GOOS == "windows" {
        cmd.SysProcAttr = &windows.SysProcAttr{
            CreationFlags: 0x08000000, // CREATE_NO_WINDOW
        }
    }

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start updater: %w", err)
    }

    // Exit main app immediately (updater waits for this process)
    logger.Info("Updater launched, exiting main application")
    return nil
}

// cmd/updater/main.go
func main() {
    // Parse command-line arguments
    sourceZip := flag.String("source", "", "Path to update zip file")
    targetDir := flag.String("target", "", "Target installation directory")
    pid := flag.Int("pid", 0, "Main application PID to wait for")
    flag.Parse()

    // Wait for main app to exit (release file locks)
    if *pid > 0 {
        process, _ := os.FindProcess(*pid)
        log.Printf("Waiting for PID %d to exit...", *pid)
        process.Wait()
        time.Sleep(1 * time.Second) // Additional buffer
    }

    // Extract update zip to target directory
    if err := extractZip(*sourceZip, *targetDir); err != nil {
        log.Fatalf("Extraction failed: %v", err)
    }

    // Relaunch application
    execPath := filepath.Join(*targetDir, "ai-commit-hub.exe")
    exec.Command(execPath).Start()

    log.Println("Update complete, exiting updater")
}
```

### Pattern 3: Wails Event-Driven Update Flow

**What:** Backend checks for updates asynchronously on startup, emits Wails Events to frontend when updates are available. Frontend listens for events and displays non-intrusive notifications. User initiates download/install, with progress events streamed via Wails Events.

**When to use:** Desktop applications where updates should not block startup and should provide user control over installation timing.

**Trade-offs:**
- ✅ Pros: Non-blocking startup, real-time progress updates, clean separation of concerns
- ❌ Cons: More complex event handling, requires frontend state management

**Example:**
```go
// Backend (app.go startup)
go func() {
    updateInfo, err := a.updateService.CheckForUpdates()
    if err != nil {
        logger.Warnf("Update check failed: %v", err)
        return
    }

    if updateInfo.HasUpdate {
        logger.Info("Update available", "version", updateInfo.LatestVersion)
        // Emit event to frontend
        runtime.EventsEmit(ctx, "update-available", updateInfo)
    }
}()

// Backend (app.go InstallUpdate)
func (a *App) InstallUpdate(downloadURL, assetName string) error {
    downloader := update.NewDownloader(tempDir)

    // Set up progress callback
    downloader.SetProgressFunc(func(downloaded, total int64) {
        // Stream progress to frontend
        runtime.EventsEmit(a.ctx, "download-progress", map[string]interface{}{
            "downloaded": downloaded,
            "total":      total,
            "percentage": float64(downloaded) / float64(total) * 100,
        })
    })

    // Download and emit events
    zipPath, err := downloader.Download(downloadURL, assetName)
    runtime.EventsEmit(a.ctx, "download-complete", map[string]interface{}{
        "path": zipPath,
    })

    // Launch installer and exit
    return a.installer.Install(zipPath)
}

// Frontend (stores/updateStore.ts)
export const useUpdateStore = defineStore('update', () => {
    const updateAvailable = ref(false)
    const updateInfo = ref<UpdateInfo | null>(null)
    const downloadProgress = ref(0)

    // Listen for update events
    EventsOn("update-available", (info: UpdateInfo) => {
        updateInfo.value = info
        updateAvailable.value = true
    })

    EventsOn("download-progress", (progress: { downloaded: number; total: number; percentage: number }) => {
        downloadProgress.value = progress.percentage
    })

    EventsOn("download-complete", (data: { path: string }) => {
        console.log("Download complete:", data.path)
        // Show "Restart to Update" button
    })

    return {
        updateAvailable,
        updateInfo,
        downloadProgress,
    }
})
```

### Pattern 4: Systray Event Integration

**What:** System tray runs in separate goroutine with proper lifecycle management. Click/double-click events show or hide the main window. Menu items provide common actions (quit, settings). Systray cleanup is coordinated with Wails shutdown.

**When to use:** Desktop applications that should minimize to tray instead of closing, providing persistent background presence.

**Trade-offs:**
- ✅ Pros: User-friendly minimization, quick access to common actions, persistent background operation
- ❌ Cons: Additional lifecycle complexity, platform-specific behavior differences, requires careful shutdown ordering

**Example:**
```go
// app.go
type App struct {
    ctx              context.Context
    systrayReady     chan struct{}
    systrayExit      *sync.Once
    windowVisible    bool
    windowMutex      sync.RWMutex
    systrayRunning   atomic.Bool
    quitting         atomic.Bool
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    // Start systray in separate goroutine with delay
    go func() {
        time.Sleep(300 * time.Millisecond)
        a.runSystray()
    }()
}

func (a *App) runSystray() {
    runtime.LockOSThread()
    a.systrayRunning.Store(true)

    systray.Run(
        a.onSystrayReady,
        func() {
            a.systrayRunning.Store(false)
            a.onSystrayExit()
        },
    )
}

func (a *App) onSystrayReady() {
    // Set icon and tooltip
    systray.SetIcon(a.getTrayIcon())
    systray.SetTooltip("AI Commit Hub - 双击打开主窗口")

    // Create menu items
    showMenu := systray.AddMenuItem("显示窗口", "显示主窗口")
    go func() {
        for range showMenu.ClickedCh {
            a.showWindow()
        }
    }()

    quitMenu := systray.AddMenuItem("退出应用", "完全退出应用")
    go func() {
        for range quitMenu.ClickedCh {
            a.quitApplication()
        }
    }()

    close(a.systrayReady)
}

func (a *App) showWindow() {
    a.windowMutex.Lock()
    defer a.windowMutex.Unlock()

    if a.windowVisible {
        return
    }

    runtime.WindowShow(a.ctx)
    a.windowVisible = true

    // Emit event to frontend
    runtime.EventsEmit(a.ctx, "window-shown", map[string]interface{}{
        "timestamp": time.Now(),
    })
}

func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        a.quitting.Store(true)
        a.showWindow() // Show window before exit to avoid "stuck" appearance
        systray.Quit() // Triggers onSystrayExit
    })
}

func (a *App) onSystrayExit() {
    if a.quitting.Load() {
        runtime.Quit(a.ctx) // Trigger Wails shutdown
    }
}

func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    if a.quitting.Load() {
        return false // Allow close
    }

    a.hideWindow() // Minimize to tray instead
    return true    // Prevent close
}
```

## Data Flow

### Startup Flow

```
[User Launches App]
    ↓
[main.go: wails.Run]
    ↓
[app.startup]
    ├─→ [Initialize Database, Services]
    ├─→ [Start Systray goroutine] → [systray.Run]
    ├─→ [Async: Check for Updates] → [GitHub API]
    │   └─→ [Emit "update-available" event]
    └─→ [Async: Preload Project Statuses]
        └─→ [Emit "startup-complete" event]
    ↓
[Frontend: Show Main Window]
```

### Update Check and Install Flow

```
[Background Check on Startup]
    ↓
[UpdateService.CheckForUpdates]
    ├─→ [Fetch GitHub Release API]
    ├─→ [Compare Versions]
    └─→ [Emit "update-available" event]
    ↓
[Frontend: Display Notification]
    ↓
[User Clicks "Update"]
    ↓
[App.InstallUpdate]
    ├─→ [Create Downloader]
    ├─→ [Download with Progress]
    │   └─→ [Emit "download-progress" events]
    ├─→ [Emit "download-complete" event]
    └─→ [Installer.Install]
        ├─→ [Launch updater.exe]
        ├─→ [Pass: source zip, target dir, main PID]
        └─→ [Exit Main App]
    ↓
[External Updater Process]
    ├─→ [Wait for main PID to exit]
    ├─→ [Extract update zip]
    └─→ [Launch updated app]
```

### Single Instance Second Instance Launch Flow

```
[User Launches Second Instance]
    ↓
[Wails SingleInstanceLock Detection]
    ↓
[First Instance: OnSecondInstanceLaunch Callback]
    ├─→ [Log second instance args]
    ├─→ [runtime.WindowShow]
    ├─→ [runtime.WindowUnminimise]
    └─→ [Emit "second-instance-args" event]
    ↓
[Frontend: Handle Arguments (if any)]
```

### Window Close/Minimize Flow

```
[User Clicks Window Close Button]
    ↓
[app.onBeforeClose]
    ├─→ [Check quitting flag]
    │   ├─→ [If true: Allow close (return false)]
    │   └─→ [If false: Minimize to tray]
    │       ├─→ [runtime.WindowHide]
    │       ├─→ [Set windowVisible = false]
    │       └─→ [Emit "window-hidden" event]
    └─→ [Return true (prevent close)]
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-1k users | Current architecture is optimal. Single-instance lock prevents duplicate processes. Update server load from GitHub API is negligible. |
| 1k-100k users | Consider caching GitHub release API responses (add ETag/If-Modified-Since support). Add telemetry for update download success rates. Implement staged rollouts for updates (release to percentage of users). |
| 100k+ users | Migrate from GitHub Releases API to dedicated update server (CDN-backed). Implement delta updates to reduce bandwidth. Add A/B testing infrastructure for update UI/flows. Consider background auto-update with user opt-out. |

### Scaling Priorities

1. **First bottleneck:** GitHub API rate limiting (60 requests/hour for unauthenticated IP addresses).
   - **Fix:** Implement response caching with 1-hour TTL. Add authentication if needed.

2. **Second bottleneck:** Update server bandwidth during major releases.
   - **Fix:** Use GitHub Releases for hosting (leveraging GitHub's CDN infrastructure). Consider delta updates for large binary sizes.

3. **Third bottleneck:** User notification spam (update available shown repeatedly).
   - **Fix:** Store "dismissed version" in user config, only notify for new versions.

## Anti-Patterns

### Anti-Pattern 1: Blocking Startup with Update Check

**What people do:** Calling `CheckForUpdates()` synchronously in `app.startup()`, preventing the application window from showing until the HTTP request completes.

**Why it's wrong:** Update servers can be slow or unreachable. A 10-second timeout blocks UI rendering, making the app appear broken or frozen on poor connections.

**Do this instead:** Always perform update checks asynchronously in a goroutine. Emit a Wails Event when the check completes. Show the window immediately, display update notification later if needed.

```go
// ❌ BAD: Blocks startup
func (a *App) startup(ctx context.Context) {
    // ...other initialization...

    updateInfo, err := a.updateService.CheckForUpdates() // BLOCKS
    if err != nil {
        // Handle error
    }
    // Window still hasn't shown yet
}

// ✅ GOOD: Non-blocking
func (a *App) startup(ctx context.Context) {
    // ...other initialization...

    go func() {
        updateInfo, err := a.updateService.CheckForUpdates()
        if err != nil {
            logger.Warnf("Update check failed: %v", err)
            return
        }
        if updateInfo.HasUpdate {
            runtime.EventsEmit(ctx, "update-available", updateInfo)
        }
    }()
    // Window shows immediately
}
```

### Anti-Pattern 2: In-Process Update Extraction

**What people do:** Downloading update zip and extracting it directly in the main application process.

**Why it's wrong:** The main application's executable file is locked by the OS on Windows. Attempting to overwrite it fails with "access denied" or "file in use" errors.

**Do this instead:** Always use an external updater process (pattern #2). The main app must exit before the updater can overwrite files.

```go
// ❌ BAD: In-process extraction
func (a *App) InstallUpdate(zipPath string) error {
    // This will fail on Windows - ai-commit-hub.exe is locked
    return extractZip(zipPath, getAppDir())
}

// ✅ GOOD: External updater
func (a *App) InstallUpdate(zipPath string) error {
    cmd := exec.Command("updater.exe", "--source", zipPath, "--pid", os.Getpid())
    cmd.Start()
    return nil // Main app exits, updater takes over
}
```

### Anti-Pattern 3: Systray Shutdown Race Conditions

**What people do:** Calling `systray.Quit()` in `app.shutdown()` without checking if it's already been called, leading to duplicate cleanup or deadlocks.

**Why it's wrong:** The systray exit handler might call `runtime.Quit()`, which triggers `app.shutdown()`, creating a circular call chain. Duplicate `systray.Quit()` calls can cause hangs.

**Do this instead:** Use `sync.Once` to ensure systray cleanup runs exactly once. Use a `quitting` flag to differentiate between "minimize to tray" and "actually quit" scenarios.

```go
// ✅ GOOD: sync.Once prevents duplicate cleanup
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        a.quitting.Store(true)
        a.showWindow()
        systray.Quit() // Only called once
    })
}

func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
    if a.quitting.Load() {
        return false // Allow close
    }
    a.hideWindow() // Minimize to tray
    return true
}
```

### Anti-Pattern 4: Direct systray.Quit() in OnShutdown

**What people do:** Calling `systray.Quit()` directly in the Wails `OnShutdown` handler without proper sequencing.

**Why it's wrong:** This creates a circular shutdown loop:
1. `OnShutdown` calls `systray.Quit()`
2. `systray.Quit()` triggers the systray exit callback
3. Exit callback calls `runtime.Quit()`
4. `runtime.Quit()` triggers `OnShutdown` again
5. Infinite loop or deadlock

**Do this instead:** The systray quit should be initiated by user action (clicking "Quit" menu item), not by the shutdown handler. Use the `quitting` flag to prevent loops.

```go
// ✅ GOOD: Systray quit triggered by user, not shutdown
func (a *App) quitApplication() {
    a.systrayExit.Do(func() {
        a.quitting.Store(true)
        systray.Quit() // User-initiated, not in shutdown
    })
}

func (a *App) shutdown(ctx context.Context) {
    // Do NOT call systray.Quit() here
    logger.Info("Application shutting down")
}
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **GitHub Releases API** | RESTful HTTP client via `UpdateService` | Check for updates on startup. Handle rate limiting with caching. |
| **GitHub Actions** | YAML workflow files in `.github/workflows/` | Separate workflows for CI (build/test) and CD (release). Use `softprops/action-gh-release` for releases. |
| **OS File Locking** | Wails `SingleInstanceLock` abstraction | Platform-specific: mutex (Windows), file lock (macOS), dbus (Linux) |
| **OS Process Management** | `os/exec` for launching updater.exe | Pass PID via command-line, use `process.Wait()` for synchronization |
| **Wails Runtime Events** | `runtime.EventsEmit` / `EventsOn` | Event-driven communication between Go backend and Vue frontend |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| **Frontend ↔ Backend** | Wails Bindings + Wails Events | Bindings for RPC-style calls (e.g., `InstallUpdate()`). Events for async notifications (e.g., `download-progress`) |
| **App ↔ Services** | Direct method calls (Go) | Services are stateless, accept context, return structs. Keep logic in services, not in `app.go` |
| **Main App ↔ Updater** | Command-line args + Process waiting | Pass: `--source <zip>`, `--target <dir>`, `--pid <main_pid>`. Updater waits for main PID to exit before extraction |
| **App ↔ Systray** | Goroutine + channels | Systray runs in separate goroutine. Use `systrayReady` channel to signal initialization. Use `atomic.Bool` for runtime status |
| **UpdateService ↔ GitHub** | HTTP GET requests | Use `http.Client` with timeout. Parse JSON response from GitHub Releases API |

## CI/CD Integration Architecture

### Build Workflow (`.github/workflows/build.yml`)

**Trigger:** Push to `main` branch, pull requests

**Steps:**
1. Checkout code
2. Set up Go (1.24.1)
3. Cache Go modules
4. Run tests: `go test ./...`
5. Build Wails app: `wails build`
6. Upload artifact (Windows binary)

**Purpose:** Verify code quality, catch regressions before release

### Release Workflow (`.github/workflows/release.yml`)

**Trigger:** Git tag pushed (e.g., `v1.2.3`)

**Steps:**
1. Checkout code
2. Set up Go
3. Fetch Git tags (needed for versioning)
4. Run tests
5. Build Wails app
6. Package update assets:
   - `ai-commit-hub-windows.zip` (main app)
   - `updater.exe` (external updater)
7. Create GitHub Release
8. Attach assets to release
9. Update `latest` tag

**Purpose:** Automated releases with assets compatible with in-app update mechanism

**Key Integration Points:**
- Build must produce zip file with updater.exe included
- Version number from Git tag must be embedded in binary (`pkg/version`)
- Release notes should be auto-generated from git commits
- Assets must be named predictably for `UpdateService` to find platform-specific downloads

## Build Order and Dependencies

### Phase 1: Foundation (No external dependencies)
1. **Single Instance Lock** (`pkg/singleinstance/`)
   - Pure Wails configuration, no services needed
   - Test by launching app twice, verifying second instance forwards args
   - **Blocks:** Nothing (can be implemented immediately)

2. **Systray Extraction** (`pkg/systray/`)
   - Extract existing systray code from `app.go`
   - Improves code organization, no new behavior
   - **Blocks:** Phase 2 (cleaner integration point for tray events)

### Phase 2: Update Infrastructure (Depends on Phase 1)
3. **UpdateService Enhancement** (`pkg/service/update_service.go`)
   - Add GitHub release caching (ETag support)
   - Add version comparison helpers
   - **Blocks:** Phase 3 (downloader needs asset URLs)

4. **Downloader with Progress** (`pkg/update/downloader.go`)
   - HTTP client with progress callbacks
   - Wails Events integration for progress streaming
   - **Blocks:** Phase 3 (installer needs downloaded files)

5. **Installer** (`pkg/update/installer.go`)
   - External updater launching logic
   - **Blocks:** Phase 3 (needs updater.exe binary)

### Phase 3: Update Execution (Depends on Phase 2)
6. **External Updater** (`cmd/updater/main.go`)
   - Zip extraction, process waiting, relaunch
   - **Blocks:** Phase 4 (testing requires complete update flow)

7. **Frontend Update Store** (`frontend/src/stores/updateStore.ts`)
   - Event listeners for update notifications
   - Progress tracking UI state
   - **Blocks:** Phase 4 (integration testing requires UI)

### Phase 4: Integration & Automation (Depends on Phase 3)
8. **CI/CD Workflows** (`.github/workflows/`)
   - Build workflow (test + artifact upload)
   - Release workflow (tag-triggered release creation)
   - **Blocks:** Final testing (requires release artifacts)

9. **End-to-End Testing**
   - Manual testing: Install app, trigger update, verify replacement
   - **Blocks:** Production release

### Critical Path

```
Phase 1: Single Instance Lock
    ↓
Phase 2: UpdateService + Downloader + Installer
    ↓
Phase 3: Updater.exe + Frontend UpdateStore
    ↓
Phase 4: CI/CD + E2E Testing
```

**Parallelization Opportunities:**
- Phase 1.2 (Systray Extraction) can be done concurrently with Phase 1.1
- Phase 2.3, 2.4, 2.5 can be developed in parallel (independent modules)
- Phase 3.6 and 3.7 can be developed in parallel (backend + frontend)

## Sources

### Wails Documentation (HIGH Confidence)
- [Wails v2.11.0 Documentation - Single Instance Lock](https://github.com/wailsapp/wails/blob/master/website/docs/guides/single-instance-lock.mdx) - Official guide for implementing single-instance applications
- [Wails v2 File Association Guide](https://github.com/wailsapp/wails/blob/master/website/versioned_docs/version-v2.10/guides/file-association.mdx) - Single instance usage for file association
- [Wails v3 Alpha Single Instance Guide](https://v3alpha.wails.io/guides/single-instance/) - Future API design shows evolution (for reference)

### Go Single Instance Patterns (MEDIUM Confidence)
- [Stack Overflow: Go Windows Single Instance](https://stackoverflow.com/questions/28889818/preventing-multiple-instances-of-a-program) - Community implementation patterns
- [GitHub Discussion: Communication between multiple instances #2441](https://github.com/wailsapp/wails/discussions/2441) - Real-world multi-instance issues

### GitHub Actions for Go Releases (MEDIUM Confidence)
- [Go Release Binary Action (Marketplace)](https://github.com/marketplace/actions/go-release-binary) - Pre-built action for Go releases
- [GoReleaser Documentation](https://goreleaser.com/) - Industry standard for Go release automation (mentioned in community discussions)
- [Using GitHub Actions to add Go binaries to a Release](https://akrabat.com/using-github-actions-to-add-go-binaries-to-a-release/) - Practical tutorial (February 2024)

### Existing Codebase Analysis (HIGH Confidence)
- `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\app.go` - Current systray and lifecycle implementation
- `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\service\update_service.go` - Existing update checking logic
- `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\pkg\update\` - Current downloader and installer patterns
- `C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\main.go` - Wails configuration and startup

---
*Architecture research for: AI Commit Hub Wails Application Enhancements*
*Researched: 2026-02-06*
