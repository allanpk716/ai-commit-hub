# Pushover Hook UI 重设计

**日期**: 2026-01-25
**作者**: Claude
**状态**: 设计完成，待实现

## 概述

将当前折叠式的 Pushover Hook 状态卡片重新设计为单行紧凑组件，并与通知状态图标合并。新的设计更简洁、交互更直观。

## 设计目标

1. **简化 UI**：将折叠卡片改为单行显示，节省垂直空间
2. **统一交互**：将通知状态图标与 Hook 状态合并到同一组件
3. **直观操作**：点击通知图标直接切换状态（创建/删除控制文件）
4. **清晰反馈**：用颜色（绿/黄/红）和图标表示 Hook 状态

## 组件设计

### 新组件：PushoverStatusRow.vue

单行状态组件，替换现有的 `PushoverStatusCard.vue` 和 `PushoverStatusBadge.vue`。

#### 布局结构

```
┌─────────────────────────────────────────────────────────────────────┐
│ 🟢 Pushover Hook v1.2.3    [📱] [💻]                    [更新 Hook] │
└─────────────────────────────────────────────────────────────────────┘
```

从左到右依次为：
1. **状态图标**：🟢(最新) / 🟡(有更新) / 🔴(未安装)
2. **标题**："Pushover Hook"
3. **版本号**：当前版本（如果已安装）
4. **通知图标**：可点击切换
   - 📱 Pushover（蓝色启用 / 灰色禁用）
   - 💻 Windows（紫色启用 / 灰色禁用）
5. **操作按钮**：
   - 未安装 → "安装 Hook"
   - 有更新 → "更新 Hook"
   - 已是最新 → "已是最新" 文本

#### 状态示例

**未安装**
```
🔴 Pushover Hook (未安装)                                    [安装 Hook]
```

**有更新，两种通知都启用**
```
🟡 Pushover Hook v1.0.0        [📱] [💻]                   [更新 Hook]
```

**最新版本，仅 Pushover 启用**
```
🟢 Pushover Hook v1.2.3        [📱] [💻]                        已是最新
```

## 交互逻辑

### 通知切换

点击通知图标切换状态，通过创建/删除控制文件实现：

- **点击 📱 图标**：
  - 启用 → 删除 `.no-pushover` 文件
  - 禁用 → 创建 `.no-pushover` 文件

- **点击 💻 图标**：
  - 启用 → 删除 `.no-windows` 文件
  - 禁用 → 创建 `.no-windows` 文件

### 视觉反馈

- 禁用状态：图标变灰 + 半透明
- 启用状态：图标高亮 + 对应颜色
- 点击：短暂动画效果
- hover：显示 tooltip 提示当前状态和点击操作

## 技术实现

### 前端组件

#### 组件结构

```vue
<template>
  <div class="pushover-status-row" :class="rowClass">
    <div class="status-left">
      <span class="status-icon">{{ statusIcon }}</span>
      <span class="status-title">Pushover Hook</span>
      <span v-if="status?.version" class="status-version">v{{ status.version }}</span>
    </div>

    <div v-if="status?.installed" class="notification-toggles">
      <button :class="{ active: isPushoverEnabled }" @click="togglePushover">
        📱
      </button>
      <button :class="{ active: isWindowsEnabled }" @click="toggleWindows">
        💻
      </button>
    </div>

    <div class="status-right">
      <span v-if="isLatest">已是最新</span>
      <button v-else-if="!status?.installed" @click="handleInstall">
        安装 Hook
      </button>
      <button v-else-if="needsUpdate" @click="handleUpdate">
        更新 Hook
      </button>
    </div>
  </div>
</template>
```

#### 计算属性

```typescript
const statusIcon = computed(() => {
  if (!status.value?.installed) return '🔴'
  if (needsUpdate.value) return '🟡'
  return '🟢'
})

const isPushoverEnabled = computed(() => {
  return status.value?.mode === 'enabled' || status.value?.mode === 'pushover_only'
})

const isWindowsEnabled = computed(() => {
  return status.value?.mode === 'enabled' || status.value?.mode === 'windows_only'
})
```

### 后端 API

#### 新增方法

```go
// ToggleNotification 切换指定项目的通知类型
func (a *App) ToggleNotification(projectPath string, notificationType string) error {
    var fileName string
    switch notificationType {
    case "pushover":
        fileName = ".no-pushover"
    case "windows":
        fileName = ".no-windows"
    default:
        return fmt.Errorf("不支持的通知类型: %s", notificationType)
    }

    filePath := filepath.Join(projectPath, fileName)

    // 切换文件：存在则删除，不存在则创建
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        os.Create(filePath) // 禁用通知
    } else {
        os.Remove(filePath) // 启用通知
    }

    return nil
}
```

#### 配置检查（应用启动时调用）

```go
// CheckPushoverConfig 检查 Pushover 环境变量是否已配置
func (a *App) CheckPushoverConfig() map[string]interface{} {
    token := os.Getenv("PUSHOVER_TOKEN")
    user := os.Getenv("PUSHOVER_USER")

    return map[string]interface{}{
        "valid": token != "" && user != "",
        "token_set": token != "",
        "user_set": user != "",
    }
}
```

### Pushover Store 扩展

```typescript
async toggleNotification(projectPath: string, type: 'pushover' | 'windows') {
  try {
    await ToggleNotification(projectPath, type)
    await this.getProjectHookStatus(projectPath)
  } catch (error) {
    console.error('切换通知失败:', error)
  }
}

async checkPushoverConfig(): Promise<boolean> {
  try {
    const result = await CheckPushoverConfig()
    this.configValid = result.valid
    return result.valid
  } catch (error) {
    this.configValid = false
    return false
  }
}
```

### CommitPanel 集成

```vue
<template>
  <!-- 替换原有的 PushoverStatusCard -->
  <PushoverStatusRow
    v-if="currentProject"
    :project-path="currentProject.path"
    :status="pushoverStatus"
    @install="handleInstallHook"
    @update="handleUpdateHook"
  />
</template>
```

## 数据流

```
应用启动
  ↓
检查 PUSHOVER_TOKEN/PUSHOVER_USER 环境变量
  ↓
用户选择项目
  ↓
获取项目 Hook 状态（检查 .no-pushover/.no-windows 文件）
  ↓
显示单行状态组件
  ↓
用户点击通知图标
  ↓
调用 ToggleNotification API（创建/删除控制文件）
  ↓
刷新状态显示
```

## 文件变更

### 新增文件

- `frontend/src/components/PushoverStatusRow.vue` - 单行状态组件

### 修改文件

- `frontend/src/components/CommitPanel.vue` - 集成新组件
- `frontend/src/stores/pushoverStore.ts` - 新增切换方法
- `app.go` - 新增 API 方法
- `wailsjs/go/main/App.js` - 自动生成的绑定

### 删除文件

- `frontend/src/components/PushoverStatusBadge.vue` - 功能已合并
- `frontend/src/components/PushoverStatusCard.vue` - 替换为新组件

## 实现步骤

1. 创建 `PushoverStatusRow.vue` 组件
2. 在 `app.go` 中添加 `ToggleNotification` 方法
3. 扩展 `pushoverStore.ts` 添加切换和配置检查方法
4. 在 `CommitPanel.vue` 中集成新组件
5. 删除旧的 `PushoverStatusBadge.vue` 和 `PushoverStatusCard.vue`
6. 测试所有交互场景
