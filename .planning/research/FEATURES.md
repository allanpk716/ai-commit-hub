# Feature Research

**Domain:** Windows Desktop Applications (Lightweight Developer Tools)
**Researched:** 2026-02-06
**Confidence:** MEDIUM
**Project Context:** AI Commit Hub - Lightweight Git commit message generation helper

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Single-Instance Enforcement** | Prevents confusion, resource conflicts, and duplicate processes. Users expect double-clicking an already-running app to activate the existing window, not show an error or open a second instance. | MEDIUM | Requires Mutex for detection + Named Pipes for IPC to pass arguments to existing instance. Must handle window activation properly (bring to foreground). |
| **Minimize to System Tray** | Standard Windows desktop app behavior since Windows 95. Users expect close button to minimize to tray, not exit. | MEDIUM | Requires intercepting window close event (`onBeforeClose`), hiding window instead of quitting. Need proper exit flow via tray menu. |
| **Tray Icon Right-Click Menu** | Universal Windows UI convention. Right-click must show context menu with standard options (Show, Exit). | LOW | Use systray library (`github.com/lutischan-ferenc/systray` or `getlantern/systray`). Menu items: "显示窗口", "退出应用". |
| **Auto-Update Checking** | Users expect apps to notify them of updates, not discover outdated versions manually. | HIGH | Requires update framework (WinSparkle, Squirrel, or custom). Need server hosting update metadata and downloads. |
| **Update Notification UX** | Non-intrusive toast notification when update available. Should not interrupt workflow or steal focus. | MEDIUM | Use Windows toast notifications (Action Center). Provide clear actions: "Update Now", "Remind Me Later", "Skip This Version". |
| **Clean Application Exit** | Users expect "Exit" to completely close app and remove tray icon. No orphaned processes. | LOW | Requires proper cleanup sequence: `systray.Quit()` → `runtime.Quit()` with flags to prevent close interception. |
| **Settings/Data Persistence** | Users expect their API keys, projects, and preferences to be saved across sessions. | LOW | Use SQLite + GORM (already implemented). Store in user home directory: `~/.ai-commit-hub/`. |
| **Window State Memory** | Users expect app to remember window position and size between sessions. | LOW | Save window bounds (x, y, width, height) and maximized state to config. Restore on startup. |

---

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Double-Click Tray Icon to Show Window** | Faster than right-click → "显示窗口". Intuitive for power users. Reduces friction when frequently accessing app. | LOW | Requires upgraded systray library (`lutischan-ferenc/systray` v1.3.0+) supporting `SetOnDClick()`. Already implemented in project. |
| **Silent Auto-Update with In-App Notification** | Updates download in background without interrupting work. User sees non-intrusive badge/toast when update ready to install. Matches Chrome/VSCode behavior. | HIGH | Requires delta updates (Squirrel), background download scheduler, and installer that supports silent mode (`/SILENT` or `/VERYSILENT`). |
| **Quick Generate from Clipboard** | Watch clipboard for git diff or commit text, auto-generate message. One-click workflow for users copying from terminal. | MEDIUM | Requires clipboard monitoring, debounce to avoid excessive polling, and UI indicator showing "listening" state. |
| **Multi-Language Commit Messages** | Generate commit messages in user's preferred language (Chinese, English, etc.) without manual prompt engineering. | LOW | Add language selector in settings. Pass as system prompt to AI provider. Differentiator vs generic tools. |
| **Commit Message Style Templates** | Preset styles: Conventional Commits, emoji-based, terse, verbose. Users pick project conventions from dropdown. | MEDIUM | Store templates in config. Add template selection UI. AI prompt includes style guidelines. |
| **Project-Specific AI Provider** | Different projects can use different AI models (e.g., expensive GPT-4 for work projects, free Ollama for personal repos). | MEDIUM | Extend config schema to support per-project AI settings. UI dropdown for project settings. |
| **Keyboard Shortcuts for Power Users** | Global hotkey (e.g., `Ctrl+Alt+G`) to generate commit message without leaving IDE. System-wide availability. | HIGH | Requires global hotkey registration (Windows API `RegisterHotKey`). Background process or always-running tray app. |
| **Commit History Analytics** | Visual dashboard of commit activity, message patterns, AI usage statistics. Helps teams understand workflows. | HIGH | Requires analytics queries, charting library (frontend), and data aggregation. Nice-to-have for v2+. |

