# cc-pushover-hook 集成设计文档

**日期**: 2025-01-23
**状态**: 设计阶段
**开发分支**: 使用 git worktree 隔离开发

---

## 概述

为 AI Commit Hub 添加 cc-pushover-hook 集成功能，允许用户为导入的 Git 项目安装和管理 Pushover 通知 Hook。

---

## 功能需求

### 核心功能

1. **自动下载扩展**: 程序启动时检查 `extensions/cc-pushover-hook/` 目录，不存在则提示用户下载
2. **一键安装**: 在 CommitPanel 中为选中项目安装 cc-pushover-hook
3. **通知模式管理**: 通过预设模式控制通知行为
4. **状态可视化**: 实时显示 Hook 安装状态和通知配置
5. **扩展管理**: 在设置界面管理 cc-pushover-hook 的下载和更新

### 预设通知模式

| 模式 | .no-pushover | .no-windows | 说明 |
|------|--------------|-------------|------|
| 全部启用 | ✗ | ✗ | 所有通知都启用 |
| 仅 Pushover | ✗ | ✓ | 仅 Pushover 通知 |
| 仅 Windows | ✓ | ✗ | 仅 Windows 桌面通知 |
| 全部禁用 | ✓ | ✓ | 不发送任何通知 |

---

## 架构设计

### 后端架构 (Go)

```
pkg/pushover/
├── service.go              # PushoverService 核心服务
├── installer.go            # install.py 调用封装
├── status.go               # 状态检测逻辑
└── repository.go           # Git 操作封装
```

#### PushoverService 方法

```go
// 检查 Hook 是否已安装
CheckHookInstalled(projectPath string) (bool, error)

// 获取 Hook 详细状态
GetHookStatus(projectPath string) (*HookStatus, error)

// 安装 Hook 到项目
InstallHook(projectPath string, force bool) error

// 卸载 Hook
UninstallHook(projectPath string) error

// 设置通知模式
SetNotificationMode(projectPath string, mode NotificationMode) error

// 克隆扩展仓库
CloneExtension() error

// 更新扩展到最新版本
UpdateExtension() error

// 获取扩展信息
GetExtensionInfo() (*ExtensionInfo, error)
```

#### 数据模型更新

```go
// GitProject 新增字段
type GitProject struct {
    // ... 现有字段

    HookInstalled      bool              `gorm:"default:false"`
    NotificationMode   string            `gorm:"default:'enabled'"` // enabled/pushover_only/windows_only/disabled
    HookVersion        string            `gorm:"size:50"`
    HookInstalledAt    *time.Time
}

type NotificationMode string

const (
    ModeEnabled        NotificationMode = "enabled"
    ModePushoverOnly   NotificationMode = "pushover_only"
    ModeWindowsOnly    NotificationMode = "windows_only"
    ModeDisabled       NotificationMode = "disabled"
)

type HookStatus struct {
    Installed          bool
    Mode               NotificationMode
    Version            string
    InstalledAt        time.Time
}

type ExtensionInfo struct {
    Downloaded         bool
    Path               string
    Version            string
    LatestVersion      string
    UpdateAvailable    bool
}
```

#### App 层导出方法

```go
// 获取项目 Hook 状态
func (a *App) GetPushoverHookStatus(projectPath string) (*HookStatus, error)

// 安装 Hook
func (a *App) InstallPushoverHook(projectPath string, force bool) error

// 设置通知模式
func (a *App) SetPushoverNotificationMode(projectPath string, mode string) error

// 获取扩展信息
func (a *App) GetPushoverExtensionInfo() (*ExtensionInfo, error)

// 克隆扩展
func (a *App) ClonePushoverExtension() error

// 更新扩展
func (a *App) UpdatePushoverExtension() error
```

---

### 前端架构 (Vue3)

#### 新增 Store: `pushoverStore.ts`

```typescript
interface PushoverState {
  extensionDownloaded: boolean
  extensionVersion: string
  updateAvailable: boolean
  projectHookStatus: Map<string, HookStatus>
}

interface HookStatus {
  installed: boolean
  mode: 'enabled' | 'pushover_only' | 'windows_only' | 'disabled'
  version: string
  installedAt: string
}

export const usePushoverStore = defineStore('pushover', {
  state: (): PushoverState => ({...}),

  actions: {
    async checkExtensionStatus() {...}
    async cloneExtension() {...}
    async updateExtension() {...}
    async getProjectHookStatus(projectPath: string) {...}
    async installHook(projectPath: string, force: boolean) {...}
    async setNotificationMode(projectPath: string, mode: string) {...}
  }
})
```

#### 新增组件

**`PushoverStatusBadge.vue`** - 状态徽章
- 位置：CommitPanel 顶部项目名称旁
- 显示：安装状态图标 + 颜色标识
- 交互：悬停显示简要信息

**`PushoverStatusCard.vue`** - 状态卡片
- 位置：CommitPanel 内，可折叠
- 内容：详细状态 + 预设模式选择器 + 操作按钮

**`PushoverManagementPanel.vue`** - 管理面板
- 位置：设置对话框内
- 内容：扩展信息 + 项目列表 + 批量操作

**`SettingsDialog.vue`** - 设置对话框（新增）
- 替换现有"打开配置文件夹"按钮
- 包含配置管理和 cc-pushover-hook 管理两个区域

---

## cc-pushover-hook 修改

### install.py 改造

#### 新增命令行参数

