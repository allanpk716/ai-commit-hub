# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

AI Commit Hub 是一个基于 Wails (Go + Vue3) 的桌面应用，用于为多个 Git 项目自动生成 AI 驱动的 commit 消息。

## 开发环境

- **操作系统**: Windows
- **后端**: Go 1.21+ + Wails v2
- **前端**: Vue 3 + TypeScript + Vite + Pinia
- **数据库**: SQLite + GORM

## 常用命令

### 开发命令
```bash
# 启动开发服务器
wails dev

# 生产构建
wails build

# 仅构建 Go 后端（绕过 Wails 绑定问题）
go build -o build/bin/ai-commit-hub.exe .

# 前端开发
cd frontend && npm run dev
```

## 功能特性

### Commit 生成和提交流程

AI Commit Hub 支持完整的 Git 工作流：

1. **生成 Commit 消息**: AI 根据暂存区更改自动生成规范的 commit 消息
2. **提交到本地**: 将生成的消息提交到本地 Git 仓库
3. **推送到远程**: 一键推送更改到远程仓库

### 推送功能说明

**使用方式：**
- 推送按钮只在本地提交成功后可用
- 推送到当前分支的同名远程分支
- 推送成功后按钮自动禁用，避免重复推送
- 切换项目或刷新状态时重置推送按钮

**错误处理：**
- 无远程仓库：显示"未配置远程仓库"错误
- 认证失败：显示"认证失败"错误
- 推送冲突：显示冲突提示，需手动解决
- 网络错误：显示具体的网络错误信息

### Pushover Hook 重装功能

**功能说明：**
- 支持为每个项目重装 Pushover Hook，保留用户的通知配置
- 只在 Hook 已安装且已是最新版本时显示"重装 Hook"按钮
- 重装会保留用户的通知配置（`.no-pushover`、`.no-windows`）
- 使用 `install.py --reinstall` 参数执行重装

**使用方式：**
1. 打开项目列表，找到 Pushover Hook 状态行
2. 确认显示"重装 Hook"按钮（已是最新版本时）
3. 点击"重装 Hook"按钮
4. 在确认对话框中点击"确定重装"
5. 等待重装完成，配置会自动保留

**注意事项：**
- 重装前会保存当前的通知配置
- 重装后自动恢复配置，无需手动设置
- 如果配置恢复失败，会记录警告但不影响重装结果

### 系统托盘功能

**功能说明：**
- 支持将应用最小化到系统托盘，后台运行
- 关闭窗口时应用不退出，继续驻留在托盘
- 通过托盘菜单可以恢复窗口或完全退出应用

**使用方式：**
1. **隐藏到托盘**: 点击窗口关闭按钮 (X)
2. **恢复窗口**: 右键点击托盘图标 → "显示窗口"
3. **退出应用**: 右键点击托盘图标 → "退出应用"

**注意事项：**
- 首次使用时会在关闭窗口后显示提示信息
- 应用启动时默认显示主窗口
- 托盘图标使用应用图标 (app-icon.png)

**技术实现：**
- 使用 `github.com/getlantern/systray` 库
- Wails `OnBeforeClose` 钩子拦截窗口关闭
- `sync.Once` 确保安全退出

### 测试命令
```bash
# Go 后端测试
go test ./... -v

# 运行单个包测试
go test ./pkg/repository -v

# 前端测试
cd frontend && npm run test        # 交互式测试
cd frontend && npm run test:run    # 单次运行
cd frontend && npm run test:ui     # UI 模式
```

### 依赖管理
```bash
# Go 依赖
go mod tidy

# 前端依赖
cd frontend && npm install
```

## 代码架构

### 分层架构

项目采用经典的分层架构，后端 Go 代码与前端 Vue 代码通过 Wails 绑定进行通信：

```
Frontend (Vue3)  ←→  Wails Bindings  ←→  Backend (Go)
     ↓                                            ↓
  Pinia Stores                          App API Methods
     ↓                                            ↓
  Components                            Service Layer
                                               ↓
                                          Repository Layer
                                               ↓
                                          SQLite (GORM)
```

### 后端架构 (Go)