---

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Multiple Document Interface (MDI)** | "I want to manage multiple Git repos in tabs within one window." | Increases UI complexity, harder to implement tray minimize, violates Windows single-window conventions. Wails v2 MDI is brittle. | **Single-window with project dropdown**. Switch projects via dropdown or sidebar. Simpler UX, easier state management. |
| **Real-Time Git Status Watching** | "Show me repo status instantly without clicking refresh." | Continuous file system polling drains battery, high CPU usage, triggers antivirus scans on `git status`. | **Manual refresh button + auto-refresh on window activate**. Check status when user switches to app, not continuously in background. |
| **In-App Terminal/Console** | "I want to run git commands directly from the app." | Reinventing the wheel. Terminal emulators are complex (ANSI escape codes, color handling, PTY allocation). Maintenance burden. | **System terminal integration**. Add "Open in Terminal" button that launches Windows Terminal/Git Bash in project directory. |
| **Full Git GUI (Merge, Branch, Rebase)** | "Why can't I do everything git-related in this app?" | Scope creep. Feature-rich Git GUIs (SourceTree, GitKraken) exist. Building a full Git GUI distracts from core value (AI commit messages). | **Focus on commit message generation**. Integrate with existing Git GUIs via "Open in [Git GUI]" button. Be a helper tool, not a replacement. |
| **Browser-Based UI (Electron)** | "I want a modern web UI with CSS animations." | Electron apps are 100MB+ bloat, memory-hungry (300MB+ RAM), slow startup. Wails provides native performance with web tech. | **Wails (current stack)**. Native Windows performance, 10MB binary, 50MB RAM. Web frontend (Vue3) with native backend. |
| **Forced Silent Updates (No User Choice)** | "Updates should just happen automatically like Chrome." | Breaks workflows if update has bugs. Users lose control. Enterprise admins need to hold updates for testing. | **Optional silent updates**. Let users choose "Auto-download updates" vs "Notify me only". Respect user agency. |
| **Splash Screen on Startup** | "I want a branded loading screen while the app starts." | Adds 1-2 seconds delay for minimal branding value. Users just want the app to open instantly. | **Minimal loading indicator**. Show spinner in center of main window if startup takes >500ms. No dedicated splash screen. |
| **System Tray Icon Animation** | "Show an animated icon when generating commit messages." | Windows tray doesn't support GIF/video animation. Would require periodic icon replacement, flickering issues. | **Tooltip status text**. Update tray tooltip text ("Generating...", "Done"). Static icon with changing text. |

---

## Feature Dependencies

```
[Single-Instance Enforcement]
    ├──requires──> [Mutex Detection]
    ├──requires──> [Named Pipes IPC]
    └──enhances──> [Tray Icon] (activates existing window)

[System Tray]
    ├──requires──> [Icon Resources] (ICO for Windows)
    ├──requires──> [Close Event Interception] (onBeforeClose)
    └──requires──> [Exit Flow Logic] (quitting flag)

[Auto-Update]
    ├──requires──> [Update Framework] (WinSparkle/Squirrel)
    ├──requires──> [Server Hosting] (releases + appcast.xml)
    ├──requires──> [Code Signing Certificate] ($200-500/year)
    └──enhances──> [User Experience] (always up-to-date)

[Silent Auto-Update]
    ├──requires──> [Auto-Update]
    ├──requires──> [Delta Updates] (Squirrel)
    └──requires──> [Silent Installer Support] (/SILENT flag)

[Keyboard Shortcuts]
    ├──requires──> [Global Hotkey Registration] (Windows API)
    ├──requires──> [Background Process] (tray app always running)
    └──conflicts──> [Single-Instance] (need IPC to send hotkey to main instance)
```

### Dependency Notes

- **[Single-Instance] requires [Mutex Detection]**: Mutex is the standard Windows pattern for detecting if another instance is already running. Named Pipes enable the second instance to pass its arguments to the first instance before exiting.

- **[System Tray] requires [Icon Resources]**: Windows requires multi-size ICO format (16x16, 32x32, 48x48, 256x256) for crisp rendering at all DPI settings. Project already implements multi-level fallback strategy (Wails ICO → PNG-generated ICO → red placeholder).

- **[System Tray] requires [Exit Flow Logic]**: Must prevent `onBeforeClose` from intercepting exit when user clicks "退出应用". Solution: `quitting` flag (`atomic.Bool`) set before calling `systray.Quit()`. Project already implements this pattern.

