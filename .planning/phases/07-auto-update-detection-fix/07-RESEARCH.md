# Phase 7: 自动更新检测修复 - Research

**Researched:** 2026-02-08
**Domain:** Go 后端 + Vue3 前端混合 (GitHub API/Atom feeds, UI 组件, 错误处理)
**Confidence:** HIGH

## Summary

本阶段研究修复 GitHub Releases 版本检测失败问题和增强用户反馈。核心发现：

1. **GitHub API 限制明显** - 未认证请求仅 60 次/小时，需要混合降级策略（API + Atom feed）
2. **Atom feeds 稳定可用** - GitHub 提供 `releases.atom` 端点，无需认证，可靠但非官方支持
3. **现有实现已完善** - 代码库已有完整的 `UpdateService`、版本比较（`golang.org/x/mod/semver`）、UI 组件，主要缺失是错误处理和用户反馈
4. **Vue3 + TypeScript 成熟** - 2026 年已为默认标准，组件驱动架构是主流

**Primary recommendation:** 优先实现 GitHub API + Atom feed 混合降级策略，增强错误提示和 UI 反馈，复用现有架构而非重写。

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/mod/semver` | latest | 语义化版本比较 | 官方库，已在代码库使用，支持预发布版本 |
| `github.com/mmcdole/gofeed` | latest | RSS/Atom feed 解析 | Go 生态标准 feed 解析库，高 benchmark 分数 (75.1) |
| `github.com/WQGroup/logger` | latest | 统一日志输出 | 项目已采用，符合 Wails 开发规范 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| 标准库 `net/http` | builtin | HTTP 客户端 | 所有网络请求，已配置 10s 超时 |
| Wails Events | builtin | 前后端事件通信 | 实时更新下载进度、版本检测状态 |
| Vue3 Composition API | builtin | 状态管理 | 已有 `updateStore.ts`，复用而非重写 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| GitHub API | GitHub GraphQL API | GraphQL 更复杂，未解决速率限制问题 |
| `gofeed` | 自定义 XML 解析 | 手写解析易出错，需处理多种 feed 格式 |
| 混合降级策略 | 纯 Atom feed | API 提供结构化数据，Atom 仅做备用 |

**Installation:**

```bash
# Go 依赖
go get github.com/mmcdole/gofeed

# 前端无新增依赖（复用现有 Vue3 + Pinia）
```

## Architecture Patterns

### Recommended Project Structure

```
pkg/service/
├── update_service.go          # 现有实现（增强错误处理）
├── update_service_test.go     # 单元测试
└── update_fallback.go         # 新增：Atom feed 降级实现

pkg/models/
└── update_info.go             # 现有模型（无需修改）

pkg/version/
├── version.go                 # 现有版本管理（已完善）
└── version_test.go

frontend/src/
├── components/
│   ├── AboutDialog.vue        # 现有组件（增强版本卡片）
│   └── VersionInfoCard.vue    # 新增：独立版本信息卡片
├── stores/
│   └── updateStore.ts         # 现有 store（增加错误状态）
└── types/
    └── update.ts              # TypeScript 类型定义
```

### Pattern 1: 混合降级策略 (Hybrid Fallback Strategy)

**What:** 优先使用 GitHub API（结构化数据），失败时降级到 Atom feed（无需认证）。

**When to use:** 所有需要高可用性的外部 API 调用场景。

**Example:**

```go
// Source: 基于 pkg/service/update_service.go 现有实现增强
func (s *UpdateService) CheckForUpdates() (*models.UpdateInfo, error) {
    // 尝试 GitHub API（优先）
    info, err := s.checkViaAPI()
    if err == nil {
        return info, nil
    }

    // 记录 API 失败
    logger.Warnf("GitHub API 失败，尝试 Atom feed: %v", err)

    // 降级到 Atom feed
    info, err = s.checkViaAtomFeed()
    if err != nil {
        // 两种方式都失败，返回清晰错误
        return nil, fmt.Errorf("无法获取版本信息：API 失败 (%w)，Atom feed 失败 (%v)", err, err)
    }

    logger.Info("通过 Atom feed 成功获取版本信息")
    return info, nil
}

