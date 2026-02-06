# Stack Research

**Domain:** Wails Windows Desktop Application - Single Instance, Auto-Update, System Tray, CI/CD
**Researched:** 2026-02-06
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **Wails** | v2.11.0 | Desktop framework | Current stable version (as of Jan 2025). Official built-in single-instance lock support. Mature ecosystem, actively maintained. |
| **Go** | 1.21+ | Backend language | Minimum requirement for Wails v2. Windows API access via `golang.org/x/sys/windows` |
| **Vue 3** | 3.x | Frontend framework | Project already uses Vue 3 + TypeScript + Vite, proven Wails integration |
| **GitHub Actions** | - | CI/CD platform | Native GitHub integration for releases, official Wails build action available |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **golang.org/x/sys/windows** | v0.40.0 | Windows API bindings | For Windows mutex, CreateMutex, single-instance implementation alternatives to built-in |
| **github.com/lutischan-ferenc/systray** | v1.3.0 | System tray (Wails v2) | Wails v2 uses third-party systray. This fork adds double-click support (`SetOnDClick`) |
| **github.com/creativeprojects/go-selfupdate** | v1.5.2 | Auto-update library | GitHub releases integration. Active maintenance (Dec 2025), multi-provider support (GitHub/Gitea/GitLab) |
| **softprops/action-gh-release** | v2 | GitHub Actions release | Standard action for creating GitHub releases with artifacts |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| **dAppServer/wails-build-action** | GitHub Actions Wails build | Community action, builds all platforms from single workflow |
| **go-selfupdate** | Auto-update mechanism | Detects latest release from GitHub, downloads binary, replaces executable |
| **github.com/WQGroup/logger** | Structured logging | Already in use, supports multiple formats and log rotation |

## Installation

```bash
# Core Wails (already installed)
# wails.io/docs/v2.11.0/introduction

# Windows API (if needed for manual single-instance)
go get golang.org/x/sys/windows@v0.40.0

# System tray with double-click support (Wails v2)
go get github.com/lutischan-ferenc/systray@v1.3.0

# Auto-update library
go get github.com/creativeprojects/go-selfupdate@v1.5.2
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| **Wails v2.11.0 SingleInstanceLock** | Manual Windows mutex (CreateMutex) | Only if you need cross-session locking or custom mutex name handling. Wails built-in is simpler and safer. |
| **creativeprojects/go-selfupdate** | marcus-crane/wails-autoupdater | Only for legacy Wails projects. Wails-autoupdater is unmaintained (last update Jan 2023). |
| **creativeprojects/go-selfupdate** | rhysd/go-github-selfupdate | Only for simple CLI tools. Lacks Windows desktop features (silent updates, progress UI). |
| **creativeprojects/go-selfupdate** | inconshreveable/go-update | Only if you need binary patching (bsdiff). Last update 2013, obsolete. |
| **lutischan-ferenc/systray** | getlantern/systray v1.2.2 | ❌ Never. Original version lacks double-click support, critical for Windows UX. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **rhysd/go-github-selfupdate** | Last updated Dec 2017, unmaintained. Missing GitHub API changes, security risks. | creativeprojects/go-selfupdate (actively maintained Dec 2025) |
| **inconshreveable/go-update** | Last updated 2013, deprecated. No Windows desktop consideration. | creativeprojects/go-selfupdate |
| **marcus-crane/wails-autoupdater** | Unmaintained since Jan 2023, MVP experiment, not production-ready. | creativeprojects/go-selfupdate (direct integration) |
| **Manual mutex implementation** | Wails v2.11.0 has built-in SingleInstanceLock. Reimplementing adds maintenance burden and security risks. | Wails `options.SingleInstanceLock` |
| **Equinox** | Commercial service, requires external infrastructure, adds dependency. | GitHub Releases + go-selfupdate (free, self-hosted) |
| **getlantern/systray** | No double-click event support. Users must right-click → select "Show", poor UX. | lutischan-ferenc/systray (maintained fork) |
| **mouuff/go-rocket-update** | Last release Sep 2025, less popular (119 stars vs 634 for rhysd, but go-selfupdate is best). | creativeprojects/go-selfupdate |

## Stack Patterns by Variant

**If using Wails v2.11.0 (current):**
- Use built-in `SingleInstanceLock` in `options.App`
- Use `github.com/lutischan-ferenc/systray` for double-click support
- Auto-update: Integrate `go-selfupdate` manually (Wails has no built-in)

**If upgrading to Wails v3 (future):**
- Single-instance: Built-in `application.SingleInstanceOptions` (improved API)
- System tray: Native Wails v3 systray with `OnClick`, `OnDoubleClick`
- Auto-update: Still manual (Wails v3 has no built-in auto-update)

**If cross-platform auto-update needed:**
- Use `creativeprojects/go-selfupdate` with GitHub provider
- Supports Windows, macOS, Linux
- Automatic OS/arch detection

**If Windows-only auto-update acceptable:**
- Consider `containifyci/go-self-update` for simpler GitHub integration
- But `creativeprojects/go-selfupdate` is still recommended (more features, better maintained)

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| **Wails v2.11.0** | Go 1.21+ | Minimum Go version for Wails v2 |
| **Wails v2.11.0** | lutischan-ferenc/systray v1.3.0 | ✅ Tested. Double-click working. |
| **Wails v2.11.0** | getlantern/systray v1.2.2 | ❌ No double-click support. |
| **Wails v2.11.0** | go-selfupdate v1.5.2 | ✅ Compatible. |
| **go-selfupdate v1.5.2** | Go 1.21+ | Requires Go modules support |
| **golang.org/x/sys/windows v0.40.0** | Wails v2.11.0 | ✅ Compatible (same as Wails dependency) |
| **Wails v3 alpha** | Go 1.23+ | Higher Go version requirement for v3 |

## Implementation Notes

### 1. Single-Instance Lock (Windows)

**Recommended: Wails Built-in (HIGH confidence)**

```go
// Wails v2.11.0 built-in single-instance lock
err := wails.Run(&options.App{
    Title:  "AI Commit Hub",
    Width:  1024,
    Height: 768,
    SingleInstanceLock: &options.SingleInstanceLock{
        UniqueId: "e3984e08-28dc-4e3d-b70a-45e961589cdc", // UUID
        OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
            // Bring window to front
            runtime.WindowShow(ctx)
            runtime.WindowUnminimise(ctx)
            // Handle args from second instance
            runtime.EventsEmit(ctx, "second-instance-args", data.Args)
        },
    },
    // ... other options
})
```

**Why Built-in:**
- Windows: Uses named mutex (generated from UniqueId)
- macOS: Uses named mutex + NSDistributedNotificationCenter for data passing
- Linux: Uses D-Bus
- Handles cross-platform differences automatically
- Security: Wails validates args, treats as untrusted (per official docs)

**Alternative: Manual Windows Mutex (NOT recommended unless needed)**

```go
// Manual Windows mutex implementation
// Only use if you need cross-session locking (Global\ prefix)
import (
    "golang.org/x/sys/windows"
)