- **[Auto-Update] requires [Code Signing Certificate]**: Without code signing, Windows SmartScreen shows scary warning ("Unrecognized app"). Certificate costs $200-500/year from DigiCert, Sectigo, etc. Essential for user trust.

- **[Silent Auto-Update] conflicts with [User Control]**: Some users prefer to review updates before installing. Provide settings option: "Auto-download and install updates" (default: off for enterprise, on for home users).

- **[Keyboard Shortcuts] conflicts with [Single-Instance]**: Global hotkey triggered in second instance must IPC to first instance. Requires full Named Pipes implementation, not just Mutex detection.

---

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] **Single-Instance Enforcement** — Prevents duplicate instances, activates existing window. Uses Mutex + basic window activation.
- [ ] **System Tray with Right-Click Menu** — Minimize to tray on close, "显示窗口" and "退出应用" menu items. Uses `getlantern/systray` (current project version).
- [ ] **Basic Auto-Update Checking** — Check for updates on startup, show dialog if new version available. Manual download and install (user runs installer). Simple HTTP endpoint for version check.
- [ ] **Settings Persistence** — Save API keys, projects, preferences to SQLite DB. Already implemented in project.
- [ ] **Window State Memory** — Remember window position and size between sessions.
- [ ] **Clean Application Exit** — Proper cleanup, no orphaned processes. Already implemented in project with `quitting` flag pattern.

**Rationale**: These features establish baseline Windows desktop app behavior. Users will perceive the app as "broken" without single-instance or tray minimize. Auto-update checking (even manual) is expected for networked apps.

---

### Add After Validation (v1.x)

Features to add once core is working and users validate value proposition.

- [ ] **Double-Click Tray Icon to Show Window** — Fast access for power users. Requires upgrading to `lutischan-ferenc/systray` v1.3.0+ for `SetOnDClick()` support.
- [ ] **Toast Notification for Updates** — Non-intrusive Windows Action Center notifications instead of modal dialog.
- [ ] **Project-Specific AI Provider** — Allow different AI models per project. Extends config schema and UI.
- [ ] **Commit Message Style Templates** — Conventional Commits, emoji-based, etc. Template system in backend.
- [ ] **Quick Generate from Clipboard** — Monitor clipboard for git diff, one-click generation. Improves workflow efficiency.
- [ ] **One-Click Installer Update** — Auto-download and prompt to install update. Still requires user approval, but no manual download.

**Trigger**: When users report "I use this daily, but I wish I could..." or quantitative usage data (e.g., >50% DAU, >10 projects per user).

---

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Silent Auto-Update** — Background download + install with in-app notification. Requires Squirrel framework and code signing certificate. High complexity, defer until >1000 users.
- [ ] **Keyboard Shortcuts for Power Users** — Global hotkey to generate commit message. Requires global hotkey registration and IPC. Wait for explicit user demand.
- [ ] **Commit History Analytics** — Visual dashboard of commit patterns. Nice-to-have analytics, not core value. Build after core features are stable.
- [ ] **Multi-Language Commit Messages** — Generate in Chinese, English, etc. Wait for international user base (>25% non-English).
- [ ] **Delta Updates** — Differential downloads to reduce bandwidth. Squirrel supports this. Implement if updates are >50MB or user base is on metered connections.

**Why Defer**: These features require significant engineering effort but don't directly validate the core hypothesis (AI commit message generation is useful). Focus on simplicity and speed to launch.

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Single-Instance Enforcement | HIGH | MEDIUM | **P1** |
| System Tray (Right-Click Menu) | HIGH | LOW | **P1** |
| Settings Persistence | HIGH | LOW (✅ already done) | **P1** |
| Clean Application Exit | HIGH | LOW (✅ already done) | **P1** |
| Window State Memory | MEDIUM | LOW | **P1** |
| Basic Auto-Update Checking | MEDIUM | MEDIUM | **P1** |
| Double-Click Tray Icon | MEDIUM | LOW (library upgrade) | **P2** |
| Toast Notification for Updates | MEDIUM | MEDIUM | **P2** |
| Commit Message Style Templates | MEDIUM | MEDIUM | **P2** |
| Project-Specific AI Provider | LOW | MEDIUM | **P2** |
| Quick Generate from Clipboard | MEDIUM | HIGH | **P3** |
| One-Click Installer Update | MEDIUM | HIGH | **P3** |
| Silent Auto-Update | HIGH | HIGH | **P3** |
| Keyboard Shortcuts for Power Users | LOW | HIGH | **P3** |
| Commit History Analytics | LOW | HIGH | **P3** |
| Multi-Language Commit Messages | LOW | MEDIUM | **P3** |
| Delta Updates | LOW | HIGH | **P3** |