func (s *UpdateService) checkViaAPI() (*models.UpdateInfo, error) {
    // 现有 fetchAllReleases() 逻辑
    // ...
}

func (s *UpdateService) checkViaAtomFeed() (*models.UpdateInfo, error) {
    // 新增实现
    url := fmt.Sprintf("https://github.com/%s/releases.atom", s.repo)

    // 使用 gofeed 解析
    fp := gofeed.NewParser()
    fp.Client = s.httpClient // 复用已有 http.Client（10s 超时）

    feed, err := fp.ParseURLWithURL(url)
    if err != nil {
        return nil, fmt.Errorf("解析 Atom feed 失败: %w", err)
    }

    if len(feed.Items) == 0 {
        return nil, fmt.Errorf("Atom feed 无内容")
    }

    // 取第一个 item（最新 release）
    latestItem := feed.Items[0]

    // 从 title 或 category 提取版本号
    latestVersion := s.extractVersionFromFeed(latestItem.Title)

    // 比较版本
    currentVersion := version.GetVersion()
    hasUpdate := s.compareVersions(latestVersion, currentVersion)

    return &models.UpdateInfo{
        HasUpdate:      hasUpdate,
        LatestVersion:  latestVersion,
        CurrentVersion: currentVersion,
        ReleaseNotes:   latestItem.Content,
        PublishedAt:    *latestItem.PublishedParsed,
        // Atom feed 无 download URL，需构造
        DownloadURL:    s.constructDownloadURL(latestVersion),
        // ...
    }, nil
}
```

### Pattern 2: 前端加载状态管理 (Loading State Management)

**What:** 检查更新时提供即时视觉反馈（loading 图标 + toast 提示）。

**When to use:** 所有耗时超过 200ms 的异步操作。

**Example:**

```typescript
// Source: 基于 frontend/src/stores/updateStore.ts 现有实现增强
async function checkForUpdates() {
  isChecking.value = true

  // 显示即时反馈
  showToast({
    message: '正在检查更新...',
    type: 'info',
    duration: 2000 // 2秒后自动消失
  })

  try {
    const { CheckForUpdates } = await import('../../wailsjs/go/main/App')
    const info = await CheckForUpdates()

    updateInfo.value = info
    hasUpdate.value = info.hasUpdate

    // 成功提示
    if (info.hasUpdate) {
      showToast({
        message: `发现新版本 ${info.latestVersion}`,
        type: 'success'
      })
    } else {
      showToast({
        message: '当前已是最新版本',
        type: 'success'
      })
    }

    return info
  } catch (error) {
    console.error('检查更新失败:', error)

    // 详细错误提示
    const errorMessage = formatUpdateError(error)
    showToast({
      message: `检查更新失败: ${errorMessage}`,
      type: 'error',
      duration: 5000 // 错误消息显示更久
    })

    throw error
  } finally {
    isChecking.value = false
  }
}

