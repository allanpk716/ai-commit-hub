# Requirements: AI Commit Hub v1.0.1

**Defined:** 2026-02-08
**Core Value:** 简化 Git 工作流 - 自动生成高质量 commit 消息

## v1.0.1 Requirements

v1.0.1 里程碑聚焦于修复日志系统和自动更新检测的关键 bug。

### 日志系统

- [ ] **LOG-01**: 日志输出格式符合 WQGroup/logger 标准格式
  - 格式：`2025-12-18 18:32:07.379 - [INFO]: 消息内容`
  - 使用 `withField` 格式器（默认）
  - 时间戳精确到毫秒

- [ ] **LOG-02**: 日志文件输出到程序根目录的 logs 文件夹
  - 日志根目录：`{程序根目录}/logs/`
  - 支持日志轮转（时间 + 大小）
  - 自动清理过期日志文件（默认 30 天）

- [ ] **LOG-03**: 所有 logger 调用使用正确的方法签名
  - 区分 `logger.Error()` 和 `logger.Errorf()`
  - 不带格式化参数使用 `logger.Error(msg)`
  - 带格式化参数使用 `logger.Errorf(fmt, args...)`

### 自动更新

- [ ] **UPD-09**: 修复 GitHub Releases 版本检测失败问题
  - 调查并修复版本比较逻辑
  - 确保正确处理预发布版本（alpha/beta/rc）
  - 修复 GitHub API 调用错误处理

- [ ] **UPD-10**: 增强更新检测错误提示
  - 网络错误时显示清晰的错误消息
  - API 速率限制时提供重试建议
  - 版本解析失败时记录详细日志

## Out of Scope

明确不在本次修复范围内的功能：

| Feature | Reason |
|---------|--------|
| 日志库替换 | WQGroup/logger 已满足需求，无需更换 |
| 新增日志功能 | 本次聚焦修复现有问题，不添加新功能 |
| 自动更新功能增强 | 本次仅修复 bug，不添加新特性 |

## Traceability

需求到阶段的映射：

| Requirement | Phase | Status |
|-------------|-------|--------|
| LOG-01 | Phase 6 | Pending |
| LOG-02 | Phase 6 | Pending |
| LOG-03 | Phase 6 | Pending |
| UPD-09 | Phase 7 | Pending |
| UPD-10 | Phase 7 | Pending |

**Coverage:**
- v1.0.1 requirements: 5 total
- Mapped to phases: 5 (100%) ✅
- Unmapped: 0

**Phase Distribution:**
- Phase 6 (日志系统修复): 3 requirements (LOG-01, LOG-02, LOG-03)
- Phase 7 (自动更新检测修复): 2 requirements (UPD-09, UPD-10)

---
*Requirements defined: 2026-02-08*
*Last updated: 2026-02-08 after roadmap creation*
