# Pushover Hook 版本比较问题修复

## 问题描述

### 现象
项目中显示的 Pushover Hook 版本格式为 `v1.6.0-1-g3871faa`，而扩展最新版本为 `v1.6.0`。系统误判认为项目需要更新，但实际上两者是同一版本。

### 根本原因

**Git describe 输出格式：**
```
v{major}.{minor}.{patch}-{num_commits}-g{commit_hash}
```

例如：
- `v1.6.0-1-g3871faa` 表示在 v1.6.0 tag 之后有 1 个提交，commit hash 为 3871faa
- `v1.6.0-15-gabc123` 表示在 v1.6.0 tag 之后有 15 个提交

**版本比较逻辑：**

1. **项目中的版本**（来自 VERSION 文件）：`v1.6.0-1-g3871faa`
2. **扩展最新版本**（使用 `git describe --tags --abbrev=0`）：`v1.6.0`
3. **比较过程**：
   - `CompareVersions` 将 `v1.6.0-1-g3871faa` 分割为主版本 `v1.6.0` 和预发布标签 `1-g3871faa`
   - 主版本号 `v1.6.0` = `v1.6.0`，相等
   - 比较预发布标签：`1-g3871faa` vs 空
   - 根据语义化版本规范：**没有预发布标签的版本更高**
   - 结论：`v1.6.0` > `v1.6.0-1-g3871faa`，误判需要更新

### 影响范围
- 所有使用 `git describe` 输出作为版本号的 Hook 安装
- 导致已安装最新版本的项目仍然显示"有更新可用"

## 解决方案

### 实现方式
在 `GetHookVersion` 函数中添加版本号清理逻辑，从 Git describe 输出中提取纯版本号。

### 代码修改

**文件：`pkg/pushover/status.go`**

1. 添加 `cleanVersion` 辅助函数：
```go
// cleanVersion 清理版本号，从 Git describe 输出中提取纯版本号
// 例如: "v1.6.0-1-g3871faa" -> "v1.6.0"
// 如果不是 Git describe 格式（如 "v1.6.0-alpha"），则保持不变
func cleanVersion(version string) string {
    // 匹配 Git describe 输出格式: v{major}.{minor}[.{patch}]-{num}-g{hash}
    re := regexp.MustCompile(`^v?(\d+\.\d+(?:\.\d+)?)-\d+-g[0-9a-f]+$`)
    if re.MatchString(version) {
        // 提取纯版本号部分
        parts := strings.Split(version, "-")
        if len(parts) >= 1 {
            return parts[0]
        }
    }
    return version
}
```

2. 在 `GetHookVersion` 中调用清理函数：
```go
// 清理版本号，移除 Git 提交信息
version = cleanVersion(version)
```

### 正则表达式说明

```regex
^v?(\d+\.\d+(?:\.\d+)?)-\d+-g[0-9a-f]+$
```

- `^v?` - 可选的 v 前缀
- `(\d+\.\d+)` - major.minor（必需）
- `(?:\.\d+)?` - 可选的 .patch 部分
- `-` - 分隔符
- `\d+` - 提交数量（数字）
- `-` - 分隔符
- `g[0-9a-f]+` - commit hash（g 前缀 + 十六进制）

### 测试用例

**单元测试：`TestCleanVersion`**
- ✅ `v1.6.0-1-g3871faa` → `v1.6.0`
- ✅ `v2.0.0-15-gabc123` → `v2.0.0`
- ✅ `v1.6.0` → `v1.6.0`（纯版本号不变）
- ✅ `v1.6.0-alpha` → `v1.6.0-alpha`（真正的预发布版本不变）

**集成测试：`TestCleanVersionAndCompare`**
- ✅ `v1.6.0-1-g3871faa` vs `v1.6.0` → 不需要更新
- ✅ `v1.5.0-3-gabc123` vs `v1.6.0` → 需要更新
- ✅ `v1.7.0-1-gdef456` vs `v1.6.0` → 不需要更新

## 验证方法

### 手动测试
1. 安装一个项目 Hook，版本为 `v1.6.0-1-g3871faa`
2. 检查扩展版本为 `v1.6.0`
3. 验证项目状态显示为"已是最新版本"，而不是"有更新可用"

### 自动化测试
```bash
cd pkg/pushover
go test -v -run TestCleanVersion
go test -v -run TestCleanVersionAndCompare
```

## 相关文件
- `pkg/pushover/status.go` - 版本清理逻辑
- `pkg/pushover/version_test.go` - 测试用例
- `pkg/pushover/version.go` - 版本比较函数

## 注意事项
- ✅ 不影响真正的预发布版本（如 `v1.6.0-alpha`）
- ✅ 向后兼容旧的版本格式
- ✅ 所有测试通过
- ✅ 不需要修改 VERSION 文件格式

## 修复日期
2026-02-01
