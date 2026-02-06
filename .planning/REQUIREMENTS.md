# Requirements: AI Commit Hub

**Defined:** 2026-02-06
**Core Value:** 简化 Git 工作流 - 自动生成高质量 commit 消息

## v1 Requirements

功能修复和增强需求，涵盖单实例锁定、自动更新、系统托盘交互和 CI/CD 自动化。

**实施顺序：**
1. **CI/CD 流水线** - 先建立发布流程，确保能构建和发布应用
2. **单实例与窗口管理** - 基础功能，独立无依赖
3. **系统托盘交互** - 修复现有功能，升级依赖
4. **自动更新系统** - 依赖 CI/CD 发布的构建产物
5. **代码质量修复** - 持续进行，确保可维护性

### Single Instance (单实例与窗口管理)

- [ ] **SI-01**: 应用启动时检测是否已有实例运行
- [ ] **SI-02**: 检测到多实例时，激活现有窗口到前台
- [ ] **SI-03**: 使用 Wails 内置 SingleInstanceLock 机制
- [ ] **SI-04**: 保存和恢复窗口状态（位置、大小）

### Auto Update (自动更新系统)

- [ ] **UPD-01**: 应用启动时后台检查 GitHub Releases 最新版本
- [ ] **UPD-02**: 主界面提供"检查更新"按钮
- [ ] **UPD-03**: 设置页面显示当前版本和最新版本信息
- [ ] **UPD-04**: 后台下载更新文件（非阻塞）
- [ ] **UPD-05**: 通过 Wails Events 流式显示下载进度
- [ ] **UPD-06**: 使用外部更新器程序替换主应用（避免文件锁定）
- [ ] **UPD-07**: 更新完成后自动重启应用
- [ ] **UPD-08**: 支持 GitHub Releases 资源命名规范

### System Tray (系统托盘交互)

- [ ] **TRAY-01**: 双击托盘图标恢复并激活主界面
- [ ] **TRAY-02**: 右键显示上下文菜单（显示/隐藏、检查更新、退出）
- [ ] **TRAY-03**: 升级到 lutischan-ferenc/systray v1.3.0 以支持双击
- [ ] **TRAY-04**: 使用 sync.Once 和 atomic.Bool 防止竞态条件
- [ ] **TRAY-05**: 区分"最小化到托盘"和"退出应用"行为

### CI/CD (持续集成与发布)

- [x] **CI-01**: 配置 GitHub Actions 工作流（.github/workflows/build.yml） ✅
- [x] **CI-02**: Tag 推送时触发自动化构建 ✅
- [x] **CI-03**: 生成 Windows 平台可执行文件 ✅
- [x] **CI-04**: 自动上传构建资源到 GitHub Releases ✅
- [x] **CI-05**: 资源文件命名遵循平台检测规范（ai-commit-hub-windows-amd64-v{version}.zip） ✅

### Code Quality (代码质量修复)

- [ ] **Q-01**: 修复 app.go:969 logger.Errorf 使用错误（使用格式字符串而非变量）
- [ ] **Q-02**: 修复 error_service_test.go 重复测试函数声明
- [ ] **Q-03**: 修复 error_service_test.go 类型错误（FrontendError vs error）
- [ ] **Q-04**: 确保所有测试通过
- [ ] **Q-05**: 确保项目可以成功编译

## v2 Requirements

未来版本考虑的功能，当前不实现。

### Advanced Update Features

- **UPD-V2-01**: 支持增量更新（Squirrel.windows）
- **UPD-V2-02**: 静默自动更新（无用户确认）
- **UPD-V2-03**: 回滚机制（更新失败时恢复旧版本）

### Enhanced Tray Features

- **TRAY-V2-01**: 托盘图标显示未读通知徽章
- **TRAY-V2-02**: 自定义托盘菜单项
- **TRAY-V2-03**: 拖拽文件到托盘图标触发提交

### Security & Signing

- **SEC-V2-01**: 代码签名证书集成
- **SEC-V2-02**: 更新包签名验证
- **SEC-V2-03**: 防篡改检查

## Out of Scope

明确排除的功能，防止范围蔓延。

| Feature | Reason |
|---------|--------|
| 完整 Git GUI | 用户已有 TortoiseGit，本工具专注 AI commit 生成 |
| Merge/Rebase 操作 | 复杂操作交给专业 Git 工具 |
| 多语言支持 | 增加维护成本，目标用户为中文开发者 |
| 云端同步配置 | 单机工具，配置本地管理即可 |
| 团队协作功能 | 非目标用户场景 |
| 插件系统 | 增加复杂度，当前需求不明确 |
| 内置终端 | 功能与现有 Git 终端工具重复 |

## Traceability

需求到阶段的映射，在路线图创建时填充。

| Requirement | Phase | Status |
|-------------|-------|--------|
| CI-01 | Phase 1 | Complete ✅ |
| CI-02 | Phase 1 | Complete ✅ |
| CI-03 | Phase 1 | Complete ✅ |
| CI-04 | Phase 1 | Complete ✅ |
| CI-05 | Phase 1 | Complete ✅ |
| SI-01 | Phase 2 | Pending |
| SI-02 | Phase 2 | Pending |
| SI-03 | Phase 2 | Pending |
| SI-04 | Phase 2 | Pending |
| TRAY-01 | Phase 3 | Pending |
| TRAY-02 | Phase 3 | Pending |
| TRAY-03 | Phase 3 | Pending |
| TRAY-04 | Phase 3 | Pending |
| TRAY-05 | Phase 3 | Pending |
| UPD-01 | Phase 4 | Pending |
| UPD-02 | Phase 4 | Pending |
| UPD-03 | Phase 4 | Pending |
| UPD-04 | Phase 4 | Pending |
| UPD-05 | Phase 4 | Pending |
| UPD-06 | Phase 4 | Pending |
| UPD-07 | Phase 4 | Pending |
| UPD-08 | Phase 4 | Pending |
| Q-01 | Phase 5 | Pending |
| Q-02 | Phase 5 | Pending |
| Q-03 | Phase 5 | Pending |
| Q-04 | Phase 5 | Pending |
| Q-05 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 25 total
- Mapped to phases: 25
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-06*
*Last updated: 2026-02-06 after roadmap creation*