// 格式化错误消息（用户友好）
function formatUpdateError(error: unknown): string {
  const err = error as { code?: string; message?: string }

  if (err.message?.includes('403')) {
    return 'GitHub API 速率限制，请稍后再试'
  }

  if (err.message?.includes('timeout')) {
    return '网络连接超时，请检查网络设置'
  }

  if (err.message?.includes('failed to fetch')) {
    return '无法连接到 GitHub，请检查网络连接'
  }

  return err.message || '未知错误'
}
```

### Pattern 3: 版本信息卡片展示 (Version Info Card)

**What:** 在关于界面使用卡片式布局展示完整版本信息。

**When to use:** 需要清晰展示结构化信息的场景。

**Example:**

```vue
<!-- Source: 基于 frontend/src/components/AboutDialog.vue 现有实现增强 -->
<template>
  <div class="about-dialog">
    <!-- ... 现有内容 ... -->

    <!-- 新增：版本信息卡片 -->
    <div class="version-info-card">
      <div class="card-header">
        <h3>版本信息</h3>
        <button
          @click="checkForUpdates"
          :disabled="isChecking"
          class="check-update-btn"
        >
          <span v-if="isChecking" class="loading-icon">⏳</span>
          <span v-else class="refresh-icon">🔄</span>
          <span>{{ isChecking ? '检查中...' : '检查更新' }}</span>
        </button>
      </div>

      <div class="card-body">
        <!-- 当前版本 -->
        <div class="info-row">
          <span class="label">当前版本:</span>
          <span class="value">{{ version }}</span>
        </div>

        <!-- 最新版本 -->
        <div v-if="updateInfo" class="info-row">
          <span class="label">最新版本:</span>
          <span class="value">{{ updateInfo.latestVersion }}</span>
        </div>

        <!-- 更新状态 -->
        <div v-if="updateInfo" class="info-row status">
          <span class="label">更新状态:</span>
          <span
            :class="['value', 'status-badge', updateInfo.hasUpdate ? 'has-update' : 'latest']"
          >
            {{ updateInfo.hasUpdate ? '有新版本可用' : '已是最新版本' }}
          </span>
        </div>

        <!-- 发布时间 -->
        <div v-if="updateInfo && updateInfo.publishedAt" class="info-row">
          <span class="label">发布时间:</span>
          <span class="value">{{ formatDate(updateInfo.publishedAt) }}</span>
        </div>

        <!-- 更新说明（折叠） -->
        <details v-if="updateInfo && updateInfo.releaseNotes" class="changelog">
          <summary>更新说明</summary>
          <div class="changelog-content" v-html="renderMarkdown(updateInfo.releaseNotes)"></div>
        </details>

        <!-- 下载链接 -->
        <div v-if="updateInfo && updateInfo.downloadURL" class="info-row">
          <a :href="updateInfo.downloadURL" target="_blank" class="download-link">
            🔗 在 GitHub 下载
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { GetVersion, GetFullVersion } from '../../wailsjs/go/main/App'
import { useUpdateStore } from '../stores/updateStore'

const updateStore = useUpdateStore()
const version = ref('加载中...')
const updateInfo = ref(null)

// 组件挂载时自动检查更新
onMounted(async () => {
  version.value = await GetVersion()
  await checkForUpdates()
})

async function checkForUpdates() {
  try {
    updateInfo.value = await updateStore.checkForUpdates()
  } catch (error) {
    console.error('检查更新失败:', error)
  }
}

function formatDate(date: Date): string {
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}
</script>

