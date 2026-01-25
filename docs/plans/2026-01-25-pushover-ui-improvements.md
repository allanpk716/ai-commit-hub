# cc-pushover-hook 扩展 UI 改进实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 修复 cc-pushover-hook 扩展的用户体验问题，包括路径点击行为、版本状态显示、冗余按钮和状态指示器。

**架构:** 前后端分离架构 - Go 后端提供 API，Vue3 前端消费。修改涉及新增后端 API 方法、修改 Vue 组件样式和逻辑。

**技术栈:** Go 1.21+, Wails v2, Vue 3, TypeScript, Pinia

---

## Task 1: 添加后端 API - 打开扩展文件夹

**文件:**
- 修改: `app.go`
- 测试: 无需单独测试（通过手动测试验证）

### Step 1: 添加 OpenExtensionFolder 方法到 app.go

在 `OpenConfigFolder()` 方法后添加新方法：

```go
// OpenExtensionFolder opens the cc-pushover-hook extension folder in system file manager
func (a *App) OpenExtensionFolder() error {
	// 获取扩展路径
	extensionPath, err := a.pushoverRepo.GetExtensionPath()
	if err != nil {
		return fmt.Errorf("failed to get extension path: %w", err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		return fmt.Errorf("extension directory not found: %s", extensionPath)
	}

	// 根据操作系统选择命令
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

### Step 2: 验证 PushoverRepository 有 GetExtensionPath 方法

检查 `pkg/pushover/repository.go` 是否有 `GetExtensionPath()` 方法。

运行: `grep -n "GetExtensionPath" pkg/pushover/repository.go`
预期: 找到方法定义

如果不存在，添加该方法：

```go
// GetExtensionPath returns the path where cc-pushover-hook extension is stored
func (r *PushoverRepository) GetExtensionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".ai-commit-hub", "extensions", "cc-pushover-hook"), nil
}
```

### Step 3: 重新生成 Wails 绑定

运行: `wails generate module`
预期: 在 `frontend/wailsjs/go/main/App.js` 中生成新的绑定

### Step 4: 提交

```bash
git add app.go pkg/pushover/repository.go frontend/wailsjs/go/main/App.js
git commit -m "feat: 添加 OpenExtensionFolder API 用于打开扩展文件夹"
```

---

## Task 2: 修改扩展信息弹窗 - 修复路径点击

**文件:**
- 修改: `frontend/src/components/ExtensionInfoDialog.vue`

### Step 1: 添加 handleOpenExtensionFolder 方法

在 `<script setup>` 中，`handleOpenConfigFolder` 方法后添加：

```typescript
// 打开扩展文件夹
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

### Step 2: 修改扩展路径点击事件

将扩展路径的 `@click` 从 `handleOpenConfigFolder` 改为 `handleOpenExtensionFolder`：

```vue
<span
  v-if="pushoverStore.extensionInfo.path"
  class="value value-path clickable"
  @click="handleOpenExtensionFolder"
>
  {{ pushoverStore.extensionInfo.path }}
</span>
```

### Step 3: 提交

```bash
git add frontend/src/components/ExtensionInfoDialog.vue
git commit -m "fix: 扩展路径点击打开扩展文件夹而非配置文件夹"
```

---

## Task 3: 修改扩展信息弹窗 - 添加"已是最新"提示

**文件:**
- 修改: `frontend/src/components/ExtensionInfoDialog.vue`

### Step 1: 添加"已是最新"提示 HTML

在版本卡片中，`update-hint` 后添加：

```vue
<div class="version-card">
  <h3>版本信息</h3>
  <!-- ... 现有内容 ... -->
  <div v-if="pushoverStore.isUpdateAvailable" class="update-hint">
    有新版本可用，建议更新扩展
  </div>
  <!-- 新增：已是最新提示 -->
  <div v-if="!pushoverStore.isUpdateAvailable && pushoverStore.isExtensionDownloaded" class="latest-hint">
    ✅ 已是最新版本
  </div>
</div>
```

### Step 2: 添加 latest-hint 样式

在 `<style scoped>` 中，`.update-hint` 样式后添加：

```css
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

### Step 3: 提交

```bash
git add frontend/src/components/ExtensionInfoDialog.vue
git commit -m "feat: 扩展已是最新时显示明确提示"
```

---

## Task 4: 修改扩展信息弹窗 - 移除"打开配置文件夹"按钮

**文件:**
- 修改: `frontend/src/components/ExtensionInfoDialog.vue`

### Step 1: 删除"打开配置文件夹"按钮

删除操作按钮区域中的以下代码（约第 97-102 行）：

```vue
<button
  class="btn btn-secondary"
  @click="handleOpenConfigFolder"
>
  打开配置文件夹
</button>
```

### Step 2: 提交

```bash
git add frontend/src/components/ExtensionInfoDialog.vue
git commit -m "refactor: 移除扩展信息弹窗中的打开配置文件夹按钮"
```

---

## Task 5: 重构扩展状态按钮组件

**文件:**
- 完全重写: `frontend/src/components/ExtensionStatusButton.vue`

### Step 1: 完全替换 ExtensionStatusButton.vue 内容

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

### Step 2: 提交

```bash
git add frontend/src/components/ExtensionStatusButton.vue
git commit -m "refactor: 扩展状态按钮改为带文字的紧凑样式"
```

---

## Task 6: 手动测试验证

### Step 1: 启动开发服务器

运行: `wails dev`
预期: 应用正常启动，前端热更新生效

### Step 2: 测试扩展路径点击

1. 打开扩展信息弹窗
2. 点击扩展路径
3. 预期: 打开 cc-pushover-hook 扩展的实际目录

### Step 3: 测试版本状态显示

1. 检查已是最新版本时显示绿色"✅ 已是最新版本"
2. 检查有更新时显示橙色提示
3. 预期: 两种状态提示都正确显示

### Step 4: 测试按钮移除

1. 打开扩展信息弹窗
2. 确认没有"打开配置文件夹"按钮
3. 预期: 按钮已移除

### Step 5: 测试状态指示器

1. 查看主界面工具栏
2. 确认显示"🔔 Pushover 扩展"按钮
3. 鼠标悬停查看完整提示
4. 点击打开弹窗
5. 预期: 按钮样式正确，功能正常

### Step 6: 测试三种状态

1. 删除扩展目录测试"未下载"状态（红色）
2. 下载扩展测试"最新"状态（绿色）
3. 模拟有更新测试"更新"状态（橙色）
4. 预期: 三种状态颜色和徽章正确

---

## Task 7: 完成和清理

### Step 1: 确认所有修改已提交

运行: `git status`
预期: 除了 Wails 生成的绑定文件外无其他未提交更改

### Step 2: 最终提交 Wails 绑定（如果有）

```bash
git add frontend/wailsjs/
git commit -m "chore: 更新 Wails 绑定"
```

### Step 3: 更新任务状态

Task #1 状态更新为 completed

---

## 检查清单

- [ ] 后端 `OpenExtensionFolder()` API 已添加并生成绑定
- [ ] 扩展路径点击打开扩展文件夹
- [ ] "已是最新"绿色提示已添加
- [ ] "打开配置文件夹"按钮已移除
- [ ] 状态指示器改为带文字的紧凑样式
- [ ] 三种状态（未下载/最新/有更新）显示正确
- [ ] 所有手动测试通过

---

## 相关文档

- 设计文档: `docs/plans/2026-01-25-pushover-ui-improvements-design.md`
- cc-pushover-hook 集成设计: `docs/plans/2025-01-23-pushover-hook-integration-design.md`
- 扩展状态功能: `docs/plans/2026-01-25-pushover-extension-status.md`
