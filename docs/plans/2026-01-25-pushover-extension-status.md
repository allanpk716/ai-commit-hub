# Pushover Hook 扩展状态可视化实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 在主界面工具栏添加 cc-pushover-hook 扩展状态指示器和详细信息对话框，让用户能够查看和管理扩展版本。

**架构:**
- 前端：新增 ExtensionStatusButton.vue 状态指示器组件 + ExtensionInfoDialog.vue 详情对话框
- 后端：新增 CheckPushoverExtensionUpdates() 和 ReclonePushoverExtension() API 方法
- 状态管理：使用现有的 pushoverStore，扩展 checkExtensionStatus 功能

**技术栈:**
- Go 1.21+ + Wails v2（后端）
- Vue 3 + TypeScript + Pinia（前端）
- 玻璃态设计风格（现有 UI）

---

## Task 1: 后端 - 添加检查扩展更新 API

**Files:**
- Modify: `app.go:488-497` (在 GetPushoverExtensionInfo 后添加)

**Step 1: 添加公开 API 方法**

在 `app.go` 中 `GetPushoverExtensionInfo()` 方法后添加新方法：

```go
// CheckPushoverExtensionUpdates 检查 cc-pushover-hook 扩展更新
func (a *App) CheckPushoverExtensionUpdates() (needsUpdate bool, currentVersion string, latestVersion string, err error) {
	if a.initError != nil {
		return false, "", "", a.initError
	}
	if a.pushoverService == nil {
		return false, "", "", fmt.Errorf("pushover service 未初始化")
	}

	needsUpdate, currentVersion, latestVersion, err = a.pushoverService.CheckForUpdates()
	if err != nil {
		return false, "", "", fmt.Errorf("检查扩展更新失败: %w", err)
	}

	return needsUpdate, currentVersion, latestVersion, nil
}
```

**Step 2: 运行 wails dev 生成绑定**

```bash
wails dev
```

Expected: Wails 重新生成绑定，前端自动刷新

**Step 3: Commit**

```bash
git add app.go
git commit -m "feat(pushover): 添加检查扩展更新 API 方法"
```

---

## Task 2: 后端 - 添加重新下载扩展 API

**Files:**
- Modify: `pkg/pushover/repository.go:153-154` (添加 Reclone 方法)
- Modify: `pkg/pushover/service.go:99-105` (在 GetExtensionVersion 后添加)
- Modify: `app.go:608-616` (在 UpdatePushoverExtension 后添加)

**Step 1: 在 RepositoryManager 中添加 Reclone 方法**

在 `repository.go` 的 `GetExtensionInfo()` 方法后添加：

```go
// Reclone 删除并重新克隆扩展
func (rm *RepositoryManager) Reclone() error {
	extensionPath := rm.GetExtensionPath()

	// 删除现有扩展目录
	if rm.IsCloned() {
		if err := os.RemoveAll(extensionPath); err != nil {
			return fmt.Errorf("删除扩展目录失败: %w", err)
		}
	}

	// 重新克隆
	return rm.Clone()
}
```

**Step 2: 在 Service 中添加 RecloneExtension 方法**

在 `service.go` 的 `GetExtensionVersion()` 方法后添加：

```go
// RecloneExtension 删除并重新下载扩展
func (s *Service) RecloneExtension() error {
	return s.repoManager.Reclone()
}
```

**Step 3: 在 App 中添加公开 API**

在 `app.go` 的 `UpdatePushoverExtension()` 方法后添加：

```go
// ReclonePushoverExtension 重新下载 cc-pushover-hook 扩展
func (a *App) ReclonePushoverExtension() error {
	if a.initError != nil {
		return a.initError
	}
	if a.pushoverService == nil {
		return fmt.Errorf("pushover service 未初始化")
	}
	return a.pushoverService.RecloneExtension()
}
```

**Step 4: 运行 wails dev 生成绑定**

```bash
wails dev
```

Expected: Wails 重新生成绑定

**Step 5: Commit**

```bash
git add pkg/pushover/repository.go pkg/pushover/service.go app.go
git commit -m "feat(pushover): 添加重新下载扩展 API"
```

---

## Task 3: 前端 - 更新类型定义

**Files:**
- Modify: `frontend/src/types/index.ts:58-64`

**Step 1: 添加 ExtensionInfo 类型**

在 `types/index.ts` 末尾添加：