**Priority key:**
- **P1: Must have for launch** — Table stakes features. Users expect these; missing = product feels incomplete.
- **P2: Should have, add when possible** — Differentiators that enhance UX. Add after MVP validation.
- **P3: Nice to have, future consideration** — Features with high complexity or uncertain value. Defer until v2+.

---

## Competitor Feature Analysis

| Feature | GitGUI (SourceTree, GitKraken) | Commit.ai (AI commit tools) | Our Approach |
|---------|-------------------------------|----------------------------|--------------|
| **Single-Instance** | ✅ Yes (standard) | ⚠️ Varies (web apps N/A) | ✅ **Implement** — Mutex + Named Pipes |
| **System Tray** | ✅ Yes | ❌ No (browser-based) | ✅ **Implement** — Minimize to tray with menu |
| **Tray Double-Click** | ✅ Yes (some) | N/A | ✅ **Implement** — Faster access |
| **Auto-Update** | ✅ Yes (built-in) | N/A | ⚠️ **Simple version** — Manual download |
| **Silent Update** | ✅ Yes (Chrome-like) | N/A | ❌ **Defer** — v2+ consideration |
| **Full Git GUI** | ✅ Yes (core feature) | ❌ No | ❌ **Anti-feature** — Focus on commit messages only |
| **In-App Terminal** | ✅ Yes (some) | ❌ No | ❌ **Anti-feature** — Use "Open in Terminal" button |
| **AI Provider Selection** | N/A | ⚠️ Varies (often hardcoded) | ✅ **Implement** — Multi-provider support (already done) |
| **Commit Style Templates** | ❌ No | ⚠️ Varies | ✅ **Implement** — Differentiator |
| **Clipboard Monitoring** | ❌ No | ⚠️ Varies | ⚠️ **Consider** — Nice workflow improvement |
| **Analytics Dashboard** | ✅ Yes (enterprise) | ❌ No | ❌ **Defer** — Not core value |

**Key Differentiators:**
1. **Lightweight vs Heavy Git GUI** — We focus on commit message generation, not full Git operations. This is our strength.
2. **Multi-Provider AI Support** — Competitors often lock into one AI model. We support OpenAI, Anthropic, DeepSeek, Ollama, etc. Already implemented.
3. **Tray Double-Click** — Simple UX improvement that many competitors overlook. Fast access for power users.
4. **Commit Style Templates** — Help teams follow conventions without manual prompt engineering. Unique differentiator.

---

## Windows UX Conventions

Based on research and official Microsoft documentation, these are the expected behaviors for Windows desktop apps:

### Single-Instance Behavior

**Standard Convention:**
- When user launches an already-running app, the existing window should be activated and brought to the foreground.
- No error dialog ("Application is already running"). This interrupts user workflow.
- If second instance has arguments (e.g., file to open), pass them to the first instance via IPC.

**Implementation Pattern:**
```
1. Try to create Mutex with unique name (e.g., "Global\com.allanpk716.ai-commit-hub")
2. If creation fails:
   - Another instance is running
   - Open Named Pipe to first instance
   - Send arguments (if any)
   - Exit second instance
3. If creation succeeds:
   - This is the first instance
   - Create Named Pipe server to listen for arguments from subsequent instances
   - Show main window
```