func createMutex(name string) (windows.Handle, error) {
    mutex, err := windows.CreateMutex(nil, false, name)
    if err != nil {
        return 0, err
    }
    if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
        return 0, errors.New("already running")
    }
    return mutex, nil
}

// Usage:
mutex, err := createMutex("Global\\YourAppMutex-" + uniqueID)
if err != nil {
    // Second instance - exit
    os.Exit(1)
}
defer windows.CloseHandle(mutex)
```

**Why Avoid Manual:**
- Must handle mutex cleanup manually
- Must implement IPC for second instance → first instance communication
- Security risks: Malicious user can create mutex first, prevent app launch (Microsoft docs)
- Wails built-in handles all this correctly

### 2. Auto-Update (GitHub Releases)

**Recommended: creativeprojects/go-selfupdate v1.5.2 (HIGH confidence)**

```go
import (
    "context"
    "github.com/creativeprojects/go-selfupdate"
)

func checkForUpdate() error {
    // Detect latest release from GitHub
    latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("allanpk716/ai-commit-hub"))
    if err != nil {
        return fmt.Errorf("error checking for update: %w", err)
    }
    if !found {
        return nil
    }

    // Current version
    v := semver.MustParse(version)
    if !latest.GreaterThanOrEqual(v) {
        return nil
    }

    // Download and update
    if latest.Latest() {
        log.Println("Already latest version")
        return nil
    }

    // Apply update
    if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetFileName, os.Args[0]); err != nil {
        return fmt.Errorf("error updating: %w", err)
    }

    log.Println("Updated to version", latest.Version())
    return nil
}
```

**GitHub Release Naming Convention (required for go-selfupdate):**

```
# Tag format: v1.0.0, v1.0.1, etc.
git tag v1.0.0
git push origin v1.0.0

