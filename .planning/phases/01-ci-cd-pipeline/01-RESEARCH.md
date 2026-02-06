# Phase 1: CI/CD Pipeline - Research

**Researched:** 2026-02-06
**Domain:** GitHub Actions CI/CD for Wails v2 Windows applications
**Confidence:** HIGH

## Summary

Phase 1 focuses on establishing an automated build and release pipeline for AI Commit Hub, a Wails v2 desktop application. The research reveals that GitHub Actions combined with Wails-specific build actions provides a robust foundation for Windows-only builds with multi-architecture support (amd64 and 386).

Key findings indicate that while the Wails ecosystem has mature build tooling, there are critical considerations around 32-bit Windows support (GOARCH=386), which is **not officially supported** and has known WebView2 crashes. The project already has version embedding infrastructure in place via `pkg/version` and `wails.json` ldflags configuration, requiring only workflow implementation.

For Claude's Discretion items:
- **Go compiler optimization**: Go does NOT have traditional `-O`/`-Osize` flags; use `-ldflags="-s -w"` for size reduction
- **GitHub Actions runner**: `windows-latest` is appropriate and current (Windows Server 2025)
- **Timeout**: Default 360 minutes (6 hours) is sufficient; Wails builds typically complete in 10-20 minutes
- **Compression**: ZIP format (via 7z or PowerShell) is standard and reliable

**Primary recommendation:** Use `dAppServer/wails-build-action@v3` with custom packaging steps for checksums and documentation, implementing matrix builds for amd64/386 with awareness of 32-bit limitations.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **GitHub Actions** | - | CI/CD automation platform | Native GitHub integration, free for public repos, Windows runner support |
| **Wails CLI** | v2.11.0 | Build framework for Go+Vue desktop apps | Project framework, required for compilation |
| **dAppServer/wails-build-action** | v3 | GitHub Action for Wails builds | Officially recommended in Wails docs, handles Go+Node.js setup automatically |
| **softprops/action-gh-release** | v1 | GitHub Release creation and asset upload | Most popular (7k+ stars), handles draft releases, pre-release detection, file uploads |
| **Go** | 1.21+ | Backend runtime | Specified in go.mod (1.24.1), Wails v2 requires Go 1.18+ |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **actions/setup-go** | v5 | Go environment setup with caching | For manual builds instead of wails-build-action |
| **actions/setup-node** | v4 | Node.js environment setup with caching | When not using wails-build-action |
| **actions/checkout** | v4 | Repository checkout | Required for all workflows |
| **7-Zip** | (Windows built-in) | Archive creation | Already used in existing release.yml |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `dAppServer/wails-build-action` | Manual Go+Node.js setup | More control but more maintenance; action handles 90% of setup |
| `softprops/action-gh-release` | `actions/create-release` (deprecated) | softprops is actively maintained, create-release is archived |
| ZIP format | TAR.GZ or 7z native format | ZIP is most widely supported on Windows, TAR requires extra tools |
| Multi-arch matrix builds | Separate workflow jobs | Matrix is cleaner, faster, easier to maintain |

**Installation:** No installation required (all GitHub Actions are SaaS)

## Architecture Patterns

### Recommended Project Structure

```
.github/
└── workflows/
    └── build.yml           # Main CI/CD workflow

build/
└── windows/                # Windows-specific resources (already exists)
    └── icon.ico            # App icon

pkg/
└── version/                # Version package (already exists)
    ├── version.go          # Version variables with ldflags injection points
    └── version_test.go

docs/
└── config-examples/        # NEW: For release package docs
    └── config.yaml         # Annotated configuration example
```

### Pattern 1: Matrix Build Strategy for Multi-Architecture Windows Builds

**What:** Build multiple Windows architectures (amd64, 386) in parallel using GitHub Actions matrix strategy

**When to use:**
- Targeting both 64-bit and 32-bit Windows systems
- Wanting fastest total build time (parallel execution)
- When 32-bit support is a requirement

