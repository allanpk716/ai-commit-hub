# Phase 1: CI/CD Pipeline - Context

**Gathered:** 2026-02-06
**Status:** Ready for planning

<domain>
## Phase Boundary

建立自动化构建和发布流程，确保代码能够自动编译、测试并发布到 GitHub Releases。核心是构建 Windows 平台可执行文件并自动上传到 Releases。

</domain>

<decisions>
## Implementation Decisions

### 产物打包规范
- 发布包包含：exe 文件 + 基础文档
- 基础文档包括：README.md + 带注释的完整配置示例
- 同时生成 SHA256 和 MD5 校验和文件
- 配置示例包含：完整 config.yaml 示例 + 注释说明各配置项作用

### 触发机制
- 支持两种触发方式：版本标签触发 + 手动触发
- 版本标签必须使用严格的 `v` 前缀（如 v1.0.0）
- 预发布版本（包含 `-beta`、`-rc` 等）标记为 Pre-release，不出现在默认更新检查中
- 推送标签到 GitHub 时自动触发构建和发布流程

### 构建矩阵策略
- v1 版本仅构建 Windows 平台
- Windows 平台支持两个架构：amd64 (x86_64) + 386 (x86_32)
- 多个架构采用并行构建策略，最快完成全部产物
- 产物命名遵循平台检测规范：`ai-commit-hub-windows-{arch}-v{version}.zip`

### 构建元数据
- 在编译时嵌入完整的版本信息：版本号 + commit SHA + 构建时间戳
- 可通过 `-v` 或 `--version` 查看嵌入的版本信息
- v1 阶段暂不进行代码签名，用户需手动点击"更多信息"才能运行
- 构建日志仅在 GitHub Actions 界面查看，不单独保存为产物

### Claude's Discretion
- Go 编译优化级别的选择（-O vs -Osize）
- GitHub Actions runner 镜像选择（windows-latest 版本）
- 构建超时时间的设置
- 产物的压缩格式和级别

</decisions>

<specifics>
## Specific Ideas

- 希望构建流程稳定可靠，避免因网络问题或临时依赖问题导致构建失败
- 发布包应该让用户下载后能够快速上手，配置示例要清晰易懂
- 产物命名要明确标识平台和架构，避免用户下载错误版本

</specifics>

<deferred>
## Deferred Ideas

- macOS 和 Linux 平台支持 — v2 或后续版本考虑
- ARM64 架构支持 — 如果有实际需求再添加
- 代码签名 — 等项目成熟后再考虑
- 自动化测试集成 — 属于代码质量阶段（Phase 5）

</deferred>

---

*Phase: 01-ci-cd-pipeline*
*Context gathered: 2026-02-06*