**App 层 (`app.go`)**: Wails 应用的入口，包含所有导出给前端的 API 方法
- 13 个公开方法：项目管理、Commit 生成、Git 操作、历史记录
- 持有 `context.Context` 用于调用 Wails runtime 方法
- 初始化时创建数据库连接和所有 Repository

**Service 层 (`pkg/service/`)**: 业务逻辑层
- `ConfigService`: AI Provider 配置管理
- `CommitService`: Commit 消息生成逻辑，使用 Wails Events 实现流式输出

**Repository 层 (`pkg/repository/`)**: 数据访问层
- `GitProjectRepository`: Git 项目 CRUD 操作
- `CommitHistoryRepository`: Commit 历史记录操作
- `db.go`: 数据库初始化和连接管理

**AI 集成 (`pkg/ai/`, `pkg/provider/`)**: AI Provider 抽象层
- `ai.AIClient` 接口定义（支持流式和非流式）
- Provider Registry 模式：动态注册和获取 AI Provider
- 支持 OpenAI、Anthropic、DeepSeek、Ollama、Google、Phind

**Git 操作 (`pkg/git/`)**: Git 命令封装
- `status.go`: 获取暂存区状态、当前分支
- `git.go`: 执行 git commit、获取 diff

**配置管理 (`pkg/config/`)**: YAML 配置文件解析
- Provider 配置、语言设置、自定义 Prompt 模板

### 前端架构 (Vue3)

**主应用 (`App.vue`)**: 布局容器
- 顶部工具栏（添加项目、设置按钮）
- 左右分栏内容区（项目列表 + Commit 面板）

**组件 (`components/`)**:
- `ProjectList.vue`: 可拖拽排序的项目列表，支持搜索过滤
- `CommitPanel.vue`: Commit 生成面板，显示项目状态、AI 设置、流式输出、历史记录

**状态管理 (`stores/`)**:
- `projectStore.ts`: 项目列表状态、CRUD 操作、排序
- `commitStore.ts`: Commit 生成状态、流式消息监听（Wails Events）
- `statusCache.ts`: 项目状态缓存管理，提供预加载、后台刷新、乐观更新等功能
- `pushoverStore.ts`: Pushover 扩展和配置管理
  - 扩展信息（下载状态、版本信息、更新检查）
  - Pushover 环境变量配置验证
  - 项目 Hook 状态从 StatusCache 读取（不重复存储）

**类型定义 (`types/index.ts`)**: TypeScript 类型与 Go 结构体同步
**类型定义 (`types/status.ts`)**: StatusCache 相关类型定义

### 程序启动流程

AI Commit Hub 采用后端预加载 + 前端缓存填充的启动策略，确保应用启动时项目状态立即可用，避免 UI 闪烁。

#### 后端启动流程（app.go:66-208）

**关键步骤：**

1. **数据库和配置初始化**：
   - 初始化 SQLite 数据库连接
   - 加载 AI Provider 配置
   - 初始化 Pushover 服务

2. **异步预加载项目状态**（核心改进）：
   - 在后台 goroutine 中执行，不阻塞主线程
   - 通过 `StartupService.Preload()` 批量获取所有项目状态
   - 包括：Git 状态、暂存区状态、Pushover Hook 状态、推送状态等

3. **发送启动完成事件**：
   - 成功时：通过 `startup-complete` 事件将预加载的状态数据传递给前端
   - 失败时：仍发送 `startup-complete` 事件（success=false），避免界面卡死

```go
// 后端启动示例
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    // 初始化服务...

    go func() {
        startupService := service.NewStartupService(ctx, a.gitProjectRepo, a.pushoverService)
        statuses, err := startupService.Preload()

        if err != nil {
            runtime.EventsEmit(ctx, "startup-complete", nil)
            return
        }

        // 发送预加载的状态数据到前端
        runtime.EventsEmit(ctx, "startup-complete", map[string]interface{}{
            "success": true,
            "statuses": statuses,
        })
    }()
}
```

#### 前端启动流程（App.vue & main.ts）

**关键步骤：**

1. **显示启动画面**（SplashScreen）：
   - 优先显示启动画面，提供视觉反馈
   - 防止用户看到状态未加载的界面

