# 架构

**分析日期：** 2024-02-05

## 模式概述

**整体：** 分层架构（Layered Architecture） + 事件驱动（Event-Driven）

**关键特征：**
- 前后端分离架构（Wails 桥接）
- 后端分层设计（App/Service/Repository）
- 前端状态管理（Pinia Stores）
- 事件驱动的流式数据流
- 系统托盘驻留模式

## 层次

### 后端架构 (Go)

**App 层 (`app.go`)**: Wails 应用的入口，包含所有导出给前端的 API 方法
- 13 个公开方法：项目管理、Commit 生成、Git 操作、历史记录
- 持有 `context.Context` 用于调用 Wails runtime 方法
- 初始化时创建数据库连接和所有 Repository
- 系统托盘管理：双击/右键事件、窗口生命周期控制

**Service 层 (`pkg/service/`)**: 业务逻辑层
- `ConfigService`: AI Provider 配置管理
- `CommitService`: Commit 消息生成逻辑，使用 Wails Events 实现流式输出
- `ProjectConfigService`: 项目特定配置管理
- `StartupService`: 应用启动时的状态预加载服务
- `UpdateService`: 应用更新管理

**Repository 层 (`pkg/repository/`)**: 数据访问层
- `GitProjectRepository`: Git 项目 CRUD 操作
- `CommitHistoryRepository`: Commit 历史记录操作
- `db.go`: 数据库初始化和连接管理

**AI 集成层 (`pkg/ai/`, `pkg/provider/`)**: AI Provider 抽象层
- `ai.AIClient` 接口定义（支持流式和非流式）
- Provider Registry 模式：动态注册和获取 AI Provider
- 支持 OpenAI、Anthropic、DeepSeek、Ollama、Google、Phind

**Git 操作层 (`pkg/git/`)**: Git 命令封装
- `status.go`: 获取暂存区状态、当前分支
- `git.go`: 执行 git commit、获取 diff

### 前端架构 (Vue3)

**主应用层 (`App.vue`)**: 布局容器和入口点
- 顶部工具栏（添加项目、设置按钮）
- 左右分栏内容区（项目列表 + Commit 面板）
- 启动画面和生命周期管理

**组件层 (`components/`)**: UI 组件
- `ProjectList.vue`: 可拖拽排序的项目列表，支持搜索过滤
- `CommitPanel.vue`: Commit 生成面板，显示项目状态、AI 设置、流式输出
- 状态相关组件：`ProjectStatusHeader.vue`、`StagedList.vue`、`UnstagedList.vue`
- 对话框组件：`SettingsDialog.vue`、`ConfirmDialog.vue`

**状态管理层 (`stores/`)**: Pinia 状态管理
- `projectStore.ts`: 项目列表状态、CRUD 操作、排序
- `commitStore.ts`: Commit 生成状态、流式消息监听（Wails Events）
- `statusCache.ts`: 项目状态缓存管理，提供预加载、后台刷新、乐观更新等功能
- `pushoverStore.ts`: Pushover 扩展和配置管理

**类型定义层 (`types/`)**: TypeScript 类型与 Go 结构体同步
- `index.ts`: 基础类型定义
- `status.ts`: StatusCache 相关类型定义

## 数据流

### 启动流程

```
main.go (Wails 启动)
    ↓
app.startup() (初始化服务和数据库)
    ↓
StartupService.Preload() (后台预加载项目状态)
    ↓
startup-complete 事件 (发送预加载数据到前端)
    ↓
App.vue 接收事件，填充 StatusCache
    ↓
隐藏启动画面，显示主界面
```

### Git 操作流程

```
用户操作 (点击提交按钮)
    ↓
乐观更新 StatusCache (立即更新 UI)
    ↓
调用 backend API (CommitProject)
    ↓
执行 git commit 命令
    ↓
强制刷新 StatusCache (获取最新状态)
    ↓
错误时回滚到之前状态
```

### 流式 Commit 生成

```
前端调用 generateCommit()
    ↓
backend 创建 CommitService
    ↓
启动 goroutine 发送 commit-delta 事件
    ↓
前端监听事件，更新 UI
    ↓
发送 commit-complete 事件结束
```

## 关键抽象

### AI Provider 抽象

```go
// pkg/ai/ai.go
type AIClient interface {
    GenerateCommitMessage(ctx context.Context, diff string, messages []string) (string, error)
    GenerateCommitMessageStream(ctx context.Context, diff string, messages []string) (<-chan string, error)
}
```

**实现：**
- 每个 Provider 实现相同接口
- Registry 模式动态注册
- 支持流式和非流式调用

### StatusCache 抽象

```typescript
// frontend/src/stores/statusCache.ts
interface ProjectStatusCache {
    gitStatus: ProjectStatus | null        // Git 状态
    stagingStatus: StagingStatus | null    // 暂存区状态
    pushoverStatus: HookStatus | null      // Pushover Hook 状态
    pushStatus: PushStatus | null          // 推送状态
    lastUpdated: number                    // 最后更新时间
    loading: boolean                       // 加载状态
    stale: boolean                         // 是否过期
}
```

**功能：**
- 缓存优先策略
- 后台静默刷新
- 乐观更新机制
- TTL 过期管理

### 系统托盘抽象

```go
// app.go
type App struct {
    systrayReady   chan struct{} // 就绪信号
    systrayExit    *sync.Once    // 退出控制
    windowVisible  bool          // 窗口可见状态
    windowMutex    sync.RWMutex  // 状态保护
    systrayRunning atomic.Bool   // 运行状态
    quitting       atomic.Bool   // 退出标志
}
```

**功能：**
- 窗口最小化到托盘
- 双击/右键事件处理
- 安全退出机制
- 多级图标回退策略

## 入口点

### 应用入口

**main.go**: Wails 应用主入口
- 嵌入前端资源 (`frontend/dist`)
- 配置应用选项（窗口、图标、生命周期钩子）
- 初始化日志系统
- 创建 Wails 应用实例

**app.go**: 后端逻辑入口
- `startup()`: 初始化数据库、配置、服务
- `shutdown()`: 清理资源
- `onBeforeClose()`: 拦截窗口关闭，转入托盘

### 前端入口

**main.ts**: Vue 应用入口
- 初始化 Pinia 状态管理
- 配置 Wails 事件监听
- 启动画面超时保护

**App.vue**: 主组件入口
- 应用布局结构
- 事件监听和处理
- 生命周期管理

## 核心设计原则

1. **关注点分离**: 后端专注业务逻辑，前端专注 UI 交互
2. **事件驱动**: 使用 Wails Events 实现前后端异步通信
3. **缓存优先**: StatusCache 提供快速响应，后台更新保持数据新鲜
4. **乐观更新**: 用户操作立即反馈，异步验证结果
5. **容错设计**: 多级回退、超时保护、错误恢复
6. **可扩展性**: Provider Registry 模式支持新增 AI 服务

---

*架构分析：2024-02-05*