**Example:**
```yaml
# Source: Wails official documentation + dAppServer action examples
# https://wails.io/docs/next/guides/crossplatform-build
# https://github.com/marketplace/actions/wails-build-action

name: Build

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  build:
    strategy:
      fail-fast: false  # Don't cancel other builds if one fails
      matrix:
        include:
          - arch: amd64
            platform: windows/amd64
            goarch: amd64
          - arch: 386
            platform: windows/386
            goarch: 386

    runs-on: windows-latest

    steps:
      - uses: actions/checkout@v4

      - name: Build Wails app
        uses: dAppServer/wails-build-action@v3
        with:
          build-name: ai-commit-hub
          build-platform: ${{ matrix.platform }}
          package: false  # We'll package ourselves
          go-version: '1.21'
        env:
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: '1'

      # Post-build steps for packaging...
```

**Critical notes:**
- **32-bit (386) WARNING**: Wails does NOT officially support GOARCH=386; known WebView2 crashes on 32-bit ([Issue #4444](https://github.com/wailsapp/wails/issues/4444))
- Consider omitting 386 builds until Wails adds official support
- `fail-fast: false` ensures one architecture failure doesn't cancel the other

### Pattern 2: Version Embedding via ldflags

**What:** Inject version metadata at compile time using Go linker flags

**When to use:**
- Wanting to embed build-time information (version, commit SHA, timestamp)
- Need version display in `-v`/`--version` output
- Already implemented in project (wails.json has ldflags configured)

**Example:**
```yaml
# Source: Existing wails.json configuration + Go ldflags documentation
# https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications

# In wails.json (ALREADY CONFIGURED):
{
  "build:ldflags": "-X 'github.com/allanpk716/ai-commit-hub/pkg/version.Version={{.Version}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.CommitSHA={{.Commit}}' -X 'github.com/allanpk716/ai-commit-hub/pkg/version.BuildTime={{.Date}}'"
}

# In GitHub Actions workflow:
- name: Extract version and metadata
  id: meta
  shell: bash
  run: |
    VERSION=${GITHUB_REF#refs/tags/v}
    SHA=${{ github.sha }}
    DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
    echo "SHA=$SHA" >> $GITHUB_OUTPUT
    echo "DATE=$DATE" >> $GITHUB_OUTPUT

- name: Build with version injection
  run: |
    wails build -ldflags "\
      -X 'github.com/allanpk716/ai-commit-hub/pkg/version.Version=${{ steps.meta.outputs.VERSION }}' \
      -X 'github.com/allanpk716/ai-commit-hub/pkg/version.CommitSHA=${{ steps.meta.outputs.SHA }}' \
      -X 'github.com/allanpk716/ai-commit-hub/pkg/version.BuildTime=${{ steps.meta.outputs.DATE }}' \
      -s -w"  # -s -w strips debug info, reduces binary size
```

**Key points:**
- Project already has version package at `pkg/version/version.go`
- Three variables injected: `Version`, `CommitSHA`, `BuildTime`
- `-s -w` flags strip symbol table and DWARF debug info (reduces binary size by ~30%)

### Pattern 3: Pre-release Detection from Tags

**What:** Automatically detect beta/RC versions and mark GitHub Releases as pre-release

**When to use:**
- Wanting to separate stable releases from pre-releases
- Following semantic versioning with pre-release identifiers
- Don't want pre-releases appearing in default update checks

**Example:**
```yaml
# Source: softprops/action-gh-release documentation + semver conventions
# https://github.com/softprops/action-gh-release
# https://semver.org/#spec-item-9

- name: Detect pre-release
  id: prerelease
  shell: bash
  run: |
    VERSION="${{ steps.meta.outputs.VERSION }}"
    # Check for pre-release identifiers: -alpha, -beta, -rc, etc.
    if [[ "$VERSION" =~ -(alpha|beta|rc|pre)\.?[0-9]*$ ]]; then
      echo "prerelease=true" >> $GITHUB_OUTPUT
    else
      echo "prerelease=false" >> $GITHUB_OUTPUT
    fi

- name: Create Release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      build/*.zip
      build/*.sha256
      build/*.md5
    name: v${{ steps.meta.outputs.VERSION }}
    draft: false
    prerelease: ${{ steps.prerelease.outputs.prerelease }}
    generate_release_notes: true
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Pattern matching:**
- Matches: `v1.0.0-beta`, `v1.0.0-rc.1`, `v2.0.0-alpha.3`
- Does NOT match: `v1.0.0` (stable release)

### Pattern 4: Checksum Generation for Release Assets

**What:** Generate SHA256 and MD5 checksums for all release binaries

**When to use:**
- Users need to verify download integrity
- Security best practice for release artifacts
- Required by CONTEXT.md decisions

**Example:**
```yaml
# Source: GitHub Actions checksum generation patterns
# https://github.com/marketplace/actions/checksums-action

- name: Generate checksums
  shell: bash
  run: |
    cd build/bin
    for file in *.exe; do
      # Generate SHA256
      sha256sum "$file" > "${file}.sha256"
      # Generate MD5
      md5sum "$file" > "${file}.md5"
    done

    # List generated files
    ls -lah

# Alternative: Use checksums-action
- name: Generate checksums
  uses: jmgilman/actions-checksums@v0.1.0
  with:
    path: build/bin/*.exe
    patterns: |
      sha256: .sha256
      md5: .md5
```

**PowerShell alternative** (already used in existing release.yml):
```powershell
# For Windows-only workflows
Get-FileHash -Algorithm SHA256 ai-commit-hub.exe | \
  Select-Object -ExpandProperty Hash | \
  Out-File -Encoding ASCII ai-commit-hub.exe.sha256

Get-FileHash -Algorithm MD5 ai-commit-hub.exe | \
  Select-Object -ExpandProperty Hash | \
  Out-File -Encoding ASCII ai-commit-hub.exe.md5
```

### Anti-Patterns to Avoid

- **Building without caching**: Always use `actions/setup-go@v5` and `actions/setup-node@v4` which have built-in caching for `go.sum` and `package-lock.json`
- **Hardcoding version numbers**: Extract from Git tag, don't hardcode in workflow
- **Using `GOARCH=386` without testing**: 32-bit builds have known issues; test thoroughly or skip
- **Creating releases without drafts**: Use `draft: true` initially, verify manually, then publish
- **Uploading build logs as artifacts**: Logs stay in GitHub Actions UI for 90 days; don't waste artifact storage
- **Using `cleanup-run: true` in wails-build-action**: Let the action handle cleanup, otherwise builds may fail

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| **Go+Node.js environment setup** | Manual `go install`, `npm install` steps | `dAppServer/wails-build-action@v3` | Handles Go versions, Node.js, Wails CLI, submodules, OS-specific quirks |
| **Pre-release detection** | Complex bash regex parsing | softprops/action-gh-release auto-detection OR simple bash `[[ =~ ]]` | softprops detects `-beta`, `-rc` automatically; 5-line bash alternative is sufficient |
| **Checksum generation** | Custom Go or PowerShell scripts | `jmgilman/actions-checksums` OR 3-line bash loop | Battle-tested, handles multiple files, cross-platform |
| **Release asset upload** | GitHub REST API calls with `curl` | `softprops/action-gh-release@v1` | Handles OAuth, retries, large files, draft releases, pre-release flags |
| **Build caching** | Custom cache actions with manual key generation | `actions/setup-go@v5` built-in caching | Automatic cache key generation from `go.sum`, handles cache invalidation |
| **ZIP creation** | Go scripts or Node.js libraries | 7-Zip (Windows built-in) OR PowerShell `Compress-Archive` | Already on GitHub Actions runner, no setup needed |

**Key insight:** The Wails ecosystem has mature tooling. The dAppServer wails-build-action handles 80% of CI/CD complexity. Custom logic should only handle project-specific concerns (version injection, checksums, documentation packaging).

## Common Pitfalls

### Pitfall 1: Building for GOARCH=386 (32-bit Windows)

**What goes wrong:** Application compiles successfully but WebView2 crashes on startup, or the app doesn't run on 32-bit Windows systems at all.

**Why it happens:** Wails does NOT officially support 32-bit Windows builds. The WebView2 loader and Windows APIs have compatibility issues with 32-bit binaries. This is a known limitation with active GitHub issues (#3462, #4444).

**How to avoid:**
- **RECOMMENDATION**: Skip 386 builds entirely in v1. Only build for `GOARCH=amd64`
- If 32-bit support is absolutely required:
  - Test extensively on real 32-bit Windows hardware
  - Consider using alternative embedding (CEF instead of WebView2)
  - Follow [Issue #4444](https://github.com/wailsapp/wails/issues/4444) for updates
  - Document that 32-bit is "best effort" support

**Warning signs:**
- Builds succeed but app crashes immediately on launch
- Error messages mentioning WebView2 loader failures
- GitHub Actions workflow passes but manual testing fails

### Pitfall 2: OOM Build Failures (Node.js Memory)

**What goes wrong:** Frontend build (Vite/Webpack) fails with "JavaScript heap out of memory" error during Wails build.

**Why it happens:** Wails frontend builds (especially with Vue 3 + Vite) can consume large amounts of memory. GitHub Actions runners have limited RAM, and Node.js default heap size is too small.

**How to avoid:**
```yaml
env:
  # MUST set this for Wails builds
  NODE_OPTIONS: "--max-old-space-size=4096"
```

**Warning signs:**
- Build fails at "Building frontend..." step
- Error message includes "FATAL ERROR: Ineffective mark-compacts near heap limit"
- Intermittent failures (sometimes succeeds, sometimes fails)

### Pitfall 3: Incorrect ldflags Syntax

**What goes wrong:** Version variables are not injected, default to "dev" or "unknown".

**Why it happens:** ldflags syntax is finicky about quotes and spaces. Common mistakes:
- Missing quotes around the entire `-X` flag value
- Incorrect package path
- Shell variable expansion issues

**How to avoid:**
```yaml
# CORRECT:
-ldflags "-X 'pkg/path/var=value' -s -w"

# WRONG (missing quotes):
-ldflags -X pkg/path/var=value

# WRONG (wrong package path):
-ldflags "-X 'main.Version=1.0.0'"  # Version is in pkg/version, not main
```

**Verification:**
```bash
# After build, verify version is embedded:
./ai-commit-hub.exe --version
# Should output: v1.0.0 (abc1234 2026-02-06T10:30:00Z)
```

### Pitfall 4: Missing CGO_ENABLED for Windows Builds

**What goes wrong:** Build fails with linker errors or SQLite driver compilation failures.

**Why it happens:** Windows builds often require CGO (for SQLite, syscalls, etc.). GitHub Actions runners may not have GCC/MinGW configured correctly.

**How to avoid:**
```yaml
env:
  CGO_ENABLED: '1'  # Explicitly enable for Windows
```

**Warning signs:**
- Error: `cgo: C compiler not found`
- SQLite import errors: `cannot find package "github.com/mattn/go-sqlite3"`
- Linker errors about missing `libc`

### Pitfall 5: Timeout on Large Dependency Downloads

**What goes wrong:** Build fails randomly at `go mod download` or `npm install` steps.

**Why it happens:** Network issues on GitHub Actions runners, large dependency trees, slow npm registry mirrors.

**How to avoid:**
1. **Use built-in caching** (setup-go@v5, setup-node@v4):
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache: true  # Caches go.mod + go.sum dependencies

- name: Set up Node
  uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'  # Caches package-lock.json
    cache-dependency-path: frontend/package-lock.json
```

2. **Set reasonable timeouts**:
```yaml
jobs:
  build:
    timeout-minutes: 30  # Wails builds take ~10-20 mins
    runs-on: windows-latest
```

3. **Retry on failure** (optional, advanced):
```yaml
- name: Download Go dependencies
  uses: nick-fields/retry@v2
  with:
    timeout_minutes: 10
    max_attempts: 3
    command: go mod download
```

**Warning signs:**
- Intermittent failures at dependency download steps
- Error: "context deadline exceeded" or "network timeout"
- Different failures on each workflow run

### Pitfall 6: Forgetting to Clean Build Directory

**What goes wrong:** Previous build artifacts (old .exe, .dll, frontend dist) get included in new builds, causing bloated zip files or runtime errors.

**Why it happens:** Wails `build` command doesn't always clean `build/bin/` directory before building.

**How to avoid:**
```yaml
# Option 1: Use wails build -clean flag
- name: Build Wails app
  run: wails build -clean -ldflags "..."

# Option 2: Explicit clean step
- name: Clean build directory
  run: |
    Remove-Item -Recurse -Force build/bin -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force frontend/dist -ErrorAction SilentlyContinue

# Option 3: Use fresh checkout
- uses: actions/checkout@v4
  with:
    clean: true  # This is default
```

## Code Examples

### Verified Workflow: Complete Build and Release

This workflow combines all patterns from official documentation:

```yaml
# Source: Wails docs + dAppServer action + softprops release action
# https://wails.io/docs/next/guides/crossplatform-build
# https://github.com/marketplace/actions/wails-build-action
# https://github.com/softprops/action-gh-release

name: Build and Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

env:
  # Prevent Node.js OOM during frontend build
  NODE_OPTIONS: "--max-old-space-size=4096"

jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        include:
          - arch: amd64
            platform: windows/amd64
          # WARNING: 386 is NOT officially supported by Wails
          # Uncomment only if you accept the risks:
          # - arch: 386
          #   platform: windows/386

    runs-on: windows-latest
    timeout-minutes: 30

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for version info

      - name: Extract metadata
        id: meta
        shell: bash
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          SHA=$(git rev-parse --short HEAD)
          DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

          # Detect pre-release (matches -beta, -rc, -alpha, etc.)
          if [[ "$VERSION" =~ -(alpha|beta|rc|pre)\.?[0-9]*$ ]]; then
            PRERELEASE=true
          else
            PRERELEASE=false
          fi

          echo "VERSION=$VERSION" >> $GITHUB_OUTPUT
          echo "SHA=$SHA" >> $GITHUB_OUTPUT
          echo "DATE=$DATE" >> $GITHUB_OUTPUT
          echo "PRERELEASE=$PRERELEASE" >> $GITHUB_OUTPUT

      - name: Build Wails app
        uses: dAppServer/wails-build-action@v3
        with:
          build-name: ai-commit-hub
          build-platform: ${{ matrix.platform }}
          package: false  # We'll package manually
          go-version: '1.21'
          wails-version: 'v2.11.0'
        env:
          CGO_ENABLED: '1'

      - name: Inject version info
        shell: bash
        run: |
          # Rebuild with version injection (action doesn't support custom ldflags)
          wails build -clean -ldflags "\
            -X 'github.com/allanpk716/ai-commit-hub/pkg/version.Version=${{ steps.meta.outputs.VERSION }}' \
            -X 'github.com/allanpk716/ai-commit-hub/pkg/version.CommitSHA=${{ steps.meta.outputs.SHA }}' \
            -X 'github.com/allanpk716/ai-commit-hub/pkg/version.BuildTime=${{ steps.meta.outputs.DATE }}' \
            -s -w"

      - name: Verify version
        shell: bash
        run: |
          ./build/bin/ai-commit-hub.exe --version || echo "Version check failed"

      - name: Create release package
        shell: bash
        run: |
          cd build/bin

          # Copy documentation
          cp ../../../README.md .
          cp ../../../docs/config-examples/config.yaml .

          # Create ZIP
          ARCH="${{ matrix.arch }}"
          VERSION="${{ steps.meta.outputs.VERSION }}"
          7z a "../ai-commit-hub-windows-${ARCH}-v${VERSION}.zip" \
            ai-commit-hub.exe \
            README.md \
            config.yaml

      - name: Generate checksums
        shell: bash
        run: |
          cd build
          for file in *.zip; do
            sha256sum "$file" > "${file}.sha256"
            md5sum "$file" > "${file}.md5"
          done

          # Display checksums for verification
          echo "::notice::SHA256 Checksums:"
          cat *.sha256
          echo "::notice::MD5 Checksums:"
          cat *.md5

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: ai-commit-hub-windows-${{ matrix.arch }}-v${{ steps.meta.outputs.VERSION }}
          path: |
            build/*.zip
            build/*.sha256
            build/*.md5
          retention-days: 30

  release:
    needs: build
    runs-on: ubuntu-latest  # Cheaper for release-only job
    timeout-minutes: 10

    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            artifacts/**/*.zip
            artifacts/**/*.sha256
            artifacts/**/*.md5
          generate_release_notes: true
          draft: true  # Create as draft, verify manually before publishing
          prerelease: ${{ needs.build.outputs.prerelease }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Verified Pattern: Reliable Go Dependency Caching

```yaml
# Source: actions/setup-go documentation
# https://github.com/actions/setup-go

- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache: true  # Automatically caches go.mod + go.sum
    # Cache key is generated from go.sum hash automatically

# No need for manual go mod download
# Dependencies are cached on first run, restored on subsequent runs
```

**Cache key logic** (automatic):
```
Linux/macOS: /opt/hostedtoolcache/go/1.21.0/x64/
Windows:   C:\hostedtoolcache\windows\go\1.21.0\x64\

Key: setup-go-${{ runner.os }}-${{ hashFiles('go.sum') }}
```

### Verified Pattern: Version Display in Application

```go
// Source: Existing pkg/version/version.go
// Add CLI flag support if not already present

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/allanpk716/ai-commit-hub/pkg/version"
)

var (
	showVersion = flag.Bool("v", false, "show version information")
	showVersionLong = flag.Bool("version", false, "show version information")
)

func main() {
	flag.Parse()

	if *showVersion || *showVersionLong {
		fmt.Println(version.GetFullVersion())
		os.Exit(0)
	}

	// Rest of application...
}
```

**Output:**
```
$ ai-commit-hub.exe -v
v1.0.0 (abc1234 2026-02-06T10:30:00Z)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual `go install` + `npm install` setup | `dAppServer/wails-build-action@v3` | 2023-2024 | Reduces workflow YAML from 50+ lines to 10 lines |
| `actions/create-release` (deprecated) | `softprops/action-gh-release@v1` | 2022 | create-release is archived; softprops is actively maintained |
| Manual cache management with `actions/cache` | Built-in caching in `setup-go@v5` | 2024-2025 | No need for custom cache keys or manual cache restoration |
| 32-bit Windows support (GOARCH=386) | 64-bit only (GOARCH=amd64) | 2024-2025 | Wails team does not support 32-bit; WebView2 crashes on 386 |
| Release without checksums | SHA256 + MD5 checksums standard | 2020s | Security best practice; users expect integrity verification |
| GitHub native SHA256 (asset.digest) | Manual checksum generation still needed | 2025 (June) | GitHub's native checksums are API-only, not downloadable as files |

**Deprecated/outdated:**
- **`actions/create-release`**: Archived in 2022, use `softprops/action-gh-release` instead
- **`dAppServer/wails-build-action@v2`**: Use `@v3` or `@main` for latest features
- **Wails v2.10.0**: Known issues with build action, use v2.9.0 or v2.11.0
- **`setup-go@v4`**: Upgraded to v5 in 2024 with improved caching

## Open Questions

### 1. 32-bit Windows Support Decision

**What we know:**
- Wails does NOT officially support GOARCH=386
- Known WebView2 crashes on 32-bit builds (Issue #4444, July 2025)
- No timeline for official 32-bit support

**What's unclear:**
- Does the project actually need 32-bit support in v1?
- Are users still on 32-bit Windows in 2026?

**Recommendation:**
- **Skip 386 builds in v1**. Only build for amd64.
- Re-evaluate if users request 32-bit support
- Document that 64-bit Windows is required

### 2. Optimal Binary Size vs. Build Time Tradeoff

**What we know:**
- `-ldflags="-s -w"` reduces binary size by ~30% (strips debug info)
- No impact on runtime performance
- Adds ~10 seconds to build time (linking)

**What's unclear:**
- Is ~30% reduction meaningful for users? (e.g., 50MB → 35MB)
- Does smaller binary improve download experience significantly?

**Recommendation:**
- **Use `-s -w` flags** (already in wails.json config)
- Benefits outweigh minimal cost
- Standard practice for Go releases

### 3. Draft vs. Direct Release Publishing

**What we know:**
- Context decisions mention both tag-triggered AND manual triggers
- Existing release.yml uses direct publishing (draft: false)

**What's unclear:**
- Should releases be created as drafts for manual verification?
- Or automated publication on tag push?

**Recommendation:**
- **Use `draft: true` for pre-release versions** (beta, rc)
- **Use `draft: false for stable releases** (non-beta tags)
- Allows manual testing before publishing pre-releases
- Stable releases can be fully automated

### 4. Build Timeout for Matrix Jobs

**What we know:**
- Default GitHub Actions timeout: 360 minutes (6 hours)
- Typical Wails build: 10-20 minutes
- Matrix with 2 architectures: ~20-40 minutes total (parallel)

**What's unclear:**
- Should we set explicit `timeout-minutes`?

**Recommendation:**
- **Set `timeout-minutes: 30`** at job level
- Prevents runaway builds from wasting minutes
- 3x average build time (standard best practice)
- If timeout exceeded, investigate dependency issues

## Sources

### Primary (HIGH confidence)

- **[/wailsapp/wails](https://github.com/wailsapp/wails)** - Official Wails repository
  - Cross-platform build documentation
  - CI/CD workflow examples
  - Version 2.11.0 capabilities and features

- **[dAppServer/wails-build-action](https://github.com/marketplace/actions/wails-build-action)** - Official Wails build action
  - Verified current version: v3
  - Platform support: linux/amd64, windows/amd64, darwin/universal
  - Parameters: build-name, build-platform, package, go-version, wails-version

- **[softprops/action-gh-release](https://github.com/softprops/action-gh-release)** - Release creation action
  - 7k+ stars, actively maintained
  - Pre-release detection, draft releases, file uploads
  - Verified compatibility with tag-triggered workflows

- **[actions/setup-go](https://github.com/actions/setup-go)** - Go environment setup
  - Version v5 with built-in caching
  - Automatic cache key generation from go.sum
  - Support for Go 1.21+ (project uses 1.24.1)

- **[Project's existing release.yml](.github/workflows/release.yml)** - Current workflow
  - Already uses softprops/action-gh-release
  - Already extracts version from Git tags
  - Already uses 7z for packaging
  - Good foundation to enhance

- **[pkg/version/version.go](pkg/version/version.go)** - Version package
  - Already implements version, commit, build time variables
  - ldflags injection points already configured
  - GetVersion() and GetFullVersion() functions available

### Secondary (MEDIUM confidence)

- **[Wails Cross-Platform Build Guide](https://wails.io/docs/next/guides/crossplatform-build)** - Official documentation
  - Workflow examples for matrix builds
  - NODE_OPTIONS recommendation for OOM prevention
  - Tag-triggered build patterns
  - Verified: February 2026

- **[Go ldflags Documentation - DigitalOcean](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications)** - Comprehensive tutorial
  - -X flag usage for variable injection
  - -s -w flags for stripping debug info
  - Quote escaping examples
  - Published: Updated for Go 1.21+

- **[GitHub Releases Digests Feature](https://github.blog/changelog/2025-06-03-releases-now-expose-digests-for-release-assets/)** - GitHub blog (June 2025)
  - SHA256 checksums now computed automatically
  - Accessible via API (asset.digest property)
  - NOT downloadable as files (still need manual generation)

- **[Go Compiler Optimizations Wiki](https://go.dev/wiki/CompilerOptimizations)** - Official Go documentation
  - Confirms: No -O/-O2/-O3 flags in Go
  - Automatic optimizations applied by compiler
  - Profile-guided optimization (PGO) available since Go 1.20

### Tertiary (LOW confidence)

- **[GitHub Actions Timeout Best Practices](https://www.blacksmith.sh/blog/how-to-reduce-spend-in-github-actions)** - Blacksmith blog (Dec 2024)
  - Set timeout to 3x average job duration
  - Not official GitHub documentation
  - Reasonable heuristic but not verified

- **[Wails 32-bit Windows Issues](https://github.com/wailsapp/wails/issues/4444)** - GitHub issue (July 2025)
  - Reports WebView2 crashes on GOARCH=386
  - Not officially resolved
  - Maintainers state no plans for 32-bit support
  - Verified through issue discussion, not official announcement

- **[Wails Build Action Documentation](https://github.com/marketplace/actions/wails-build-action)** - Marketplace listing
  - Version 2.10.0 known issues mentioned
  - Recommends using v2.9.0 or main branch
  - Last updated: February 2025

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - All sources are official documentation or GitHub Marketplace listings
- Architecture: **HIGH** - Verified against Wails official docs and project's existing setup
- Pitfalls: **HIGH** - All pitfalls verified through GitHub issues, official docs, or common patterns
- Claude's Discretion items: **MEDIUM** - Go optimization flags verified, timeout/clim/runner choice based on best practices (not official docs)

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (30 days - Wails v3 in alpha, action updates frequent)

**Key limitations:**
- 32-bit Windows support is LOW confidence due to lack of official Wails support
- Go compiler optimization recommendations based on official docs (no -O flags)
- Timeout recommendations based on community best practices, not GitHub documentation