2. **监听启动完成事件**：
   - 监听后端 `startup-complete` 事件
   - 成功时：将预加载的状态数据填充到 StatusCache
   - 失败时：降级为懒加载模式，使用时再获取状态

3. **超时保护机制**（main.ts:13-28）：
   - 30 秒超时保护，强制隐藏启动画面
   - 避免后端预加载失败导致界面卡死

```typescript
// main.ts - 超时保护
const startupTimeout = setTimeout(() => {
  if (startupStore.isVisible) {
    console.warn('启动超时，强制隐藏启动画面')
    startupStore.complete()
  }
}, 30000)

// App.vue - 监听后端事件
EventsOn('startup-complete', async (data: { success?: boolean; statuses?: Record<string, any> } | null) => {
  if (data?.success && data?.statuses) {
    const statusCache = useStatusCache()
    // 填充预加载的状态数据到缓存
    for (const [path, status] of Object.entries(data.statuses)) {
      statusCache.updateCache(path, status)
    }
  }
  // 隐藏启动画面
  showSplash.value = false
})
```

#### 启动流程设计理念

- **性能优化**：后端批量预加载减少网络请求次数
- **容错设计**：预加载失败自动降级为懒加载模式
- **用户体验**：启动画面 + 立即可用的状态，避免 UI 闪烁
- **超时保护**：防止后端预加载卡死导致界面无法使用

### 项目状态更新机制

StatusCache 实现了完整的状态管理生命周期，包括缓存、更新、乐观更新和后台刷新。

#### 状态缓存策略

**缓存结构**（`ProjectStatusCache`）：

```typescript
interface ProjectStatusCache {
  gitStatus: ProjectStatus | null        // Git 状态（分支、提交信息等）
  stagingStatus: StagingStatus | null    // 暂存区状态（已暂存文件）
  untrackedCount: number                 // 未跟踪文件数量
  pushoverStatus: HookStatus | null      // Pushover Hook 状态
  pushStatus: PushStatus | null          // 推送状态
  lastUpdated: number                    // 最后更新时间（时间戳）
  loading: boolean                       // 是否正在加载
  error: string | null                   // 错误信息
  stale: boolean                         // 是否已过期
}
```

**缓存过期判断**：

- 默认 TTL（Time To Live）：30 秒
- 通过 `isExpired(path)` 判断缓存是否过期
- 过期的缓存在后台静默刷新，不影响当前显示

#### 乐观更新机制（Optimistic Updates）

**核心方法**：`updateOptimistic(path, updates)`（statusCache.ts:170-195）

**工作流程**：

1. **立即更新 UI**：Git 操作（提交、暂存等）后立即更新缓存，无需等待后端确认
2. **保存回滚快照**：保存更新前的状态，用于失败时回滚
3. **返回回滚函数**：如果操作失败，调用回滚函数恢复原状态

```typescript
// 示例：Git 提交后的乐观更新
const rollback = statusCache.updateOptimistic(projectPath, {
  hasUncommittedChanges: false,
  untrackedCount: updatedUntrackedCount
})

try {
  // 执行 Git 提交
  await CommitProject(projectPath, commitMessage)
  // 提交成功，触发后台刷新以获取最新状态
  await statusCache.refresh(projectPath, { force: true })
} catch (error) {
  // 提交失败，回滚状态
  rollback?.()
  throw error
}
```

**使用场景**：
- Git 提交后立即更新提交状态
- 文件暂存后更新暂存区状态
- 推送操作后更新推送状态

#### 后台刷新机制（Background Refresh）

**核心方法**：`refresh(path, options)`（statusCache.ts:342-407）

**工作流程**：

1. **防重复请求**：如果已有相同请求在进行中，跳过本次刷新
2. **TTL 检查**：如果未强制刷新且缓存未过期，跳过刷新
3. **并发获取状态**：同时获取 Git 状态、暂存区状态、Pushover 状态等
4. **更新缓存**：将获取的最新状态更新到缓存中

```typescript
// 后台刷新示例
await statusCache.refresh(projectPath, {
  force: false,    // 是否强制刷新（忽略 TTL）
  silent: true     // 是否静默刷新（不显示加载状态）
})
```

