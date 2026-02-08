# Phase 7: 自动更新检测修复 - Context

**Gathered:** 2026-02-08
**Status:** Ready for planning

## Phase Boundary

修复 GitHub Releases 版本检测失败问题,增强错误提示和重试机制。确保用户能够清晰看到版本信息和更新状态。

## Implementation Decisions

### 版本信息展示
- **卡片式布局**: 版本信息以卡片形式展示在关于界面
- **完整信息**: 卡片包含
  - 当前版本号
  - 最新版本号
  - 更新状态(已是最新/有新版本可用)
  - 最新版本发布时间
  - GitHub Releases 下载链接
  - 更新说明(changelog)

### 版本检测方案
- **混合降级策略**:
  - 优先使用 GitHub API (更快,结构化数据)
  - API 失败时降级到 RSS/Atom Feed (无需认证,更可靠)
  - 两种方式都失败时显示清晰的错误消息

### 用户交互
- **检查更新按钮反馈** (Claude's Discretion):
  - 点击时显示加载状态
  - 提供即时反馈(loading 图标 + toast 提示)
  - 1-2 秒内完成检测

- **界面刷新时机**:
  - 应用启动时自动检测版本
  - 用户点击"检查更新"按钮时手动刷新
  - 检测结果实时更新卡片内容

### 调试和验证
- **集成测试**: 使用真实网络请求测试
  - 测试场景:
    - GitHub API 正常响应
    - GitHub API 失败,降级到 RSS
    - 网络断开/超时
    - API 速率限制
  - 验证版本比较逻辑(预发布版本 alpha/beta/rc)
  - 确认错误消息清晰且可操作

### Claude's Discretion
- 检查更新按钮的具体交互细节(loading 样式、toast 文案、动画时长)
- API 超时时间和重试次数
- RSS feed 解析的容错处理
- 日志输出的详细程度

## Specific Ideas

- **用户反馈的问题**:
  - 关于界面看不到版本信息(当前版本、最新版本)
  - 点击"检查更新"没有反馈
  - 无法判断检测是否成功

- **期望体验**:
  - 关于页面一目了然地看到版本状态
  - 检测更新时提供即时反馈
  - 即使 GitHub API 失败也能通过 RSS 获取版本

## Deferred Ideas

无 - 讨论保持在 Phase 7 范围内

---

*Phase: 07-auto-update-detection-fix*
*Context gathered: 2026-02-08*