```bash
python install.py [OPTIONS]

选项:
  -t, --target-dir PATH    目标项目路径（必需）
  --force                  强制重新安装，覆盖现有文件
  --non-interactive        非交互模式，不询问确认
  --skip-diagnostics       跳过安装后的诊断
  --quiet                  静默模式，减少输出
  --version                显示版本信息
```

#### 输出格式

非交互模式下最后一行输出 JSON 结果：

```json
{"status": "success", "hook_path": "/path/to/.claude/hooks/pushover-hook", "version": "1.0.0"}
```

或错误：

```json
{"status": "error", "message": "错误描述"}
```

#### 保持向后兼容

无参数调用时保持原有交互式行为。

---

## 交互流程

### 启动检查流程

```
程序启动
  ↓
检查 extensions/cc-pushover-hook/ 目录
  ↓
不存在 → 显示提醒横幅（可点击前往设置）
存在 → 获取版本信息，存入状态
```

### 安装 Hook 流程

```
用户在 CommitPanel 点击"安装 Hook"
  ↓
检查扩展是否已下载
  ↓
未下载 → 提示前往设置界面下载
已下载 → 调用 InstallPushoverHook(projectPath)
  ↓
后端执行 install.py -t projectPath --non-interactive
  ↓
解析 JSON 输出
  ↓
成功 → 更新数据库，刷新 UI
失败 → 显示错误信息
```

### 设置通知模式流程

```
用户选择预设模式
  ↓
调用 SetPushoverNotificationMode(projectPath, mode)
  ↓
后端操作文件：
  - 全部启用：删除两个标记文件
  - 仅 Pushover：创建 .no-windows
  - 仅 Windows：创建 .no-pushover
  - 全部禁用：创建两个标记文件
  ↓
更新数据库 NotificationMode
  ↓
前端刷新状态
```

---

## UI 设计

### 设置对话框布局

```
┌─────────────────────────────────────┐
│ 设置                          [×]   │
├─────────────────────────────────────┤
│                                     │
│ ┌─ 配置管理 ────────────────────┐  │
│ │ • 打开配置文件夹               │  │
│ └─────────────────────────────┘  │
│                                     │
│ ┌─ cc-pushover-hook 管理 ──────┐  │
│ │ 状态: ✅ 已下载 (v1.0.0)       │  │
│ │ [检查更新] [重新下载]          │  │
│ │                                │  │
│ │ 已安装的项目:                  │  │
│ │ • ProjectA (全部启用)          │  │
│ │ • ProjectB (仅 Pushover)       │  │
│ └─────────────────────────────┘  │
│                                     │
│           [关闭]                    │
└─────────────────────────────────────┘
```

### CommitPanel 状态显示

```
┌─────────────────────────────────────┐
│ ProjectName              🔔 ●────── │  ← 状态徽章
├─────────────────────────────────────┤
│                                     │
│ ┌─ Hook 状态 ─────── [▼] ─────┐  │
│ │ ✅ Pushover Hook 已启用         │  │
│ │ 模式: 全部启用                  │  │
│ │ [更改模式] [卸载]               │  │
│ └─────────────────────────────┘  │
│                                     │
│ ... 原有 commit 生成内容 ...        │
└─────────────────────────────────────┘
```

### 提醒横幅

```
┌─────────────────────────────────────┐
│ 🔔 cc-pushover-hook 扩展未下载      │
│ [前往设置] [忽略]                   │
└─────────────────────────────────────┘
```

---

## 错误处理

### 网络错误

- clone/update 失败显示友好提示
- 提供"重试"按钮

### 安装失败

- 解析 install.py 错误输出
- 显示针对性错误信息

### 状态异常

- 检测到损坏文件时提示"重新安装"

### 边界情况

- 项目路径变更：重新验证路径
- 旧版本文件：自动清理
- 未下载扩展：提示用户先下载
- 并发安装：loading 状态防重复

---

## 文件结构

### 后端新增文件

```
pkg/pushover/
├── service.go
├── installer.go
├── status.go
└── repository.go

extensions/                         # 运行时创建
└── cc-pushover-hook/              # git clone 到此
```

### 前端新增文件

```
frontend/src/
├── stores/
│   └── pushoverStore.ts
├── components/
│   ├── PushoverStatusBadge.vue
│   ├── PushoverStatusCard.vue
│   ├── PushoverManagementPanel.vue
│   └── SettingsDialog.vue
└── types/
    └── pushover.ts                 # 类型定义
```

---

## 开发计划

1. **阶段一**: 后端基础功能
   - 创建 `pkg/pushover/` 模块
   - 实现 PushoverService 核心方法
   - 数据库迁移添加新字段

2. **阶段二**: cc-pushover-hook 改造
   - 修改 install.py 支持命令行参数
   - 测试非交互式安装

3. **阶段三**: 前端基础
   - 创建 pushoverStore
   - 实现 PushoverStatusBadge 组件

4. **阶段四**: 管理界面
   - 实现 SettingsDialog
   - 实现 PushoverManagementPanel

5. **阶段五**: 集成测试
   - 端到端测试完整流程
   - 错误处理测试

---

## 依赖

- Go 1.21+
- Python 3.6+（调用 install.py）
- Git（clone/pull 操作）

---

## 参考资料

- cc-pushover-hook README: `C:\WorkSpace\agent\cc-pushover-hook\README.md`
- cc-pushover-hook install.py: `C:\WorkSpace\agent\cc-pushover-hook\install.py`
- Wails 文档: https://wails.io/docs/next/introduction