**刷新策略**：
- **缓存优先**：优先使用缓存数据提供快速响应
- **后台更新**：缓存过期后在后台静默刷新，不影响 UI
- **强制刷新**：用户操作（如手动刷新按钮）后强制刷新，忽略 TTL

#### Git 操作后的状态同步

**操作流程**：

1. **乐观更新**：操作前立即更新 UI，提供即时反馈
2. **执行操作**：调用后端 API 执行 Git 操作
3. **强制刷新**：操作成功后强制刷新状态，确保数据一致性
4. **错误回滚**：操作失败时回滚到操作前状态

```typescript
// 完整示例：Git 提交操作
async function commitProject(projectPath: string, message: string) {
  const statusCache = useStatusCache()

  // 1. 乐观更新：立即更新 UI
  const rollback = statusCache.updateOptimistic(projectPath, {
    hasUncommittedChanges: false,
    lastCommitTime: Date.now()
  })

  try {
    // 2. 执行 Git 提交
    await CommitProject(projectPath, message)

    // 3. 强制刷新：获取最新状态
    await statusCache.refresh(projectPath, { force: true })

  } catch (error) {
    // 4. 错误回滚：恢复原状态
    rollback?.()
    throw error
  }
}
```

### StatusCache 层

`frontend/src/stores/statusCache.ts` 是状态缓存层，用于优化项目状态的加载和更新性能。

**核心功能：**

1. **预加载（Preload）**: 应用启动时批量加载所有项目状态，避免 UI 闪烁
2. **缓存优先（Cache-First）**: 切换项目时立即返回缓存数据，提供快速响应
3. **后台刷新（Background Refresh）**: 静默更新过期缓存以保持数据新鲜度
4. **乐观更新（Optimistic Updates）**: 用户操作后立即更新 UI，异步验证结果
5. **错误恢复（Error Recovery）**: 失败时使用过期缓存或显示友好错误提示

**统一状态管理：**

StatusCache 是项目状态的唯一数据源，管理以下内容：
- Git 状态（分支、提交信息等）
- 暂存区状态（已暂存文件）
- 未跟踪文件数量
- **Pushover Hook 状态**（是否安装、版本信息等）

**使用方法：**

```typescript
import { useStatusCache } from '@/stores/statusCache'

const statusCache = useStatusCache()

// 获取缓存状态（立即返回，无等待）
const status = statusCache.getStatus(projectPath)

// 获取 Pushover 状态
const pushoverStatus = statusCache.getPushoverStatus(projectPath)

// 刷新状态（如果缓存未过期可能跳过）
await statusCache.refresh(projectPath)

// 强制刷新（忽略 TTL）
await statusCache.refresh(projectPath, { force: true })

// 批量预加载
await statusCache.preload(projectPaths)
```

**缓存配置：**

```typescript
statusCache.updateOptions({
  ttl: 30000,              // 缓存过期时间（毫秒），默认 30 秒
  backgroundRefresh: true  // 是否在后台刷新过期缓存
})
```

**生命周期：**

1. 应用启动时调用 `statusCache.init()` 预加载所有项目
2. 用户切换项目时调用 `getStatusOrRefresh()` 获取状态
3. Git 操作后调用 `updateOptimistic()` 立即更新 UI
4. 后端通过 Wails Events 发送状态变更事件，自动使缓存失效

**测试：**

```bash
cd frontend
npm run test:run  # 运行单元测试
```

测试文件位于 `frontend/src/stores/__tests__/statusCache.spec.ts`

### Wails 事件流

流式 Commit 生成使用 Wails Events 实现：

```
Frontend          Backend
   |                 |
 generateCommit()   |
   |---------------->|
   |                 Create CommitService
   |                 Start goroutine
   |                 |
   |     commit-delta event (streaming)
   |<----------------|
   |     update UI
   |                 |
   |<----------------|  repeat...
   |                 |
   |     commit-complete event
   |<----------------|
```

## 开发规则

### 通用规则

1. **使用中文回答问题和编写文档**

2. **BAT 脚本规范**: BAT 脚本中不要使用中文（避免编码问题）

3. **临时文件位置**: 临时测试代码、数据统一放在项目根目录的 `tmp/` 文件夹中

