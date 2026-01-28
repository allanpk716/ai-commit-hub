# 启动画面与项目状态预加载功能设计

**日期**: 2026-01-28
**状态**: 设计阶段
**优先级**: 高

## 概述

在应用启动时显示欢迎画面，同时在后台预加载所有项目的 Pushover Hook 版本和 Git 状态。完成后在项目列表中直观显示状态指示器，无需点击进入项目即可了解项目状态。

## 功能需求

### 启动画面
- 显示应用 Logo 和版本号
- 动态加载进度条和状态消息
- 自动完成并进入主界面

### 项目状态指示
- **Pushover 更新**: 向上箭头图标（⬆️）
- **未提交更改**: 回转图标（🔄）
- **未跟踪文件**: 加号图标 + 数量（➕ N）

## 设计方案

### 1. 启动画面组件 (SplashScreen.vue)

**视觉元素**
- 居中的应用 Logo
- 应用名称 "AI Commit Hub"
- 版本号显示
- 进度条（0-100%）
- 当前加载阶段文字说明

**加载阶段**
| 阶段 | 进度 | 任务 |
|------|------|------|
| 初始化 | 10% | 数据库连接、配置加载 |
| 扩展检查 | 20% | 检查 Pushover 扩展状态 |
| 项目扫描 | 50% | 并发扫描所有项目状态 |
| 完成 | 100% | 准备进入主界面 |

### 2. 后端架构

**新增 API 方法**

```go
// StartupPreload 启动预加载（在 startup 中调用 goroutine）
func (a *App) StartupPreload()

// GetProjectsWithStatus 获取带状态的项目列表
func (a *App) GetProjectsWithStatus() ([]models.GitProject, error)
```

**数据模型扩展**

```go
type GitProject struct {
    // ... 现有字段

    // 运行时状态（不持久化）
    HasUncommittedChanges bool `json:"has_uncommitted_changes" gorm:"-"`
    UntrackedCount       int  `json:"untracked_count" gorm:"-"`
    PushoverNeedsUpdate  bool `json:"pushover_needs_update" gorm:"-"`
}
```

**事件流**
```
app.startup()
  ↓
StartupPreload() goroutine
  ↓
EventsEmit("startup-progress", {stage, percent, message})
  ↓
前端监听并更新 UI
  ↓
EventsEmit("startup-complete")
  ↓
自动切换到主界面
```

### 3. 前端状态管理

**新增 Store: startupStore.ts**
- 管理启动画面状态
- 监听 Wails Events
- 存储预加载的项目数据

**修改 ProjectList.vue**

在项目卡片下方添加状态行：
```vue
<div class="project-status-row">
  <span v-if="project.has_uncommitted_changes"
        class="status-indicator uncommitted"
        title="有未提交更改">🔄</span>
  <span v-if="project.untracked_count > 0"
        class="status-indicator untracked"
        :title="`${project.untracked_count} 个未跟踪文件`">
    ➕ {{ project.untracked_count }}
  </span>
  <span v-if="project.pushover_needs_update"
        class="status-indicator update"
        title="Pushover 插件可更新">⬆️</span>
</div>
```

**样式规范**
| 状态 | 颜色 | 图标 |
|------|------|------|
| 未提交 | #f97316 (橙色) | 🔄 |
| 未跟踪 | #eab308 (黄色) | ➕ |
| Pushover 更新 | #3b82f6 (蓝色) | ⬆️ |

### 4. 错误处理

**部分失败处理**
- 单个项目检查失败不影响其他项目
- 在进度消息中记录失败原因
- 标记该项目为 "检查失败"

**超时策略**
- 单个项目检查：3 秒超时
- 总启动时间：30 秒超时
- 超时后强制进入主界面

**降级方案**
- 预加载失败时仍可正常启动
- 用户点击项目时再进行状态检查
- 保留现有的按需检查方法

## 实施计划

### Phase 1: 核心功能
1. 创建 `SplashScreen.vue` 组件
2. 实现 `StartupPreload()` 后端方法
3. 添加项目状态缓存结构
4. 建立事件流通信

### Phase 2: 状态检查
1. 修改 `ProjectList.vue` 添加状态图标
2. 实现并发状态检查逻辑
3. 添加 Git 状态扫描

### Phase 3: 优化完善
1. 添加超时和错误处理
2. 性能优化（goroutine pool）
3. 动画和交互细节

## 技术要点

- **并发控制**: 使用 `errgroup` 或 goroutine pool 控制并发数
- **内存管理**: 状态数据不持久化，仅运行时使用
- **事件驱动**: 使用 Wails Events 实现前后端通信
- **降级策略**: 确保预加载失败不影响主功能

## 测试场景

- 空项目（首次启动）
- 单项目/多项目（10+）
- 所有状态组合
- 项目路径不存在
- 非 Git 仓库
- 性能测试（50+ 项目）