<style scoped>
.version-info-card {
  background: #f9fafb;
  border-radius: 8px;
  padding: 20px;
  margin: 24px 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.check-update-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.check-update-btn:hover:not(:disabled) {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.check-update-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #e5e7eb;
}

.info-row:last-child {
  border-bottom: none;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.has-update {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.latest {
  background: #d1fae5;
  color: #065f46;
}

.changelog {
  margin-top: 12px;
}

.changelog summary {
  cursor: pointer;
  color: #3b82f6;
  font-weight: 500;
}

.changelog-content {
  margin-top: 8px;
  padding: 12px;
  background: white;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.6;
}

.download-link {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
}

.download-link:hover {
  text-decoration: underline;
}
</style>
```

### Anti-Patterns to Avoid

- **错误信息过度技术化**: 不展示原始错误堆栈给用户，使用友好提示
- **无限制重试**: API 速率限制后应等待而非立即重试（使用 24 小时缓存）
- **阻塞 UI**: 检查更新不应阻塞主线程，使用异步 + 加载状态
- **忽略预发布版本**: 现有实现已正确处理，保持 `IsPrerelease` 字段

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| RSS/Atom feed 解析 | 手写 XML 解析 | `github.com/mmcdole/gofeed` | 需处理 RSS/Atom/JSON Feed 多种格式，边界情况复杂 |
| 版本比较 | 字符串比较 | `golang.org/x/mod/semver` | 需处理预发布版本 (alpha/beta/rc)，已内置 |
| HTTP 超时控制 | context 手动管理 | 复用现有 `http.Client` | 已配置 10s 超时，避免重复造轮子 |
| Markdown 渲染 | 正则替换 | `marked` 或 `markdown-it` | Vue 生态标准，防止 XSS 攻击 |

**Key insight:** Feed 解析和版本比较有大量边界情况（如 "v1.0.0-beta.1" vs "v1.0.0"），手写极易出错。现有 `UpdateService` 已正确使用 `semver` 库，只需增强错误处理。

## Common Pitfalls

### Pitfall 1: GitHub API 速率限制耗尽

**What goes wrong:** 未认证请求仅 60 次/小时，频繁检查更新导致 403 错误。

**Why it happens:** 未实现缓存机制，每次打开关于界面都调用 API。

**How to avoid:**
- 使用 24 小时缓存（现有代码已实现）
- 优先使用 API，失败时降级到 Atom feed
- 清晰提示用户速率限制错误

**Warning signs:** 日志中出现频繁的 "403 Forbidden" 或 "rate limit exceeded"。

### Pitfall 2: Atom feed 版本号解析错误

**What goes wrong:** Atom feed 的 `title` 可能包含非版本信息（如 "Release v1.0.0"）。

**Why it happens:** 直接使用 `title` 而非提取版本号部分。

**How to avoid:**
```go
// 使用正则提取版本号
func (s *UpdateService) extractVersionFromFeed(title string) string {
    // 匹配 v1.2.3 或 1.2.3 格式
    re := regexp.MustCompile(`v?\d+\.\d+\.\d+(-[0-9A-Za-z-]+)?`)
    match := re.FindString(title)
    if match == "" {
        return title // 降级：返回原始 title
    }
    return match
}
```

**Warning signs:** 版本比较失败，`semver.IsValid()` 返回 false。

### Pitfall 3: UI 状态不同步

**What goes wrong:** 用户点击"检查更新"按钮无反馈，不知道是否在执行。

**Why it happens:** 未设置 `isChecking` 状态或未禁用按钮。

**How to avoid:**
- 立即设置 `isChecking.value = true`
- 按钮添加 `:disabled="isChecking"`
- 显示 loading 图标（如旋转的 🔄）
- 显示 toast 提示（"正在检查更新..."）

**Warning signs:** 用户多次点击按钮导致重复请求。

### Pitfall 4: 预发布版本处理不当

**What goes wrong:** 将 beta 版本误判为"最新稳定版本"。

**Why it happens:** 未检查 `IsPrerelease` 字段或 `semver.Prerelease()` 返回值。

**How to avoid:**
- 现有代码已正确处理（`models.UpdateInfo` 有 `IsPrerelease` 字段）
- UI 上显示预发布标识（如 "v1.0.0-beta.1"）
- 允许用户选择是否接收预发布版本

**Warning signs:** 用户体验到不稳定版本。

### Pitfall 5: Atom feed 未官方支持

**What goes wrong:** GitHub 随时可能修改 Atom feed 格式或移除端点。

**Why it happens:** Atom feeds 是" undocumented and unsupported"（GitHub 官方文档明确）。

**How to avoid:**
- 仅将 Atom feed 作为降级方案，非主要数据源
- 记录日志：依赖 Atom feed 时应警告
- 监控 GitHub 变更（如订阅 GitHub Changelog）

**Warning signs:** Atom feed 解析频繁失败或返回空数据。

## Code Examples

Verified patterns from official sources:

### Atom Feed 解析（带超时和错误处理）

```go
// Source: https://github.com/mmcdole/gofeed (Context7 文档)
package service

import (
    "context"
    "fmt"
    "time"
    "github.com/mmcdole/gofeed"
    "github.com/WQGroup/logger"
)

func (s *UpdateService) checkViaAtomFeed() (*models.UpdateInfo, error) {
    // 创建 10 秒超时的 context
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    url := fmt.Sprintf("https://github.com/%s/releases.atom", s.repo)
    logger.WithField("url", url).Info("尝试通过 Atom feed 获取版本")

    fp := gofeed.NewParser()
    fp.Client = s.httpClient // 复用已有的 http.Client

    feed, err := fp.ParseURLWithContext(url, ctx)
    if err != nil {
        return nil, fmt.Errorf("解析 Atom feed 失败: %w", err)
    }

    if len(feed.Items) == 0 {
        return nil, fmt.Errorf("Atom feed 无内容")
    }

    // 取第一个 item（最新 release）
    latestItem := feed.Items[0]

    logger.WithFields(map[string]interface{}{
        "title": latestItem.Title,
        "published": latestItem.PublishedParsed,
    }).Info("Atom feed 解析成功")

    // 后续处理...
    return s.convertFeedItemToUpdateInfo(latestItem)
}
```

### GitHub API 速率限制检测

```go
// Source: GitHub REST API 文档 + 现有实现
func (s *UpdateService) isRateLimitError(err error) bool {
    if err == nil {
        return false
    }

    errStr := err.Error()

    // 检查 403 或 "rate limit" 关键字
    return strings.Contains(errStr, "403") ||
           strings.Contains(errStr, "rate limit") ||
           strings.Contains(errStr, "API rate limit exceeded")
}

// 使用示例
func (s *UpdateService) CheckForUpdates() (*models.UpdateInfo, error) {
    info, err := s.checkViaAPI()
    if err != nil {
        if s.isRateLimitError(err) {
            // 速率限制，尝试返回缓存
            if s.cachedResult != nil {
                logger.Warn("遇到速率限制错误，返回缓存结果")
                return s.cachedResult, nil
            }

            // 降级到 Atom feed
            return s.checkViaAtomFeed()
        }
        return nil, err
    }
    return info, nil
}
```

### 前端 Toast 提示（Vue3 Composition API）

```typescript
// Source: 基于 Vue3 2026 最佳实践
import { ref } from 'vue'

// 简单的 toast 实现（或使用 vue-toastification 等库）
const toastVisible = ref(false)
const toastMessage = ref('')
const toastType = ref<'info' | 'success' | 'error'>('info')

function showToast(options: {
  message: string
  type: 'info' | 'success' | 'error'
  duration?: number
}) {
  toastMessage.value = options.message
  toastType.value = options.type
  toastVisible.value = true

  // 自动隐藏
  if (options.duration !== 0) {
    const duration = options.duration || 3000
    setTimeout(() => {
      toastVisible.value = false
    }, duration)
  }
}

// 在 checkForUpdates 中使用
async function checkForUpdates() {
  isChecking.value = true
  showToast({ message: '正在检查更新...', type: 'info', duration: 2000 })

  try {
    const info = await CheckForUpdates()
    showToast({ message: '检查完成', type: 'success' })
    return info
  } catch (error) {
    const message = formatUpdateError(error)
    showToast({ message, type: 'error', duration: 5000 })
    throw error
  } finally {
    isChecking.value = false
  }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 纯 GitHub API | GitHub API + Atom feed 降级 | 2026-02-08 (本研究) | 提高可用性，避免速率限制 |
| 无版本展示 | 卡片式版本信息展示 | 2026-02-08 (本研究) | 用户体验提升，信息透明 |
| 静默失败 | 清晰错误提示 + toast 反馈 | 2026-02-08 (本研究) | 用户明确知道发生了什么 |
| 阻塞式检查 | 异步检查 + 加载状态 | 2026-02-08 (本研究) | UI 不卡顿，体验流畅 |

**Deprecated/outdated:**
- **GitHub Atom feeds 作为主要数据源**: GitHub 官方明确标注 "undocumented and unsupported"，应仅作降级方案
- **未认证频繁 API 调用**: 60 次/小时限制容易触发，必须使用缓存
- **技术性错误消息**: 用户不理解 "403 Forbidden"，需要友好提示

## Open Questions

1. **Atom feed 长期可靠性**
   - **What we know**: GitHub 官方标注 Atom feeds 为 "undocumented and unsupported"，2025 年 1 月将数据保留期从 90 天缩短到 30 天
   - **What's unclear**: GitHub 是否会在 2026 年完全移除 Atom feeds
   - **Recommendation**: 仅作为降级方案，监控 GitHub Changelog，随时准备切换到纯 API（需认证）或自建代理服务

2. **是否需要 GitHub Personal Access Token**
   - **What we know**: 认证请求限制从 60 次/小时提升到 5,000 次/小时
   - **What's unclear**: 用户是否愿意提供 token（隐私和便利性权衡）
   - **Recommendation**: 暂不实现，仅在用户频繁触发速率限制时考虑添加可选配置

3. **预发布版本是否默认提示**
   - **What we know**: 现有实现已正确识别 `IsPrerelease`，UI 可区分显示
   - **What's unclear**: 用户是否希望默认接收预发布版本通知
   - **Recommendation**: 默认提示预发布版本，添加用户偏好设置（"接收预发布版本更新"开关）

## Sources

### Primary (HIGH confidence)

- **[/mmcdole/gofeed](https://github.com/mmcdole/gofeed)** - RSS/Atom feed 解析库，Context7 查询了 parse timeout、custom HTTP client、context support 等主题
- **[GitHub REST API - Rate Limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)** - 官方文档，确认未认证 60 次/小时限制
- **[GitHub REST API - Releases](https://docs.github.com/en/rest/releases/releases)** - 官方 API 文档，确认 `/repos/{owner}/{repo}/releases` 端点
- **[现有代码库]** - `pkg/service/update_service.go`, `pkg/version/version.go`, `frontend/src/stores/updateStore.ts`, `frontend/src/components/AboutDialog.vue` - 实际实现参考

### Secondary (MEDIUM confidence)

- **[creativeprojects/go-selfupdate](https://github.com/creativeprojects/go-selfupdate)** (WebSearch) - Go 应用自动更新库，验证了外部更新器模式是标准做法
- **[2026年Vue3生态插件推荐对比指南](https://blog.csdn.net/Rysxt_/article/details/156677180)** (WebSearch, 2026-01-08) - 确认 Vue3 + TypeScript 是 2026 年默认标准
- **[Vue Best Practices: A Practical Guide](https://cloudinary.com/guides/web-performance/vue-best-practices)** (WebSearch, 2025-12-19) - 验证组件驱动架构和 `<script setup>` 是主流

### Tertiary (LOW confidence)

- **[How to get the rss feed of github release](https://stackoverflow.com/questions/53988462)** (WebSearch) - 社区讨论，确认 Atom feed URL 格式（`/releases.atom`），但需警惕可靠性
- **[Joplin Desktop updater rate limit issue](https://github.com/laurent22/joplin/issues/14079)** (WebSearch, 2026-01-11) - 实际案例，验证速率限制是真实问题

## Metadata

**Confidence breakdown:**
- **Standard stack**: HIGH - 所有库均有官方文档或 Context7 验证，gofeed、semver、Wails 均为项目已有依赖
- **Architecture**: HIGH - 基于现有代码库分析，混合降级策略有实际案例（go-selfupdate）支持
- **Pitfalls**: HIGH - GitHub API 速率限制有官方文档确认，Atom feed 不可靠有 GitHub 官方声明

**Research date:** 2026-02-08
**Valid until:** 2026-03-10 (30 days - GitHub API 政策稳定，但 Atom feed 可能随时变化)

**Researcher notes:**
- 现有 `UpdateService` 实现已相当完善，主要缺失是错误处理和 UI 反馈
- 建议优先实现混合降级策略，避免重写现有逻辑
- Atom feed 虽然不可靠，但作为降级方案可显著提升用户体验
- 前端 Vue3 + TypeScript 生态成熟，无需引入新依赖