4. **脚本修复原则**: 修复脚本时优先在原脚本上修改，非必需不要新建脚本

5. **图片处理**: 使用截图 MCP 或 Agent 能力前，确保图片尺寸小于 1000x1000

6. **计划文件**: 项目计划文件统一放在 `docs/plans/` 目录

7. **日志库使用**: 统一使用 `github.com/WQGroup/logger` 日志库
   - 基本日志级别：`Debug()`、`Info()`、`Warn()`、`Error()`
   - 格式化版本：`Debugf()`、`Infof()`、`Warnf()`、`Errorf()`
   - 支持多种格式器：JSON、文本、结构化日志
   - 支持 YAML 配置文件配置日志行为
   - 支持日志轮转（时间/大小）和自动清理
   - 线程安全，支持并发写入

### Wails 开发规范

1. **调试流程**:
   - 使用 `wails dev` 启动开发服务器
   - 使用浏览器技能（dev-browser）进行前后端通信和交互测试
   - 修改 Go 代码后需要重启 Wails
   - 修改前端代码支持热更新

2. **API 方法命名**: 导出的方法使用大写开头（Go 导出规则），会自动生成驼峰命名的 JavaScript 绑定

3. **错误处理**: 所有 API 方法应检查 `a.initError`，如果数据库初始化失败应返回错误

4. **Wails Events**: 流式输出使用 `runtime.EventsEmit()` 发送事件，前端使用 `EventsOn()` 监听

5. **控制台窗口隐藏**（Windows 平台）：

   **问题描述：**
   - 在 Windows 上执行外部命令（如 git）时，会出现控制台窗口闪烁
   - 影响用户体验，尤其是频繁执行命令时

   **解决方案：**
   - 使用自定义 `Command` 函数封装 `exec.Cmd`
   - 在 Windows 平台下设置 `CREATE_NO_WINDOW` 标志

   ```go
   // app.go:32-45
   import (
       "os/exec"
       "runtime" as stdruntime
       "golang.org/x/sys/windows"
   )

   // Command creates a new exec.Cmd with hidden window on Windows
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

   **使用示例：**

   ```go
   // 在所有外部命令执行时使用 Command 函数
   cmd := Command("git", "status", "--porcelain")
   cmd.Dir = projectPath
   output, err := cmd.CombinedOutput()
   ```

   **注意事项：**
   - 所有外部命令（git、python 等）都必须使用 `Command` 函数
   - 不能直接使用 `exec.Command`，否则会导致控制台窗口闪烁
   - Unix/Linux 平台不受影响，`CREATE_NO_WINDOW` 标志仅在 Windows 下生效

6. **日志输出规范**：
   - 统一使用 `github.com/WQGroup/logger` 日志库
   - 不要使用 `fmt.Printf` 或 `log.Println` 输出日志
   - 生产构建后的日志应输出到文件，避免控制台输出

   ```go
   import "github.com/WQGroup/logger"

   // 正确的日志输出
   logger.Info("AI Commit Hub starting up...")
   logger.Errorf("数据库初始化失败: %v", err)

   // 错误的日志输出（避免使用）
   fmt.Printf("应用启动: %s\n", version)
   log.Println("错误:", err)
   ```

### Git 提交规范

- 使用 Conventional Commits 格式
- 中文提交消息
- 示例: `feat: 添加项目拖拽排序功能`

## 配置文件位置

- **Windows**: `C:\Users\<username>\.ai-commit-hub\`
- **macOS/Linux**: `~/.ai-commit-hub/`

### 配置文件
- `config.yaml`: AI Provider 配置、语言设置
- `ai-commit-hub.db`: SQLite 数据库
- `prompts/`: 自定义 Prompt 模板目录

## 常见问题

### Wails 绑定生成错误

Windows 上可能出现 `wailsbindings.exe: %1 is not a valid Win32 application` 错误。

**解决方案**:
1. 删除临时目录下的 wbindings 文件
2. 重新运行 `wails dev`
3. 或使用已有的绑定文件，直接 `go build`

### 前后端类型同步

Go 结构体修改后需要重新生成 Wails 绑定。确保 `frontend/src/types/index.ts` 中的 TypeScript 类型与 Go 结构体保持一致。
