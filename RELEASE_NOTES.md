# Release Notes - v1.0.0

**Release Date**: 2026-02-07
**Version**: v1.0.0
**Status**: Stable Release 🎉

## 🎉 里程碑版本

AI Commit Hub v1.0.0 是项目的第一个主要稳定版本。本版本聚焦于构建完整可靠的桌面应用基础设施，从 CI/CD 自动化到单实例管理，从系统托盘修复到自动更新系统，全面提升了应用的质量和用户体验。

---

## ✨ 主要功能

### 1. 🔧 CI/CD 自动化构建和发布
- ✅ GitHub Actions 工作流配置完成
- ✅ Wails 自动构建和版本注入
- ✅ 产物自动打包（exe + 文档 + 校验和）
- ✅ 自动发布到 GitHub Releases
- ✅ 支持预发布版本检测
- ✅ 构建时间从 15 秒优化到 10 秒

### 2. 🪟 单实例锁定和窗口管理
- ✅ 防止多实例运行，自动激活现有窗口
- ✅ 窗口位置和大小自动保存和恢复
- ✅ 跨进程窗口激活（使用 Wails Events）
- ✅ 窗口边界验证（防止超出屏幕）
- ✅ 支持 Windows 多用户配置

### 3. 🔔 系统托盘交互完善
- ✅ 双击托盘图标显示/隐藏窗口
- ✅ 右键菜单：显示/隐藏、检查更新、退出
- ✅ 使用 `sync.Once` 和 `atomic.Bool` 防止竞态条件
- ✅ 升级 `systray` 库到 v1.3.0
- ✅ 区分"最小化到托盘"和"退出应用"

### 4. 🔄 自动更新系统
- ✅ 启动时后台检查 GitHub Releases 最新版本
- ✅ 支持预发布版本（beta/alpha）检测
- ✅ 后台下载更新文件，支持断点续传
- ✅ Wails Events 流式显示下载进度
- ✅ 外部更新器程序（嵌入主程序，3.2MB）
- ✅ 更新文件备份和回滚机制
- ✅ 更新完成后自动重启应用
- ✅ 主界面"关于"对话框显示当前版本

### 5. 🐛 代码质量提升
- ✅ 修复所有编译错误（10 处 logger 调用错误）
- ✅ 修复所有测试失败（2 个测试）
- ✅ 删除重复函数和未使用导入
- ✅ 统一使用 `github.com/WQGroup/logger` 日志库
- ✅ 测试通过率 100%

---

## 📦 技术改进

### 架构优化
- **分层架构**: App → Service → Repository → Database
- **事件驱动**: Wails Events 实现前后端通信
- **状态管理**: Pinia stores + StatusCache 优化
- **外部更新器**: 突破 Windows 文件锁定限制

### 性能提升
- **预加载**: 应用启动时批量加载项目状态
- **缓存策略**: StatusCache TTL 30 秒
- **并发获取**: Git 状态、Pushover 状态并发加载
- **乐观更新**: 用户操作后立即更新 UI

### 用户体验
- **版本显示**: 主界面和关于对话框显示当前版本
- **更新提示**: 实时下载进度和安装确认
- **窗口状态**: 位置和大小自动恢复
- **托盘交互**: 双击显示/隐藏，右键菜单完整

---

## 🔧 修复的问题

### 编译错误修复 (10 处)
1. `app.go:1234` - logger.Errorf 改为 logger.Error（非常量格式字符串）
2. `app.go:1658` - logger.Warn 改为 logger.Warnf（格式化参数）
3. `commit_service.go` - 批量修复 8 处 logger.Errorf 格式字符串问题

### 测试失败修复 (2 个)
1. `update_service_test.go::TestFindPlatformAsset` - 修复测试数据平台名称格式
2. `pushover_update_test.go::TestPushoverNotificationModes` - 修复测试文件路径错误

### 代码质量修复
1. 删除 `error_service_test.go` 中重复的 `TestFrontendError_JSON` 函数
2. 删除 `error_service_test.go` 中未使用的 `logger` 导入

---

## 📊 质量指标

### 代码统计
- **总阶段数**: 5 个
- **总计划数**: 14 个
- **完成进度**: 14/14 (100%)
- **持续时间**: 2 天

### 提交统计
- **总提交数**: 27+ commits
- **代码修改**: 2000+ lines
- **新增文件**: 20+ files
- **测试覆盖**: 0 failures

### 质量指标
- **编译错误**: 0 ✅
- **测试失败**: 0 ✅
- **代码重复**: 0 ✅
- **未使用导入**: 0 ✅

---

## 🚀 安装和升级

### 新用户安装
1. 下载 `ai-commit-hub-windows-amd64-v1.0.0.zip`
2. 解压缩到任意目录
3. 运行 `ai-commit-hub.exe`
4. 首次运行会自动创建配置目录

### 从旧版本升级
1. **自动升级**（推荐）:
   - 启动应用时会自动检测到新版本
   - 点击"立即更新"按钮
   - 等待下载完成（约 40MB）
   - 确认安装并重启应用

2. **手动升级**:
   - 下载最新版本压缩包
   - 关闭当前运行的应用
   - 解压缩并覆盖旧文件
   - 启动新版本

### 验证安装
- 主界面左上角显示当前版本号
- 点击"关于"按钮查看详细版本信息
- 格式：`v1.0.0 (commit-sha date-time)`

---

## 📝 已知限制

### 当前限制
1. **平台支持**: 仅支持 Windows（macOS/Linux 待后续版本）
2. **AI Provider**: 需要用户自行配置 API Key
3. **更新方式**: Windows 需要外部更新器程序

### 未来改进方向
1. **跨平台支持**: macOS 和 Linux 客户端
2. **内置 AI**: 本地 LLM 支持（Ollama）
3. **多语言**: 国际化支持（i18n）
4. **主题系统**: 深色模式、自定义主题

---

## 🙏 致谢

**开发**: Claude Code (Anthropic Sonnet 4.5)
**项目管理**: GSD (Get Shit Done) Workflow
**协作模式**: AI 驱动的敏捷开发

感谢所有参与测试和反馈的用户！

---

## 📚 相关文档

- **GitHub Releases**: https://github.com/allanpk716/ai-commit-hub/releases
- **项目文档**: https://github.com/allanpk716/ai-commit-hub
- **问题反馈**: https://github.com/allanpk716/ai-commit-hub/issues

---

## 📅 版本历史

### v1.0.0 (2026-02-07) - Stable Release
- ✅ 完成 CI/CD Pipeline
- ✅ 完成单实例和窗口管理
- ✅ 完成系统托盘修复
- ✅ 完成自动更新系统
- ✅ 完成代码质量提升

### v0.2.0-beta.1 (Previous Release)
- Beta 版本，用于测试自动更新系统

---

**下载链接**: [GitHub Releases](https://github.com/allanpk716/ai-commit-hub/releases/tag/v1.0.0)

**校验和**: 发布时会自动生成 SHA256 校验和文件

**签名**: Windows 可执行文件未签名（可能触发 SmartScreen 警告）

---

🎉 **感谢使用 AI Commit Hub！**