**Sources:**
- [Single Instance WinForm App in C# with Mutex and Named Pipes](https://www.autoitconsulting.com/site/development/single-instance-winform-app-csharp-mutex-named-pipes/) (HIGH confidence)
- [Single-Instance .NET Apps: Mutexes, Named Pipes, UX](https://www.dotnet-guide.com/how-to-restrict-a-program-to-single-instance-in-net.html) (HIGH confidence)
- [App instancing with the app lifecycle API - Microsoft Learn](https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/applifecycle/applifecycle-instancing) (HIGH confidence)

---

### System Tray Interactions

**Standard Convention:**
- **Right-click**: Show context menu with standard options (Show, Exit).
- **Double-click**: Open main window (optional but expected by power users).
- **Single-click**: No standard behavior. Some apps show window, others show menu, others do nothing.
- **Tooltip**: Show app name and brief status text.

**Implementation Pattern:**
```
1. Intercept window close event (onBeforeClose)
   - Set flag: isUserQuitting = false
   - Hide window: runtime.WindowHide(ctx)
   - Return true (prevent close)
2. Tray menu items:
   - "显示窗口" (Show Window) → Call showWindow()
   - Separator
   - "退出应用" (Exit Application) → Set isUserQuitting = true, call systray.Quit()
3. Double-click (optional):
   - systray.SetOnDClick(func() { showWindow() })
4. Exit flow:
   - systray.Quit() → onSystrayExit() → Check isUserQuitting → If true, runtime.Quit()
```

**Current Project Status:**
- ✅ Tray implemented with `getlantern/systray` v1.2.2
- ✅ Right-click menu working
- ⚠️ Double-click not supported (requires `lutischan-ferenc/systray` v1.3.0+)
- ✅ Exit flow working with `quitting` flag pattern

**Sources:**
- [Notification Area - Win32 apps - Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/uxguide/winenv-notification) (HIGH confidence)
- [Those notification icons, with their clicks, double-clicks, right-clicks... - Microsoft Old New Thing](https://devblogs.microsoft.com/oldnewthing/20090430-00/?p=18393) (HIGH confidence)
- [Correct behaviour for tray icon click - Stack Overflow](https://stackoverflow.com/questions/1050477/correct-behaviour-for-tray-icon-click) (MEDIUM confidence)

---

### Auto-Update UX Patterns

**Standard Convention:**
- **Check on startup**: App should check for updates when launched (not continuously in background).
- **Notification style**: Non-intrusive toast notification (Windows Action Center), not modal dialog.
- **User actions**: "Update Now", "Remind Me Later" (dismiss), "Skip This Version" (ignore this specific version).
- **Download behavior**: Download in background, notify when ready to install.
- **Install behavior**: Prompt user to close app and install, or install on next restart.

**Implementation Options:**

| Framework | Silent Update | Delta Updates | Complexity | Maturity |
|-----------|--------------|---------------|------------|----------|
| **WinSparkle** | ⚠️ Partial (installer must support) | ❌ No | LOW | HIGH (ported from macOS Sparkle) |
| **Squirrel** | ✅ Yes | ✅ Yes | HIGH | HIGH (used by Slack, VSCode, Discord) |
| **Google Omaha** | ✅ Yes | ✅ Yes | VERY HIGH | HIGH (Chrome's updater) |
| **Custom HTTP + simple version check** | ❌ No | ❌ No | LOW | MEDIUM (DIY) |

**Recommended for MVP:**
- **Custom HTTP + simple version check** (LOW complexity)
  - GitHub Releases hosts binaries and `version.json` metadata
  - App checks `https://releases.allanpk716.com/ai-commit-hub/version.json` on startup
  - If new version available, show modal dialog with release notes
  - User clicks "Download" → opens browser to releases page
  - User downloads and runs installer manually

**Recommended for v1.x:**
- **WinSparkle** (LOW-MEDIUM complexity)
  - Handles update checks, downloads, and installer launching
  - XML appcast format (compatible with Sparkle framework)
  - Shows native UI dialogs
  - Silent updates require installer support (InnoSetup `/SILENT` flag)

**Recommended for v2+:**
- **Squirrel.Windows** (HIGH complexity)
  - Delta updates (only download changed files)
  - Silent background updates
  - Auto-restart after update
  - Requires significant refactoring (app must be installed per-user, not portable)

**Sources:**
- [The best update frameworks for Windows - Omaha Consulting](https://omaha-consulting.com/best-update-framework-for-windows) (MEDIUM confidence)
- [WinSparkle - GitHub](https://github.com/vslavik/winsparkle) (HIGH confidence)
- [Support completely silent installation · Issue #21 - WinSparkle](https://github.com/vslavik/winsparkle/issues/21) (MEDIUM confidence)
- [Automatic updates to your Windows desktop application - /dev/solita](https://dev.solita.fi/2016/03/14/automatic-updater-for-windows-desktop-app.html) (MEDIUM confidence)

---

## Sources

### Single-Instance Research
- [Single Instance WinForm App in C# with Mutex and Named Pipes - AutoIt Consulting](https://www.autoitconsulting.com/site/development/single-instance-winform-app-csharp-mutex-named-pipes/) — HIGH confidence, 2018
- [Single-Instance .NET Apps: Mutexes, Named Pipes, UX - dotnet-guide.com](https://www.dotnet-guide.com/how-to-restrict-a-program-to-single-instance-in-net.html) — HIGH confidence, 2025
- [App instancing with the app lifecycle API - Microsoft Learn](https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/applifecycle/applifecycle-instancing) — HIGH confidence, 2025
- [Single instance application: how to open existing window? - AvaloniaUI Discussion #17854](https://github.com/AvaloniaUI/Avalonia/discussions/17854) — MEDIUM confidence, 2024
- [Single Instance Application for .NET 6 or 7 - medo64.com](https://medo64.com/posts/single-instance-application-for-net-6-or-7/) — MEDIUM confidence, 2022

### System Tray Research
- [Notification Area - Win32 apps - Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/uxguide/winenv-notification) — HIGH confidence (official Microsoft UX guidelines)
- [Those notification icons, with their clicks, double-clicks, right-clicks... - Microsoft Old New Thing Blog](https://devblogs.microsoft.com/oldnewthing/20090430-00/?p=18393) — HIGH confidence, 2009
- [System Tray Icon Double / Single Click Issue - Stack Overflow](https://stackoverflow.com/questions/23457600/system-tray-icon-double-single-click-issue) — MEDIUM confidence
- [Enable Double-clicking of Tray Icon - eM Client Forum](https://forum.emclient.com/t/enable-double-clicking-of-tray-icon/36500) — LOW confidence (user forum)
- Project Documentation: `docs/lessons-learned/windows-tray-icon-implementation-guide.md` — HIGH confidence (project-specific, tested)

### Auto-Update Research
- [The best update frameworks for Windows - Omaha Consulting](https://omaha-consulting.com/best-update-framework-for-windows) — MEDIUM confidence, 2025
- [WinSparkle GitHub Repository](https://github.com/vslavik/winsparkle) — HIGH confidence (official documentation)
- [Support completely silent installation · Issue #21 - WinSparkle](https://github.com/vslavik/winsparkle/issues/21) — MEDIUM confidence (2014, still open)
- [WinSparkle silent automatic update - Stack Overflow](https://stackoverflow.com/questions/32716678/winsparkle-silent-automatic-update) — MEDIUM confidence, 2015
- [Automatic update a Windows application - Stack Overflow](https://stackoverflow.com/questions/4769615/automatic-update-a-windows-application) — MEDIUM confidence, 2011
- [Automatic updates to your Windows desktop application - /dev/solita](https://dev.solita.fi/2016/03/14/automatic-updater-for-windows-desktop-app.html) — MEDIUM confidence, 2016 (Squirrel guide)

### Windows Update Notification UX
- [What is the best way to notify a user that updates are available in a desktop app - UX Stack Exchange](https://ux.stackexchange.com/questions/129949/what-is-the-best-way-to-notify-a-user-that-updates-are-available-in-a-desktop-ap) — MEDIUM confidence
- [5 Great Ways To Communicate Product Updates To Your Users - UX Studio Team](https://www.uxstudioteam.com/ux-blog/communicating-product-updates) — MEDIUM confidence
- [Update Notifications on Windows Desktop - Joplin App Discussion](https://discourse.joplinapp.org/t/update-notifications-on-windows-desktop/45828) — LOW confidence (user feedback)

### Project Documentation
- `docs/development/wails-development-standards.md` — Project Wails development standards
- `docs/lessons-learned/windows-tray-icon-implementation-guide.md` — Comprehensive Windows tray implementation guide
- `docs/fixes/tray-icon-doubleclick-fix.md` — Double-click feature implementation
- `app.go` — Current implementation with tray, single-instance patterns (not yet fully implemented)

---

## Quality Gate Checklist

- [x] Categories are clear (table stakes vs differentiators vs anti-features)
- [x] Complexity noted for each feature (LOW/MEDIUM/HIGH)
- [x] Dependencies between features identified (dependency graph)
- [x] Windows UX conventions covered (single-instance, tray, update UX)

---

**Feature research for:** AI Commit Hub (Windows Desktop Application)
**Researched by:** Claude (GSD Project Researcher)
**Date:** 2026-02-06
**Next Steps:** Use this research to define roadmap phases and feature priorities in roadmap creation.
