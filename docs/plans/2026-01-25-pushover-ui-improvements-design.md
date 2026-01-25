# cc-pushover-hook 扩展 UI 改进设计

**日期**: 2026-01-25
**状态**: 设计阶段
**优先级**: 中

## 问题概述

cc-pushover-hook 扩展的当前 UI 存在以下用户体验问题：

1. **扩展路径点击错误** - 点击"扩展路径"打开的是 ai-commit-hub 配置文件夹，而非扩展实际目录
2. **版本状态不清晰** - 已是最新版本时缺少明确提示
3. **冗余按钮** - 扩展信息弹窗中有不需要的"打开配置文件夹"按钮
4. **状态指示器不明确** - 主界面的扩展状态只有小图标，看不出是哪个插件

## 设计目标

- 提高扩展管理的直观性和易用性
- 统一界面语言和交互模式
- 减少用户困惑，明确区分不同功能入口

## 解决方案

### 1. 修复扩展路径点击行为

**问题**: 点击扩展路径打开错误的文件夹

**解决方案**: 添加专用 API 方法打开扩展目录

#### 后端修改 (`app.go`)

```go
// OpenExtensionFolder opens the cc-pushover-hook extension folder
func (a *App) OpenExtensionFolder() error {
    extensionPath, err := a.pushoverRepo.GetExtensionPath()
    if err != nil {
        return fmt.Errorf("failed to get extension path: %w", err)
    }

    // 检查目录是否存在
    if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
        return fmt.Errorf("extension directory not found: %s", extensionPath)
    }

    var cmd *exec.Cmd
    switch stdruntime.GOOS {
    case "windows":
        cmd = exec.Command("explorer", extensionPath)
    case "darwin":
        cmd = exec.Command("open", extensionPath)
    default:
        cmd = exec.Command("xdg-open", extensionPath)
    }

    return cmd.Start()
}
```

#### 前端修改 (`ExtensionInfoDialog.vue`)

```vue
<!-- 修改扩展路径点击事件 -->
<span
  v-if="pushoverStore.extensionInfo.path"
  class="value value-path clickable"
  @click="handleOpenExtensionFolder"
>
  {{ pushoverStore.extensionInfo.path }}
</span>
```

```typescript
// 添加新方法
async function handleOpenExtensionFolder() {
  try {
    const { OpenExtensionFolder } = await import('../../wailsjs/go/main/App')
    await OpenExtensionFolder()
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : '未知错误'
    pushoverStore.error = `打开扩展文件夹失败: ${message}`
  }
}
```

### 2. 添加"已是最新"版本提示

**问题**: 当前版本时缺少明确的状态反馈

**解决方案**: 添加绿色提示条，与更新提示对称

#### 前端修改 (`ExtensionInfoDialog.vue`)

```vue
<!-- 在版本卡片中添加 -->
<div class="version-card">
  <h3>版本信息</h3>
  <div class="info-row">
    <span class="label">当前版本:</span>
    <span class="value">{{ pushoverStore.extensionInfo.current_version }}</span>
  </div>

  <!-- 有更新提示（现有） -->
  <div v-if="pushoverStore.isUpdateAvailable" class="info-row">
    <span class="label">最新版本:</span>
    <span class="value text-accent">{{ pushoverStore.extensionInfo.latest_version }}</span>
  </div>
  <div v-if="pushoverStore.isUpdateAvailable" class="update-hint">
    有新版本可用，建议更新扩展
  </div>

  <!-- 新增：已是最新提示 -->
  <div v-if="!pushoverStore.isUpdateAvailable && pushoverStore.isExtensionDownloaded" class="latest-hint">
    ✅ 已是最新版本
  </div>
</div>
```

```css
/* 新增样式 */
.latest-hint {
  margin-top: var(--space-sm);
  padding: var(--space-sm);
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: var(--radius-sm);
  color: #22c55e;
  font-size: 13px;
  text-align: center;
}
```

### 3. 移除"打开配置文件夹"按钮

**问题**: 扩展信息弹窗中不需要打开 ai-commit-hub 配置文件夹

**解决方案**: 删除操作按钮区域中的该按钮

#### 前端修改 (`ExtensionInfoDialog.vue`)

删除以下按钮（第 97-102 行）：