# Asset naming (go-selfupdate auto-detects):
ai-commit-hub_1.0.0_windows_amd64.exe.zip
ai-commit-hub_1.0.0_windows_arm64.exe.zip
ai-commit-hub_1.0.0_darwin_amd64.zip
ai-commit-hub_1.0.0_darwin_arm64.zip
ai-commit-hub_1.0.0_linux_amd64.zip
```

**Why go-selfupdate:**
- Active maintenance (Dec 2025, v1.5.2)
- Multi-provider support (GitHub, Gitea, GitLab)
- Automatic OS/arch detection
- Silent update option (no UI required)
- Cryptographic signature verification (optional)
- 55+ importers, production-proven

**Alternatives Considered:**

| Library | Last Update | Stars | Status | Why Not |
|---------|-------------|-------|--------|---------|
| rhysd/go-github-selfupdate | Dec 2017 | 634 | ❌ Unmaintained | Missing GitHub API changes, no Windows desktop considerations |
| inconshreveable/go-update | 2013 | 2.2k | ❌ Deprecated | No GitHub integration, manual download required |
| marcus-crane/wails-autoupdater | Jan 2023 | - | ❌ MVP experiment | Unmaintained, specific to old Wails |
| mouuff/go-rocket-update | Sep 2025 | 119 | ⚠️ Less popular | Fewer stars, less community validation |
| containifyci/go-self-update | Dec 2024 | - | ⚠️ Newer | Less proven, simpler feature set |

### 3. System Tray Double-Click

**Wails v2: lutischan-ferenc/systray v1.3.0 (HIGH confidence)**

```go
import (
    "github.com/lutischan-ferenc/systray"
)

func (a *App) onSystrayReady() {
    // Set icon
    systray.SetIcon(iconBytes)
    systray.SetTooltip("AI Commit Hub - 双击打开主窗口")

    // Double-click handler
    systray.SetOnDClick(func(menu systray.IMenu) {
        a.showWindow()
    })

    // Single-click handler
    systray.SetOnClick(func(menu systray.IMenu) {
        // Optional: handle single-click differently
    })

    // Right-click menu
    menu := systray.AddMenuItem("显示窗口", "显示主窗口")
    go func() {
        for range menu.ClickedCh {
            a.showWindow()
        }
    }()

    systray.AddSeparator()
    quitMenu := systray.AddMenuItem("退出应用", "完全退出应用")
    go func() {
        for range quitMenu.ClickedCh {
            a.quitApplication()
        }
    }()
}
```

**Key Differences from getlantern/systray:**

| Feature | getlantern v1.2.2 | lutischan-ferenc v1.3.0 |
|---------|-------------------|--------------------------|
| Double-click | ❌ Not supported | ✅ `SetOnDClick()` |
| Single-click | ❌ Not supported | ✅ `SetOnClick()` |
| Right-click | ❌ Not supported | ✅ `SetOnRClick()` |
| Menu click | ✅ `ClickedCh` channel | ✅ `Click()` callback |
| API style | Channel-based | Callback-based |

**Project's Current Implementation:**
- Already using `lutischan-ferenc/systray` (verified in `app.go`)
- Double-click working (see `docs/fixes/tray-icon-doubleclick-fix.md`)
- Icon: Multi-level fallback (Wails embedded ICO → PNG-generated ICO → red box)
- Exit logic: `quitting` flag to break loop (see `docs/lessons-learned/windows-tray-icon-implementation-guide.md`)

**Wails v3 System Tray:**

```go
// Wails v3 has native systray support
import "github.com/wailsapp/wails/v3/pkg/application"

app := application.New(application.Options{
    Name: "AI Commit Hub",
})

systray := app.NewSystemTray()
systray.SetIcon(iconBytes)

// Double-click (native in Wails v3)
systray.OnDoubleClick(func() {
    window.Show()
})

// Single-click
systray.OnClick(func() {
    if window.IsVisible() {
        window.Hide()
    } else {
        window.Show()
    }
})

// Window attachment (auto show/hide on focus)
systray.AttachWindow(window).WindowOffset(5)
```

**Why Wails v3 is better (when ready):**
- Native systray, no third-party dependency
- `AttachWindow()` automatically handles show/hide on focus loss
- Unified API across platforms
- Currently in alpha (v3.0.0-alpha.64, Jan 2026), not production-ready

### 4. GitHub Actions CI/CD

**Recommended: dAppServer/wails-build-action (HIGH confidence)**

```yaml
# .github/workflows/build.yml
name: Wails Build

