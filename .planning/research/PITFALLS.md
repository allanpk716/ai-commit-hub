# Pitfalls Research

**Domain:** Wails Windows Desktop Application Development
**Researched:** 2026-02-06
**Confidence:** MEDIUM

## Critical Pitfalls

### Pitfall 1: Single-Instance Deadlock with Wails

**What goes wrong:**
Application hangs or becomes unresponsive when implementing single-instance locking. Multiple processes can be created, or the second instance cannot communicate with the first instance.

**Why it happens:**
- Using `sync.Mutex` or other Go synchronization primitives across process boundaries (doesn't work)
- Implementing custom single-instance logic that conflicts with Wails' internal initialization
- Not using Wails' built-in `SingleInstanceLock` option correctly
- File-based locks that can become stale if the application crashes

**How to avoid:**
- Use Wails' built-in `SingleInstanceLock` option in `wails.json`:
  ```json
  {
    "author": "Your Name",
    "name": "App Name",
    "singleInstanceLock": true
  }
  ```
- If using Windows Mutex directly, ensure proper error handling and cleanup
- Use `LockOSThread()` for Windows message loops if implementing custom solution
- Always implement lock release in `defer` statements

**Warning signs:**
- Second instance opens successfully (should not happen)
- First instance becomes unresponsive when second instance starts
- Mutex/lock files remain after application crashes
- Deadlock detected with `go run -race`

**Phase to address:**
**Phase: Infrastructure Setup** - Must be configured before beta testing

---

### Pitfall 2: Auto-Update Breaking Running Application

**What goes wrong:**
Update process fails, corrupts the application, or leaves the system in an unusable state. Common failures include:
- Application cannot replace its own executable while running
- Update process crashes halfway through
- New version fails to start, leaving user with no working application
- File locking issues prevent file replacement

**Why it happens:**
- Trying to replace files that are in use (Windows locks executable files)
- Not waiting for the main process to exit before starting update
- No rollback mechanism if update fails
- Insufficient disk space checks before download
- Update process lacks necessary permissions
- Not handling Windows UAC prompts correctly

**How to avoid:**
- Use a separate updater process (`cmd/updater/main.go`) that waits for main process to exit
- Implement backup mechanism before replacing files (`.bak` files)
- Check disk space before downloading updates
- Use `MoveFileEx` with `MOVEFILE_DELAY_UNTIL_REBOOT` as fallback for locked files
- Test rollback scenarios explicitly
- Verify update integrity (SHA256 checksums)
- Implement startup validation that can rollback to backup

**Warning signs:**
- "Access denied" errors during file replacement
- Application shows "file in use" errors
- Disk full errors during download
- New version crashes on startup
- Temporary files left in update directories

**Phase to address:**
**Phase: Auto-Update Implementation** - Critical for production stability

---

### Pitfall 3: System Tray Event Handling Race Conditions

**What goes wrong:**
Tray icon clicks are lost, double-clicks don't work, or application crashes when interacting with system tray. Event handlers fire multiple times or not at all.

**Why it happens:**
- Multiple goroutines accessing tray state without synchronization
- Calling `systray.Quit()` from within event handlers without proper synchronization
- Race conditions between window show/hide and tray interactions
- Not using `sync.Once` for one-time initialization
- Timing issues - systray initialized before Wails fully ready

**How to avoid:**
- Use `sync.RWMutex` to protect window visibility state
- Use `sync.Once` for one-time operations (like quit logic)
- Use `atomic.Bool` flags for state changes (like `quitting` flag in exit flow)
- Add delays between systray and Wails initialization:
  ```go
  go func() {
      time.Sleep(300 * time.Millisecond)  // Wait for Wails init
      a.runSystray()
  }()
  ```
- Use `LockOSThread()` in systray goroutine on Windows
- Avoid calling `runtime.Quit()` and `systray.Quit()` in sequence without coordination

**Warning signs:**
- Tray menu clicks have no effect
- Application crashes when clicking tray icon
- Double-click events don't fire or fire erratically
- "Quit" option doesn't work
- Race detector shows issues with `go run -race`

**Phase to address:**
**Phase: System Tray Implementation** - Must be tested before beta

---

### Pitfall 4: GitHub Actions Build Failures for Wails Apps

**What goes wrong:**
CI/CD builds fail intermittently or produce non-functional executables. Common issues:
- "Not a valid Win32 application" error
- WebView2 runtime missing in CI environment
- Frontend build artifacts not included
- Version injection not working
- Build cache corruption

**Why it happens:**
- Go version incompatibilities (e.g., Go 1.25.x with Wails 2.10.2)
- Corrupted `wailsbindings` executables in build cache
- Missing or incorrect `ldflags` for version injection
- Platform-specific dependencies not installed
- Node.js version mismatches for frontend build
- Insufficient build timeouts

**How to avoid:**
- Pin Go version in GitHub Actions (use `.go-version` file):
  ```yaml
  - uses: actions/setup-go@v4
    with:
      go-version: '1.21'  # Don't use 1.25.x with Wails 2.x
  ```
- Clear Wails build cache in CI:
  ```bash
  wails build -clean
  ```
- Install WebView2 runtime explicitly on Windows runners
- Use specific Node.js version in frontend build
- Verify build outputs with basic smoke tests
- Build both debug and release variants in CI
- Cache Go modules and node_modules separately

**Warning signs:**
- Build succeeds locally but fails in CI
- "wailsbindings.exe: %1 is not a valid Win32 application"
- Intermittent build failures
- Binary crashes on startup on CI machines
- Version returns "dev" instead of injected version

**Phase to address:**
**Phase: CI/CD Setup** - Must validate before tagging releases

---

### Pitfall 5: Windows-Specific Issues (UAC, File Locking, Permissions)

**What goes wrong:**
- Application requires admin privileges unexpectedly (UAC prompts)
- Cannot write to user directories (Program Files, etc.)
- File operations fail silently
- Application fails to start on Windows 7 (missing WebView2)
- August 2025 Windows Update causes unexpected UAC behavior

**Why it happens:**
- Not embedding application manifest
- Writing to install directory instead of user profile
- Assuming admin privileges
- Not checking WebView2 runtime availability
- Not handling Windows 11 24H2 UAC changes

**How to avoid:**
- Write application data to user profile (`%USERPROFILE%\.ai-commit-hub`)
- Never write to Program Files or Windows directories
- Embed Windows manifest requesting `asInvoker` (no elevation needed):
  ```xml
  <requestedExecutionLevel level="asInvoker" uiAccess="false" />
  ```
- Check WebView2 runtime availability at startup, provide download link if missing
- Test on Windows 10/11 with different UAC settings
- Support portable deployment (no installer required)
- Handle Windows August 2025 UAC bug - unexpected prompts may appear

**Warning signs:**
- UAC prompt appears on every launch
- Configuration/settings cannot be saved
- "Access denied" errors in logs
- Application won't start on Windows 7
- Application crashes after Windows updates

**Phase to address:**
**Phase: Windows Compatibility** - Test across Windows versions

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Bypassing single-instance lock during development | Faster testing | Multi-instance bugs in production | Never - causes data corruption |
| Using file-based locks for single instance | Simple to implement | Stale locks after crashes | Only for early prototypes |
| Hard-coding update URLs | No infrastructure needed | No rollback, breaking changes | Never - security risk |
| Skipping update verification | Faster iteration | Corrupted installations | Never - unacceptable |
| Systray init without synchronization | Fewer lines of code | Race conditions, crashes | Never - see above |
| Writing to install directory | Simpler deployment | Permission issues, UAC | Never - Windows pitfall |
| Using `runtime.Quit()` without coordination | Simple exit logic | Systray deadlocks | Never - see Pitfall 3 |

---

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| GitHub Releases API | No authentication, hitting rate limits | Use conditional auth, handle 403 gracefully |
| WebView2 (Windows) | Assuming it's always installed | Check at startup, provide installer |
| System Tray | Initializing before Wails ready | Delay 300ms, use separate goroutine |
| File System (Windows) | Writing to Program Files | Write to `%USERPROFILE%` or `os.UserConfigDir()` |
| Mutex (single instance) | Using Go sync.Mutex across processes | Use Wails SingleInstanceLock or Windows Mutex |
| Auto-updater | Replacing running executable | Separate updater process with PID wait |
| GitHub Actions | Not pinning Go/Node versions | Pin versions in workflow, test locally first |

---

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| No update caching | API rate limiting, slow startup | Cache release info for 15 min | At 100+ daily active users |
| Synchronous update check | UI freezes on startup | Check in goroutine, emit event | Immediately |
| No download resume | Failed updates on slow connections | Support HTTP Range requests | On unstable networks |
| Loading all projects at once | Slow startup, high memory | Lazy load, pagination | At 50+ projects |
| No thumbnail caching | Slow git history rendering | Cache status locally | After 1000+ commits |

---

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Downloading updates over HTTP | Man-in-the-middle attacks | Always use HTTPS for GitHub API |
| No signature verification | Malicious updates | Verify SHA256 checksums |
| Running updater with elevated privileges | Privilege escalation | Run updater as same user as main app |
| Storing API keys in plaintext | Credential theft | Use Windows Credential Manager |
| Executing update without validation | Code injection | Verify checksums before extracting |
| Hardcoded update URLs | Supply chain attacks | Use GitHub Releases API, verify repo |

---

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Blocking UI during update check | App feels slow | Check silently in background |
| No "skip version" option | Repeated annoying prompts | Store skipped version in DB |
| Forced immediate restart | Lost work | Allow user to defer restart |
| No progress indication | Uncertainty during download | Show download progress bar |
| Silent auto-install | Unexpected interruptions | Prompt before installing |
| No rollback on failure | Complete app breakage | Auto-rollback to backup |
| Admin requirement prompts | Reduced adoption | Design for standard user (no UAC) |

---

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Single Instance Lock**: Often missing cross-session testing — verify works across RDP sessions and fast user switching
- [ ] **Auto-Update**: Often missing rollback testing — verify app still works if update fails mid-way
- [ ] **System Tray**: Often missing double-click testing — verify double-click works on high DPI displays
- [ ] **GitHub Actions**: Often missing artifact testing — verify downloaded zip actually runs
- [ ] **UAC Handling**: Often missing non-admin testing — verify works without admin privileges
- [ ] **WebView2 Detection**: Often missing Windows 7 testing — verify graceful failure without WebView2
- [ ] **File Locking**: Often missing in-use testing — verify update works when app is running
- [ ] **Version Injection**: Often missing release build testing — verify VERSION is not "dev" in release

---

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Single-instance deadlock | HIGH | Kill all instances via Task Manager, delete lock file, add `LockOSThread()` |
| Corrupted update | MEDIUM | Restore from `.bak` backup, implement fallback installer download |
| Systray race condition | LOW | Add `sync.RWMutex`, `atomic.Bool` flags, use `go run -race` to detect |
| CI build failure | LOW | Clear `wailsbindings` cache, rebuild with `-clean` flag |
| UAC permission issues | MEDIUM | Reinstall to user directory, add application manifest |
| Missing WebView2 | MEDIUM | Provide WebView2 installer link, add detection at startup |

---

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Single-instance deadlock | Infrastructure Setup | Test launching 2+ instances, verify only 1 runs |
| Auto-update breaking app | Auto-Update Implementation | Test failed update, verify rollback works |
| Systray race conditions | System Tray Implementation | Run with `go run -race`, test rapid clicks |
| CI/CD build failures | CI/CD Setup | Build on clean machine, verify binary runs |
| Windows UAC/permissions | Windows Compatibility | Test as non-admin, verify no UAC prompts |
| File locking issues | Auto-Update Implementation | Test update while app running |
| Performance issues | Load Testing | Test with 100+ projects, 1000+ commits |
| Security vulnerabilities | Security Review | Run security audit, verify HTTPS usage |

---

## Sources

### Official Documentation
- [Wails Single Instance Lock Guide](https://wails.io/docs/guides/single-instance-lock/) - HIGH confidence (official docs)
- [Wails Troubleshooting](https://wails.io/docs/guides/troubleshooting) - HIGH confidence (official docs)
- [Wails Options Reference](https://wails.io/docs/reference/options/) - HIGH confidence (official docs)
- [GitHub Actions Cross-platform Build](https://wails.golang.ac.cn/docs/guides/crossplatform-build) - HIGH confidence (official docs)

### GitHub Issues & Discussions
- [Wails Issue #1351 - Single Process Option](https://github.com/wailsapp/wails/issues/1351) - MEDIUM confidence (community discussion)
- [Wails Issue #1178 - Self-Updating Support](https://github.com/wailsapp/wails/issues/1178) - MEDIUM confidence (feature request)
- [Wails Issue #950 - EventManager Race Condition](https://github.com/wailsapp/wails/issues/950) - MEDIUM confidence (race condition report)
- [Wails Issue #4551 - Go 1.25.0 Incompatibility](https://github.com/wailsapp/wails/issues/4551) - HIGH confidence (version compatibility issue)

### Community Resources
- [Wails Auto-Update Guide (Dev.J, June 2024)](https://blog.stackademic.com/do-you-use-wails-and-need-automatic-updates-5fdba1485692) - MEDIUM confidence (community implementation)
- [Wails CI/CD Guide (CSDN, Dec 2025)](https://blog.csdn.net/gitblog_01191/article/details/154858418) - LOW confidence (unverified)
- [Systray Click Events Discussion](https://github.com/getlantern/systray/issues/30) - MEDIUM confidence (feature discussion)

### Windows-Specific Issues
- [Windows UAC Bug After August 2025 Update (4sysops)](https://4sysops.com/archives/windows-uac-bug-after-the-august-2025-update/) - HIGH confidence (verified report)
- [Microsoft Support - Unexpected UAC Prompts](https://support.microsoft.com/en-us/topic/unexpected-uac-prompts-when-running-msi-repair-operations-after-installing-the-august-2025-windows-security-update-5806f583-e073-4675-9464-fe01974df273) - HIGH confidence (official Microsoft)
- [Windows Internals - Thread Management](https://medium.com/windows-os-internals/windows-internals-thread-management-part-2-75cfed18f9ca) - MEDIUM confidence (technical reference)

### Go Concurrency
- [Race Conditions in Go and Detection (Medium, Jan 2025)](https://medium.com/@debug-ing/race-conditions-in-go-and-race-detection-07b4a46bf1a9) - MEDIUM confidence (technical guide)
- [Google Project Zero - Windows Race Conditions](https://projectzero.google/2025/12/windows-exploitation-techniques.html) - HIGH confidence (security research)

### Project Experience
- `docs/lessons-learned/windows-tray-icon-implementation-guide.md` - HIGH confidence (project-specific)
- `docs/fixes/tray-icon-doubleclick-fix.md` - HIGH confidence (project-specific)
- `docs/fixes/systray-exit-fix.md` - HIGH confidence (project-specific)
- `docs/development/wails-development-standards.md` - HIGH confidence (project-specific)

---

*Pitfalls research for: AI Commit Hub (Wails Windows Desktop Application)*
*Researched: 2026-02-06*