```vue
<button
  class="btn btn-secondary"
  @click="handleOpenConfigFolder"
>
  打开配置文件夹
</button>
```

**注意**: 保留 `handleOpenConfigFolder` 方法，因为设置对话框 (`SettingsDialog.vue`) 中仍需要使用。

### 4. 改进主界面状态指示器

**问题**: 当前只有小图标，无法识别是哪个插件

**解决方案**: 改为带文字的紧凑按钮样式

#### 完整重构 (`ExtensionStatusButton.vue`)

```vue
<template>
  <button
    @click="openDialog"
    class="extension-status-btn"
    :class="statusClass"
    :title="statusTitle"
  >
    <span class="btn-icon">🔔</span>
    <span class="btn-text">Pushover 扩展</span>
    <span class="status-badge">{{ statusBadge }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'

const emit = defineEmits<{
  open: []
}>()

const pushoverStore = usePushoverStore()

const statusClass = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return 'status-error'
  if (pushoverStore.isUpdateAvailable) return 'status-update'
  return 'status-ok'
})

const statusBadge = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return '!'
  if (pushoverStore.isUpdateAvailable) return '↑'
  return '✓'
})

const statusTitle = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return 'cc-pushover-hook 扩展未下载'
  if (pushoverStore.isUpdateAvailable)
    return `cc-pushover-hook 有更新可用 (v${pushoverStore.extensionInfo.latest_version})`
  return `cc-pushover-hook 已是最新版本 (v${pushoverStore.extensionInfo.current_version})`
})

function openDialog() {
  emit('open')
}

onMounted(async () => {
  await pushoverStore.checkExtensionStatus()
})
</script>

<style scoped>
.extension-status-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-md);
  border: none;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  color: white;
  min-width: 120px;
}

.extension-status-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.btn-icon {
  font-size: 14px;
}

.btn-text {
  flex: 1;
}

.status-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  font-size: 11px;
  font-weight: bold;
}

/* 状态变体 */
.status-ok {
  background: linear-gradient(135deg, #10b981, #059669);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.status-update {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.status-error {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}
</style>
```

#### 工具栏布局调整 (`App.vue`)

确保按钮间距一致：

```vue
<div class="toolbar-actions">
  <button @click="openAddProject" class="btn btn-primary">
    <span class="icon">＋</span>
    <span>添加项目</span>
  </button>

  <!-- Pushover 扩展按钮 -->
  <ExtensionStatusButton @open="extensionDialogOpen = true" />

  <button @click="openSettings" class="btn btn-secondary">
    <span class="icon">⚙</span>
    <span>设置</span>
  </button>
</div>
```

## 修改文件清单

| 文件 | 修改类型 | 优先级 |
|------|----------|--------|
| `app.go` | 新增方法 | 高 |
| `pkg/pushover/repository.go` | 可能需要添加 `GetExtensionPath()` | 高 |
| `frontend/src/components/ExtensionInfoDialog.vue` | 修改 | 高 |
| `frontend/src/components/ExtensionStatusButton.vue` | 重构 | 高 |
| `frontend/wailsjs/go/main/App.js` | 自动生成（Wails 绑定） | - |

## 测试计划

1. **扩展路径点击测试**
   - 点击路径应打开扩展实际目录
   - 扩展未下载时应显示错误提示

2. **版本状态测试**
   - 最新版本时显示绿色"已是最新"提示
   - 有更新时显示橙色更新提示
   - 未下载时不显示版本信息

3. **按钮移除测试**
   - 确认弹窗中无"打开配置文件夹"按钮
   - 设置对话框中该按钮仍正常工作

4. **状态指示器测试**
   - 三种状态显示正确颜色和徽章
   - 按钮点击正确打开扩展信息弹窗
   - 工具提示显示完整信息

## 实施注意事项

1. Wails 绑定生成后需重启开发服务器
2. 确保 `GetExtensionPath()` 方法已存在于 PushoverRepository
3. 测试时需覆盖三种状态：未下载、已下载最新、有更新
4. 保持与其他按钮的视觉一致性

## 后续优化建议

1. 考虑添加扩展自动更新功能
2. 添加扩展健康检查（如验证关键文件存在）
3. 考虑支持多个扩展的可扩展架构
