# Windows 控制台窗口隐藏实现

## 问题描述

在 Windows 平台上，当 Go 程序使用 `os/exec` 执行外部命令（如 git）时，会出现控制台窗口闪烁的问题。

### 症状

- 每次执行外部命令时，会短暂出现一个控制台窗口
- 频繁执行命令时（如循环执行 git 命令），闪烁现象非常明显
- 严重影响用户体验，尤其是桌面应用程序

### 根本原因

Windows 平台上的 `exec.Command` 默认会创建一个新的控制台窗口来显示外部命令的输出。即使在命令执行完成后窗口会自动关闭，但这个闪烁过程仍然可见。

## 解决方案

使用自定义 `Command` 函数封装 `exec.Cmd`，通过设置 Windows 进程创建标志来隐藏控制台窗口。

### 实现步骤

#### 1. 导入必要的包

```go
import (
    "os/exec"
    stdruntime "runtime"
    "golang.org/x/sys/windows"
)
```

#### 2. 创建自定义 Command 函数

在 `app.go` 中定义：

```go
// Command creates a new exec.Cmd with hidden window on Windows
// This prevents console windows from popping up when running external commands
func Command(name string, args ...string) *exec.Cmd {
    cmd := exec.Command(name, args...)

    // On Windows, hide the console window to prevent popups
    if stdruntime.GOOS == "windows" {
        cmd.SysProcAttr = &windows.SysProcAttr{
            CreationFlags: 0x08000000, // CREATE_NO_WINDOW
        }
    }

    return cmd
}
```

#### 3. 使用自定义 Command 函数

**在所有外部命令执行时使用 `Command` 函数：**

```go
// ✅ 正确：使用自定义 Command 函数
cmd := Command("git", "status", "--porcelain")
cmd.Dir = projectPath
output, err := cmd.CombinedOutput()
if err != nil {
    return fmt.Errorf("failed to get git status: %w", err)
}

// ❌ 错误：直接使用 exec.Command
cmd := exec.Command("git", "status", "--porcelain")  // 会导致控制台窗口闪烁
cmd.Dir = projectPath
output, err := cmd.CombinedOutput()
```

## 技术细节

### CREATE_NO_WINDOW 标志

`0x08000000` 是 Windows API 中的 `CREATE_NO_WINDOW` 标志值，它告诉系统在创建新进程时不要显示控制台窗口。

### 平台兼容性

这个实现只在 Windows 平台上生效：

```go
if stdruntime.GOOS == "windows" {
    cmd.SysProcAttr = &windows.SysProcAttr{
        CreationFlags: 0x08000000,
    }
}
```

在 Unix/Linux 平台上，这段代码不会执行，因此不会影响其他平台的正常运行。

## 注意事项

### 1. 必须使用自定义 Command 函数

所有外部命令（git、python、node 等）都必须使用 `Command` 函数，不能直接使用 `exec.Command`。

### 2. 检查所有 exec.Command 调用

使用以下命令搜索项目中所有使用 `exec.Command` 的地方：

```bash
# Windows PowerShell
Select-String -Path "*.go" -Pattern "exec\.Command"

# Unix/Linux
grep -rn "exec\.Command" --include="*.go"
```

### 3. 代码审查要点

在代码审查时，检查以下几点：

- ✅ 是否使用了 `Command()` 函数而不是 `exec.Command()`
- ✅ 是否正确导入 `golang.org/x/sys/windows` 包
- ✅ 是否设置了 `cmd.Dir` 工作目录（如果需要）

## 常见问题

### Q: 为什么不直接使用 `exec.Command`？

A: `exec.Command` 在 Windows 上默认会创建新的控制台窗口，导致窗口闪烁。自定义 `Command` 函数通过设置 `CREATE_NO_WINDOW` 标志来隐藏窗口。

### Q: 这个方法适用于所有外部命令吗？

A: 是的，适用于所有通过 `os/exec` 执行的外部命令，包括 git、python、node、系统命令等。

### Q: 在 Linux/macOS 上需要特殊处理吗？

A: 不需要。代码中的平台检查确保只在 Windows 上应用这个设置，其他平台不受影响。

### Q: 如何验证是否生效？

A: 在 Windows 上运行应用，观察执行外部命令时是否还有控制台窗口闪烁。如果实现正确，应该不会看到任何控制台窗口。

## 实际应用示例

### Git 命令执行

```go
// pkg/git/status.go
func GetStatus(projectPath string) ([]string, error) {
    cmd := Command("git", "status", "--porcelain")
    cmd.Dir = projectPath

    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("git status failed: %w", err)
    }

    // 处理输出
    lines := strings.Split(string(output), "\n")
    return lines, nil
}
```

### Python 脚本执行

```go
cmd := Command("python", "script.py", "--input", "data.txt")
cmd.Dir = scriptDir

output, err := cmd.CombinedOutput()
if err != nil {
    return fmt.Errorf("python script failed: %w", err)
}
```

## 相关文档

- Wails 开发规范：`docs/development/wails-development-standards.md`（包含完整的使用示例和最佳实践）
- CLAUDE.md：项目根目录（快速参考）

## 参考资料

- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows)
- [Windows Process Creation Flags](https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags)
- [os/exec package](https://pkg.go.dev/os/exec)