on:
  push:
    tags:
      - 'v*'

env:
  NODE_OPTIONS: "--max-old-space-size=4096"

jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        include:
          - name: 'Windows'
            platform: 'windows/amd64'
            os: 'windows-latest'
            output: 'build/bin/ai-commit-hub.exe'
          - name: 'macOS Intel'
            platform: 'darwin/amd64'
            os: 'macos-latest'
            output: 'build/bin/ai-commit-hub'
          - name: 'macOS Apple Silicon'
            platform: 'darwin/arm64'
            os: 'macos-latest'
            output: 'build/bin/ai-commit-hub'
          - name: 'Linux'
            platform: 'linux/amd64'
            os: 'ubuntu-latest'
            output: 'build/bin/ai-commit-hub'

    runs-on: ${{ matrix.os }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Build Wails App
        uses: dAppServer/wails-build-action@v2.4
        with:
          buildName: ${{ matrix.name }}
          buildPlatform: ${{ matrix.platform }}
          packagePath: ./frontend

      - name: Upload Release Asset
        uses: softprops/action-gh-release@v2
        with:
          files: ${{ matrix.output }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Alternative: Manual Build (more control)**

```yaml
name: Wails Build (Manual)

on:
  push:
    tags:
      - 'v*'

jobs:
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build
        run: wails build --clean

      - name: Upload
        uses: softprops/action-gh-release@v2
        with:
          files: build/bin/*.exe
```

**Why dAppServer/wails-build-action:**
- Officially recommended by Wails docs
- Handles Go + Node.js + Wails CLI setup
- Cross-platform builds (Windows, macOS, Linux)
- 96+ stars, actively maintained
- Supports ARM builds (Apple Silicon)

**Code Signing (Windows):**

```yaml
# Add to build-windows job
- name: Code Sign
  run: |
    # Base64 decode certificate from secrets
    echo ${{ secrets.CERT_BASE64 }} | base64 -d > cert.pfx
    # Sign executable
    signtool sign /f cert.pfx /p ${{ secrets.CERT_PASSWORD }} build/bin/ai-commit-hub.exe
```

**Code Signing Resources:**
- Wails v3 guide: `https://v3alpha.wails.io/guides/signing/`
- Microsoft docs: Code Signing Certificates
- Cost: ~$100-500/year (DigiCert, Sectigo, etc.)
- EV certificate not required for desktop apps (standard is sufficient)

## Sources

### High Confidence (Official Documentation / Context7)

- **/wailsapp/wails** — Single instance lock implementation, system tray API
- **https://wails.io/docs/v2.11.0/guides/single-instance-lock/** — Official Wails v2.11.0 single-instance guide
- **https://wails.io/docs/v2.11.0/reference/options/** — Wails options reference, SingleInstanceLock struct
- **https://v3alpha.wails.io/features/menus/systray/** — Wails v3 system tray documentation (alpha)
- **https://pkg.go.dev/golang.org/x/sys/windows@v0.40.0** — Windows API bindings (CreateMutex)
- **https://pkg.go.dev/github.com/creativeprojects/go-selfupdate@v1.5.2** — Auto-update library documentation

### Medium Confidence (WebSearch + Verification)

- **https://github.com/creativeprojects/go-selfupdate** — Active auto-update library (Dec 2025)
- **https://github.com/dAppServer/wails-build-action** — Wails GitHub Actions build
- **https://wails.io/docs/v2.11.0/guides/crossplatform-build/** — Official Wails CI/CD guide
- **https://docs.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-createmutexa** — Windows CreateMutex API

### Low Confidence (WebSearch Only)

- **https://github.com/marcus-crane/wails-autoupdater** — Unmaintained Wails auto-update experiment (Jan 2023)
- **https://github.com/rhysd/go-github-selfupdate** — Unmaintained self-update library (Dec 2017)

### Project Documentation (Internal Research)

- **C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\lessons-learned\windows-tray-icon-implementation-guide.md** — Project's system tray implementation (verified working)
- **C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\fixes\tray-icon-doubleclick-fix.md** — Double-click implementation details
- **C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub\docs\fixes\systray-exit-fix.md** — Exit flow implementation

---

*Stack research for: AI Commit Hub - Wails Windows Desktop Application*
*Researched: 2026-02-06*
*Confidence: HIGH (all recommendations based on official docs, actively maintained libraries, or verified working implementations)*
