---
phase: 01-ci-cd-pipeline
plan: 03
subsystem: ci-cd
tags: github-actions, release, workflow, artifacts, checksums

# Dependency graph
requires:
  - phase: 01-ci-cd-pipeline
    plan: 02
    provides: Build artifacts (ZIP, SHA256, MD5) with version injection
provides:
  - Automatic GitHub Release creation on version tag push
  - Pre-release detection and flagging (alpha, beta, rc, pre)
  - Auto-generated release notes from commit messages
  - Complete artifact package (ZIP + checksums) published to releases
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Two-job workflow pattern (build on Windows, release on Ubuntu)
    - Job outputs for cross-job data sharing
    - Artifact download/upload for build/release separation
    - softprops/action-gh-release for release automation

key-files:
  created: []
  modified:
    - .github/workflows/build.yml - Enhanced with release job
    - .github/workflows/release.yml - Removed (obsolete)

key-decisions:
  - "Separate jobs: build on Windows (requires Wails), release on Ubuntu (cheaper)"
  - "Job outputs: VERSION and PRERELEASE shared between jobs"
  - "Pre-release detection: regex pattern for alpha/beta/rc/pre tags"
  - "Auto-generated release notes: leverage GitHub's automatic changelog"
  - "Draft: false for stable releases - immediate publication"

patterns-established:
  - "Pattern: Multi-job workflows with needs dependency"
  - "Pattern: Job outputs for cross-job communication"
  - "Pattern: Artifact staging between jobs"
  - "Pattern: Pre-release detection via semantic versioning"

# Metrics
duration: 1min
completed: 2026-02-06
---

# Phase 1 Plan 3: Automatic Release Creation Summary

**GitHub Release automation with pre-release detection and auto-generated changelog using softprops/action-gh-release**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-06T02:14:20Z
- **Completed:** 2026-02-06T02:15:37Z
- **Tasks:** 2
- **Files modified:** 2 (1 enhanced, 1 deleted)

## Accomplishments

- **Complete CI/CD pipeline** - Build → Package → Release fully automated
- **Pre-release detection** - Alpha/beta/rc/pre versions automatically flagged
- **Auto-generated release notes** - GitHub creates changelog from commits
- **Cost optimization** - Release job runs on Ubuntu (cheaper than Windows)

## Task Commits

Each task was committed atomically:

1. **Task 1-3: Add release job to workflow** - `f36a0fd` (feat)
   - Added build job outputs (VERSION, PRERELEASE)
   - Created separate release job with ubuntu-latest runner
   - Configured artifact download from build job
   - Set up softprops/action-gh-release with all artifacts

2. **Task 4: Remove obsolete release.yml** - `afeb344` (feat)
   - Deleted old workflow with incorrect ldflags
   - Consolidated all CI/CD into single build.yml

**Plan metadata:** Not applicable (summary created in same session)

## Files Created/Modified

- `.github/workflows/build.yml` - Enhanced with release job
  - Added job outputs for VERSION and PRERELEASE
  - Added release job (ubuntu-latest, 10min timeout)
  - Configured softprops/action-gh-release@v1
  - Artifact download/upload pattern established
- `.github/workflows/release.yml` - **Deleted** (obsolete)

## Workflow Structure

### Trigger Behavior
```yaml
on:
  push:
    tags:
      - 'v*'  # Matches v1.0.0, v2.1.3-beta, etc.
  workflow_dispatch:  # Manual trigger from GitHub UI
```

### Job Architecture

**Build Job** (Windows-latest, 30min timeout)
1. Extract version metadata (VERSION, SHA, DATE, PRERELEASE)
2. Build Wails app with version injection
3. Create ZIP package with exe + README + config
4. Generate SHA256 and MD5 checksums
5. Upload artifacts (30-day retention)
6. **Export outputs:** VERSION, PRERELEASE

**Release Job** (Ubuntu-latest, 10min timeout)
1. Download artifacts from build job
2. Create GitHub Release with:
   - All artifacts (.zip, .sha256, .md5)
   - Version name (v{VERSION})
   - Auto-generated release notes
   - Pre-release flag (from PRERELEASE output)

### Release Artifact Contents

Each release includes:
- `ai-commit-hub-windows-amd64-v{VERSION}.zip` - Application package
- `ai-commit-hub-windows-amd64-v{VERSION}.zip.sha256` - SHA256 checksum
- `ai-commit-hub-windows-amd64-v{VERSION}.zip.md5` - MD5 checksum

### Pre-release Detection

```bash
if [[ "${VERSION}" =~ -(alpha|beta|rc|pre) ]]; then
  echo "PRERELEASE=true" >> $GITHUB_OUTPUT
else
  echo "PRERELEASE=false" >> $GITHUB_OUTPUT
fi
```

**Examples:**
- `v1.0.0` → PRERELEASE=false (stable release)
- `v1.0.0-alpha` → PRERELEASE=true
- `v2.1.3-beta.1` → PRERELEASE=true
- `v1.5.0-rc.2` → PRERELEASE=true

### Release Notes Generation

GitHub automatically generates release notes by comparing commits:
- Since the previous release tag
- Formats into changelog with commit grouping
- Includes contributors and commit links

## Decisions Made

**Job separation:** Build on Windows (Wails requirement), release on Ubuntu (cost optimization)
- Build job needs Windows for Wails/CGO
- Release job only needs GitHub API, Ubuntu is cheaper

**Draft releases:** Set to `draft: false`
- Stable versions publish immediately
- Pre-release versions marked but still published
- Faster feedback loop for users

**Permissions:** Explicit `contents: write` for release job
- Modern GitHub Actions requires explicit permissions
- Allows release creation without PAT (uses GITHUB_TOKEN)

**Artifact staging:** Upload in build, download in release
- Clean separation of concerns
- Allows artifact inspection before release
- Supports manual workflow reruns

## Deviations from Plan

None - plan executed exactly as written.

## Authentication Gates

None - no external authentication required.

## Next Phase Readiness

**CI/CD Pipeline Complete:**
- ✅ Automatic builds on version tags
- ✅ Version injection into binary
- ✅ Artifact packaging (ZIP + README + config)
- ✅ Dual checksums (SHA256 + MD5)
- ✅ Automatic GitHub Release creation
- ✅ Pre-release detection
- ✅ Auto-generated release notes

**Ready for Phase 2: Core Features**
All CI/CD infrastructure in place. Future releases will be fully automated when version tags are pushed.

**Usage:**
```bash
# Create stable release
git tag v1.0.0
git push origin v1.0.0

# Create pre-release
git tag v1.0.0-alpha
git push origin v1.0.0-alpha
```

---
*Phase: 01-ci-cd-pipeline*
*Plan: 03*
*Completed: 2026-02-06*
