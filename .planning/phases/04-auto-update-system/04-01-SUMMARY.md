---
phase: 04-auto-update-system
plan: 01
subsystem: update-service
tags: [semver, github-api, version-checking, prerelease, background-tasks]

# Dependency graph
requires:
  - phase: 03-system-tray-fixes
    provides: systray integration with stub check-update menu
provides:
  - Version comparison supporting prerelease identifiers (alpha/beta/rc)
  - Update service using /releases endpoint with caching
  - Background update checking every 24 hours
  - UpdateInfo model with prerelease metadata
affects: [04-02-download-install, ui-settings]

# Tech tracking
tech-stack:
  added: [golang.org/x/mod/semver v0.32.0]
  patterns: [version caching with TTL, background goroutines with ticker, semver sorting]

key-files:
  created: []
  modified: [pkg/version/version.go, pkg/service/update_service.go, pkg/models/update_info.go, app.go]

key-decisions:
  - "Used golang.org/x/mod/semver for robust version comparison with prerelease support"
  - "24-hour cache TTL to balance freshness with API rate limits"
  - "Fallback to cached results on rate limit errors for resilience"
  - "Background check runs independently of startup check for immediate feedback"

patterns-established:
  - "Version normalization: Ensure 'v' prefix for semver library compatibility"
  - "Cache pattern: Use sync.RWMutex for thread-safe cached reads"
  - "Error resilience: Return stale cache on rate limit instead of failing"
  - "Background tasks: Launch goroutine in service StartBackgroundCheck() method"

# Metrics
duration: 8min
completed: 2026-02-07
---

# Phase 4 Plan 1: 版本检测和 UI 集成 Summary

**使用 semver 库实现预发布版本比较和 24 小时后台自动检查更新**

## Performance

- **Duration:** 8 min
- **Started:** 2026-02-07T09:50:30Z
- **Completed:** 2026-02-07T09:58:00Z
- **Tasks:** 4
- **Files modified:** 4

## Accomplishments

- **版本比较支持预发布标识符** - 使用 golang.org/x/mod/semver 实现标准语义化版本比较，正确处理 alpha/beta/rc 预发布版本
- **从 /releases 端点获取所有版本** - 替换 /releases/latest，支持检测预发布版本，使用 semver.Sort 找到最新版本
- **24 小时后台自动检查** - 启动时立即检查，之后每 24 小时后台自动检查更新
- **缓存机制提升可靠性** - 24 小时缓存避免频繁 API 调用，速率限制时返回缓存结果

## Task Commits

Each task was committed atomically:

1. **Task 1: 增强版本比较支持预发布版本** - `5ad7cd5` (feat)
2. **Task 2: 重构 UpdateService 使用 /releases 端点** - `91a4c7c` (feat)
   - Includes Task 3: 增强 UpdateInfo 模型
3. **Task 4: 完善 app.go 更新检查集成** - `6781dc0` (feat)

**Plan metadata:** (pending final commit)

## Files Created/Modified

- `pkg/version/version.go` - 添加 semver 库支持，新增 NormalizeVersion、SafeCompareVersions、IsPrerelease 函数
- `pkg/version/version_test.go` - 添加预发布版本比较测试用例
- `pkg/service/update_service.go` - 重构使用 /releases 端点，添加缓存和后台检查
- `pkg/models/update_info.go` - 添加 IsPrerelease 和 PrereleaseType 字段
- `app.go` - 连接 StartBackgroundCheck() 和实现 checkUpdateStub()

## Decisions Made

- **使用 golang.org/x/mod/semver** - 官方维护的语义化版本库，支持预发布版本比较（alpha < beta < rc < stable）
- **24 小时缓存 TTL** - 平衡新鲜度和 API 速率限制，避免频繁调用 GitHub API
- **速率限制时返回缓存** - 提高服务韧性，即使 API 失败也能返回已知版本信息
- **独立的后台检查 goroutine** - 使用 time.NewTicker(24*time.Hour) 定时触发，不阻塞主线程

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- **ParseVersion 签名变更** - 添加 prerelease 返回参数后，需要同步更新测试文件中的调用（4 值改为 5 值）。已修复。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

**Ready for next phase:**

- UpdateService 完整实现版本检测，支持预发布版本
- 后台自动检查已启动，每 24 小时运行
- UpdateInfo 包含完整版本信息（包括预发布标识）
- 托盘菜单"检查更新"按钮已连接到真实 API

**Blockers/Concerns:**

- 托盘通知显示更新提醒功能待实现（Phase 4-02）
- 设置页面版本信息展示待集成（Phase 4-03 或 4-04）

**Verification needed:**

- 启动应用确认后台检查日志
- 点击托盘"检查更新"确认功能正常
- 检查日志确认 24 小时定时器启动

---
*Phase: 04-auto-update-system*
*Completed: 2026-02-07*