```typescript
// Pushover Hook 扩展信息
export interface ExtensionInfo {
  downloaded: boolean      // 是否已下载
  path: string            // 扩展路径
  version: string         // 当前版本
  current_version: string // 当前版本（同 version）
  latest_version: string  // 最新版本
  update_available: boolean // 是否有可用更新
}

// Pushover Hook 状态
export interface HookStatus {
  installed: boolean
  mode: string          // 'silent' | 'normal' | 'verbose'
  version: string       // Hook 版本
  installed_at: string  // 安装时间（ISO 8601）
}

// Hook 安装结果
export interface InstallResult {
  success: boolean
  message: string
}

// 通知模式
export type NotificationMode = 'silent' | 'normal' | 'verbose'
```

**Step 4: Commit**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(types): 添加 Pushover 扩展类型定义"
```

---

## Task 4: 前端 - 更新 pushoverStore

**Files:**
- Modify: `frontend/src/stores/pushoverStore.ts:188-228`

**Step 1: 添加 checkForExtensionUpdates 方法**

在 `pushoverStore.ts` 的 `checkForUpdates()` 方法后添加：

```typescript
  /**
   * 检查扩展自身更新（而非项目 Hook）
   */
  async function checkForExtensionUpdates() {
    loading.value = true
    error.value = null

    try {
      const { CheckPushoverExtensionUpdates } = await import('../../wailsjs/go/main/App')
      const result = await CheckPushoverExtensionUpdates()
      return {
        updateAvailable: result[0] as boolean,
        currentVersion: result[1] as string,
        latestVersion: result[2] as string
      }
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `检查扩展更新失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }
```

**Step 2: 添加 recloneExtension 方法**

在上一步方法后继续添加：

```typescript
  /**
   * 重新下载扩展（删除并克隆）
   */
  async function recloneExtension() {
    loading.value = true
    error.value = null

    try {
      const { ReclonePushoverExtension } = await import('../../wailsjs/go/main/App')
      await ReclonePushoverExtension()
      await checkExtensionStatus()
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : '未知错误'
      error.value = `重新下载扩展失败: ${message}`
      throw e
    } finally {
      loading.value = false
    }
  }
```

**Step 3: 导出新方法**

在 `return` 语句中添加：

```typescript
  return {
    // ... 现有导出
    checkForExtensionUpdates,
    recloneExtension,
  }
```

**Step 4: Commit**

```bash
git add frontend/src/stores/pushoverStore.ts
git commit -m "feat(store): 添加扩展更新检查和重新下载方法"
```

---

## Task 5: 前端 - 创建扩展状态按钮组件

**Files:**
- Create: `frontend/src/components/ExtensionStatusButton.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <button
    @click="openDialog"
    class="extension-status-btn"
    :class="statusClass"
    :title="statusTitle"
  >
    <span class="status-indicator"></span>
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

const statusTitle = computed(() => {
  if (!pushoverStore.isExtensionDownloaded) return '扩展未下载'
  if (pushoverStore.isUpdateAvailable) return `有更新可用 (${pushoverStore.extensionInfo.latest_version})`
  return `已更新到 ${pushoverStore.extensionInfo.current_version || '最新版本'}`
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
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid var(--border-default);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.extension-status-btn:hover {
  transform: scale(1.1);
  border-color: var(--border-hover);
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  position: relative;
}

.status-ok .status-indicator {
  background: var(--accent-success, #10b981);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.status-ok .status-indicator::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 8px;
  color: white;
}

.status-update .status-indicator {
  background: var(--accent-warning, #f59e0b);
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
}

.status-update .status-indicator::after {
  content: '↑';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 8px;
  color: white;
}

.status-error .status-indicator {
  background: var(--accent-error, #ef4444);
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.status-error .status-indicator::after {
  content: '!';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 10px;
  color: white;
  font-weight: bold;
}
</style>
```

**Step 2: Commit**

```bash
git add frontend/src/components/ExtensionStatusButton.vue
git commit -m "feat(component): 创建扩展状态按钮组件"
```

---

## Task 6: 前端 - 创建扩展信息对话框组件

**Files:**
- Create: `frontend/src/components/ExtensionInfoDialog.vue`

**Step 1: 创建对话框组件**

```vue
<template>
  <div v-if="open" class="dialog-overlay" @click.self="close">
    <div class="dialog">
      <!-- Header -->
      <div class="dialog-header">
        <div class="header-title">
          <span class="icon">🔌</span>
          <h2>Pushover Hook 扩展</h2>
        </div>
        <button @click="close" class="close-btn">×</button>
      </div>

      <!-- Content -->
      <div class="dialog-content">
        <!-- Status Card -->
        <div class="status-card">
          <div class="status-row">
            <span class="label">状态：</span>
            <span class="value" :class="statusValueClass">
              {{ statusText }}
            </span>
          </div>
          <div class="status-row">
            <span class="label">路径：</span>
            <span class="value path" @click="openFolder" title="点击打开文件夹">
              {{ extensionInfo.path || '未下载' }}
            </span>
          </div>
        </div>

        <!-- Version Card -->
        <div class="version-card" v-if="extensionInfo.downloaded">
          <div class="version-row">
            <span class="label">当前版本：</span>
            <span class="value">{{ extensionInfo.current_version || '未知' }}</span>
          </div>
          <div class="version-row">
            <span class="label">最新版本：</span>
            <span class="value">{{ extensionInfo.latest_version || '检查中...' }}</span>
          </div>
          <div class="version-diff" v-if="updateAvailable">
            <span class="diff-badge">有新版本可用</span>
          </div>
          <div class="version-diff" v-else-if="extensionInfo.current_version === extensionInfo.latest_version">
            <span class="diff-badge success">已是最新版本</span>
          </div>
        </div>

        <!-- Actions -->
        <div class="actions">
          <button
            @click="checkUpdates"
            :disabled="loading"
            class="btn btn-secondary"
          >
            <span v-if="!loading">🔄 检查更新</span>
            <span v-else>检查中...</span>
          </button>
          <button
            v-if="updateAvailable"
            @click="updateExtension"
            :disabled="loading"
            class="btn btn-primary"
          >
            <span v-if="!loading">⬇️ 更新扩展</span>
            <span v-else>更新中...</span>
          </button>
          <button
            v-if="!extensionInfo.downloaded"
            @click="recloneExtension"
            :disabled="loading"
            class="btn btn-primary"
          >
            <span v-if="!loading">⬇️ 下载扩展</span>
            <span v-else>下载中...</span>
          </button>
          <button
            @click="openGitHub"
            class="btn btn-link"
          >
            <span>在 GitHub 查看</span>
            <span class="external-icon">↗</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import { OpenConfigFolder } from '../../wailsjs/go/main/App'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const pushoverStore = usePushoverStore()
const loading = ref(false)
const error = ref<string | null>(null)

const extensionInfo = computed(() => pushoverStore.extensionInfo)
const updateAvailable = computed(() => pushoverStore.isUpdateAvailable)

const statusText = computed(() => {
  if (!extensionInfo.value.downloaded) return '未下载'
  return '已下载'
})

const statusValueClass = computed(() => {
  if (!extensionInfo.value.downloaded) return 'error'
  return 'success'
})

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    pushoverStore.checkExtensionStatus()
  }
})

function close() {
  emit('close')
}

async function checkUpdates() {
  loading.value = true
  error.value = null
  try {
    await pushoverStore.checkForExtensionUpdates()
    await pushoverStore.checkExtensionStatus()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '检查更新失败'
  } finally {
    loading.value = false
  }
}

async function updateExtension() {
  loading.value = true
  error.value = null
  try {
    await pushoverStore.updateExtension()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '更新失败'
  } finally {
    loading.value = false
  }
}

async function recloneExtension() {
  if (!confirm('确定要重新下载扩展吗？这将删除当前的扩展文件。')) {
    return
  }
  loading.value = true
  error.value = null
  try {
    await pushoverStore.recloneExtension()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '下载失败'
  } finally {
    loading.value = false
  }
}

async function openFolder() {
  try {
    await OpenConfigFolder()
  } catch (e: unknown) {
    console.error('打开文件夹失败:', e)
  }
}

function openGitHub() {
  window.open('https://github.com/allanpk716/cc-pushover-hook', '_blank')
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
  backdrop-filter: blur(4px);
}

.dialog {
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  width: 500px;
  max-width: 90vw;
  box-shadow: var(--shadow-xl);
  animation: dialog-in 0.3s ease-out;
}

@keyframes dialog-in {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg) var(--space-xl);
  border-bottom: 1px solid var(--glass-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.header-title .icon {
  font-size: 20px;
}

.header-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.close-btn:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.dialog-content {
  padding: var(--space-xl);
}

.status-card,
.version-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  margin-bottom: var(--space-md);
}

.status-row,
.version-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-sm);
}

.status-row:last-child,
.version-row:last-child {
  margin-bottom: 0;
}

.label {
  font-size: 13px;
  color: var(--text-secondary);
}

.value {
  font-size: 13px;
  color: var(--text-primary);
  font-family: var(--font-mono);
}

.value.success {
  color: var(--accent-success, #10b981);
}

.value.error {
  color: var(--accent-error, #ef4444);
}

.value.path {
  cursor: pointer;
  text-decoration: underline;
  text-decoration-style: dotted;
}

.value.path:hover {
  color: var(--accent-primary);
}

.version-diff {
  margin-top: var(--space-md);
  text-align: center;
}

.diff-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning, #f59e0b);
}

.diff-badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--accent-success, #10b981);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-lg);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid var(--border-default);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-elevated);
  border-color: var(--border-hover);
}

.btn-primary {
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  color: white;
  border-color: transparent;
}

.btn-primary:hover:not(:disabled) {
  box-shadow: var(--glow-primary);
}

.btn-link {
  background: none;
  border-color: transparent;
  color: var(--accent-primary);
}

.btn-link:hover {
  text-decoration: underline;
}

.external-icon {
  font-size: 12px;
}
</style>
```

**Step 2: Commit**

```bash
git add frontend/src/components/ExtensionInfoDialog.vue
git commit -m "feat(component): 创建扩展信息对话框组件"
```

---

## Task 7: 前端 - 集成到 App.vue

**Files:**
- Modify: `frontend/src/App.vue:16-25` (工具栏区域)
- Modify: `frontend/src/App.vue:51-65` (script setup 区域)

**Step 1: 在工具栏添加扩展状态按钮**

在 `App.vue` 的工具栏中，在"设置"按钮前添加扩展状态按钮：

```vue
      <div class="toolbar-actions">
        <button @click="openAddProject" class="btn btn-primary">
          <span class="icon">＋</span>
          <span>添加项目</span>
        </button>
        <!-- 扩展状态按钮 -->
        <ExtensionStatusButton @open="extensionDialogOpen = true" />
        <button @click="openSettings" class="btn btn-secondary">
          <span class="icon">⚙</span>
          <span>设置</span>
        </button>
      </div>
```

**Step 2: 添加对话框组件**

在模板的 SettingsDialog 后添加：

```vue
    <!-- Settings Dialog -->
    <SettingsDialog v-model="settingsOpen" />

    <!-- Extension Info Dialog -->
    <ExtensionInfoDialog :open="extensionDialogOpen" @close="extensionDialogOpen = false" />
```

**Step 3: 在 script 中导入组件和状态**

在 script setup 中添加：

```typescript
import ExtensionStatusButton from './components/ExtensionStatusButton.vue'
import ExtensionInfoDialog from './components/ExtensionInfoDialog.vue'
```

**Step 4: 添加对话框状态**

在 `const settingsOpen = ref(false)` 后添加：

```typescript
const extensionDialogOpen = ref(false)
```

**Step 5: 移除调试代码**

删除 `onMounted` 中的调试代码（第 68-73 行）：

```typescript
onMounted(async () => {
  await projectStore.loadProjects()
})
```

**Step 6: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat(ui): 集成扩展状态指示器和对话框"
```

---

## Task 8: 测试完整功能

**Files:** None (集成测试)

**Step 1: 启动开发服务器**

```bash
wails dev
```

Expected: 应用启动，工具栏显示扩展状态按钮

**Step 2: 测试场景**

1. **扩展已下载且最新**
   - 指示器显示绿色圆点 + ✓
   - 点击显示"已是最新版本"
   - 按钮：检查更新、GitHub 链接

2. **扩展有更新可用**
   - 模拟：手动修改 `extensions/cc-pushover-hook/VERSION`
   - 指示器显示橙色圆点 + ↑
   - 点击显示版本差异
   - "更新扩展"按钮可用

3. **扩展未下载**
   - 删除 `extensions/cc-pushover-hook` 目录
   - 指示器显示红色圆点 + !
   - 点击显示"未下载"
   - "下载扩展"按钮可用

4. **重新下载功能**
   - 确认对话框显示
   - 确认后删除并重新克隆

5. **错误处理**
   - 断网状态下检查更新
   - 显示友好的错误提示

**Step 3: 浏览器测试**

使用开发者技能检查：
- DOM 结构正确
- 事件监听正确
- 状态更新流畅

**Step 4: Commit**

```bash
git commit --allow-empty -m "test(pushover): 完成扩展状态可视化功能测试"
```

---

## 验收标准

- [ ] 工具栏显示扩展状态指示器
- [ ] 状态颜色正确反映扩展状态（绿/橙/红）
- [ ] 点击指示器打开详细信息对话框
- [ ] 对话框显示扩展下载状态、路径、版本信息
- [ ] "检查更新"功能正常工作
- [ ] "更新扩展"在有更新时可用并成功更新
- [ ] "下载扩展"在未下载时可用并成功下载
- [ ] "重新下载"先确认再删除并重新克隆
- [ ] "在 GitHub 查看"打开正确页面
- [ ] 所有操作有 Loading 状态反馈
- [ ] 错误情况有友好提示
- [ ] UI 风格与现有设计一致

---

## 后续优化（可选）

1. 添加后台自动检查更新（应用启动时）
2. 添加更新通知横幅（有更新时自动显示）
3. 添加操作日志显示区域
4. 支持切换到不同的扩展版本
