# Pushover Hook UI 重设计实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将折叠式 Pushover Hook 状态卡片重设计为单行紧凑组件，集成通知状态切换功能

**Architecture:** 创建新的 Vue 单行组件替换现有折叠卡片，新增后端 API 方法处理通知状态切换（通过创建/删除 .no-pushover/.no-windows 控制文件），在应用启动时检查 Pushover 环境变量配置

**Tech Stack:** Vue 3 + TypeScript (前端), Go 1.21+ (后端), Wails v2 (前后端通信)

---

## Task 1: 创建 PushoverStatusRow.vue 组件

**Files:**
- Create: `frontend/src/components/PushoverStatusRow.vue`

**Step 1: 创建组件文件结构和模板**

创建 `frontend/src/components/PushoverStatusRow.vue`，包含单行布局模板：

```vue
<template>
  <div class="pushover-status-row">
    <div class="status-left">
      <span class="status-icon">{{ statusIcon }}</span>
      <span class="status-title">Pushover Hook</span>
      <span v-if="status?.version" class="status-version">v{{ status.version }}</span>
      <span v-if="!status?.installed" class="status-text">(未安装)</span>
    </div>

    <div v-if="status?.installed" class="notification-toggles">
      <button
        class="notify-btn"
        :class="{ active: isPushoverEnabled, disabled: !isPushoverEnabled }"
        :title="pushoverTooltip"
        :disabled="loading"
        @click="togglePushover"
      >
        <span class="notify-icon">📱</span>
      </button>
      <button
        class="notify-btn"
        :class="{ active: isWindowsEnabled, disabled: !isWindowsEnabled }"
        :title="windowsTooltip"
        :disabled="loading"
        @click="toggleWindows"
      >
        <span class="notify-icon">💻</span>
      </button>
    </div>

    <div class="status-right">
      <span v-if="isLatest && status?.installed" class="latest-badge">已是最新</span>
      <button
        v-else-if="!status?.installed"
        class="action-btn btn-primary"
        :disabled="loading"
        @click="handleInstall"
      >
        {{ loading ? '处理中...' : '安装 Hook' }}
      </button>
      <button
        v-else-if="needsUpdate"
        class="action-btn btn-update"
        :disabled="loading"
        @click="handleUpdate"
      >
        {{ loading ? '更新中...' : '更新 Hook' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { usePushoverStore } from '../stores/pushoverStore'
import type { HookStatus } from '../types/pushover'

interface Props {
  projectPath: string
  status?: HookStatus | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  install: []
  update: []
}>()

const pushoverStore = usePushoverStore()
const localLoading = ref(false)
const updateInfo = ref<{ updateAvailable: boolean; currentVersion: string; latestVersion: string } | null>(null)

// 状态图标
const statusIcon = computed(() => {
  if (!props.status?.installed) return '🔴'
  if (needsUpdate.value) return '🟡'
  return '🟢'
})

// Pushover 是否启用
const isPushoverEnabled = computed(() => {
  if (!props.status) return false
  return props.status.mode === 'enabled' || props.status.mode === 'pushover_only'
})

// Windows 是否启用
const isWindowsEnabled = computed(() => {
  if (!props.status) return false
  return props.status.mode === 'enabled' || props.status.mode === 'windows_only'
})

// 是否需要更新
const needsUpdate = computed(() => {
  if (!props.status?.installed) return false
  return props.status.version === 'unknown' ||
         (updateInfo.value?.updateAvailable)
})

// 是否是最新版本
const isLatest = computed(() => {
  return props.status?.installed &&
         props.status.version !== 'unknown' &&
         !needsUpdate.value
})

// Tooltip 文本
const pushoverTooltip = computed(() => {
  return isPushoverEnabled.value ? '点击禁用 Pushover 通知' : '点击启用 Pushover 通知'
})

const windowsTooltip = computed(() => {
  return isWindowsEnabled.value ? '点击禁用 Windows 通知' : '点击启用 Windows 通知'
})

// 切换 Pushover 通知
async function togglePushover() {
  if (localLoading.value) return
  localLoading.value = true
  try {
    await pushoverStore.toggleNotification(props.projectPath, 'pushover')
  } finally {
    localLoading.value = false
  }
}

// 切换 Windows 通知
async function toggleWindows() {
  if (localLoading.value) return
  localLoading.value = true
  try {
    await pushoverStore.toggleNotification(props.projectPath, 'windows')
  } finally {
    localLoading.value = false
  }
}

// 安装 Hook
async function handleInstall() {
  emit('install')
}

// 更新 Hook
async function handleUpdate() {
  emit('update')
}

// 检查更新
async function checkForUpdates() {
  if (!props.status?.installed) return
  try {
    updateInfo.value = await pushoverStore.checkForUpdates(props.projectPath)
  } catch (e) {
    console.error('检查更新失败:', e)
  }
}

// 监听 status 变化
watch(() => props.status, (newStatus) => {
  if (newStatus?.installed) {
    checkForUpdates()
  }
}, { immediate: true })
</script>

<style scoped>
.pushover-status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) var(--space-lg);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  gap: var(--space-md);
  margin-bottom: var(--space-md);
  transition: all var(--transition-fast);
}

.pushover-status-row:hover {
  border-color: var(--border-hover);
}

.status-left {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex: 1;
  min-width: 0;
}

.status-icon {
  font-size: 16px;
  line-height: 1;
  flex-shrink: 0;
}

.status-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

.status-version {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.status-text {
  font-size: 13px;
  color: var(--text-muted);
  white-space: nowrap;
}

.notification-toggles {
  display: flex;
  gap: var(--space-xs);
  flex-shrink: 0;
}

.notify-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 2px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
  padding: 0;
}

.notify-btn:hover:not(:disabled) {
  transform: scale(1.1);
  border-color: var(--accent-primary);
}

.notify-btn.active {
  border-color: var(--accent-primary);
  background: rgba(6, 182, 212, 0.15);
}

.notify-btn.disabled {
  opacity: 0.4;
  filter: grayscale(1);
}

.notify-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.notify-icon {
  font-size: 18px;
  line-height: 1;
}

.status-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.latest-badge {
  font-size: 12px;
  color: var(--text-muted);
  padding: var(--space-xs) var(--space-sm);
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.action-btn {
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.btn-update {
  background: rgba(245, 158, 11, 0.2);
  color: var(--accent-warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.btn-update:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.3);
}
</style>
```

**Step 2: 验证组件文件已创建**

Run: `ls frontend/src/components/PushoverStatusRow.vue`
Expected: 文件存在

**Step 3: 提交**

```bash
git add frontend/src/components/PushoverStatusRow.vue
git commit -m "feat: 创建 PushoverStatusRow 单行状态组件"
```

---

## Task 2: 后端添加 ToggleNotification API 方法

**Files:**
- Modify: `app.go` (在适当位置添加新方法)

**Step 1: 在 app.go 中添加 ToggleNotification 方法**

在 `app.go` 文件末尾（在 `initError` 检查之后）添加：

```go
// ToggleNotification 切换指定项目的通知类型
// 通过创建或删除 .no-pushover 或 .no-windows 文件来实现
func (a *App) ToggleNotification(projectPath string, notificationType string) error {
	logger.Infof("切换通知状态: 项目=%s, 类型=%s", projectPath, notificationType)

	// 检查初始化错误
	if a.initError != nil {
		return fmt.Errorf("应用未正确初始化: %w", a.initError)
	}

	// 验证项目路径
	if projectPath == "" {
		return fmt.Errorf("项目路径不能为空")
	}

	// 验证通知类型
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

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建文件来禁用通知
			file, err := os.Create(filePath)
			if err != nil {
				logger.Errorf("创建禁用文件失败: %v", err)
				return fmt.Errorf("创建禁用文件失败: %w", err)
			}
			file.Close()
			logger.Infof("已禁用 %s 通知: 创建 %s", notificationType, fileName)
			return nil
		}
		// 其他错误
		logger.Errorf("检查文件失败: %v", err)
		return fmt.Errorf("检查文件失败: %w", err)
	}

	// 文件存在，删除文件来启用通知
	if fileInfo.IsDir() {
		return fmt.Errorf("%s 是目录，不是文件", fileName)
	}

	if err := os.Remove(filePath); err != nil {
		logger.Errorf("删除禁用文件失败: %v", err)
		return fmt.Errorf("删除禁用文件失败: %w", err)
	}

	logger.Infof("已启用 %s 通知: 删除 %s", notificationType, fileName)
	return nil
}
```

**Step 2: 运行 wails dev 重新生成绑定**

Run: `wails dev`
Expected: 服务器启动，Wails 自动生成新的 JavaScript 绑定
等待: 看到 "Server started" 或类似消息后按 Ctrl+C 停止

**Step 3: 验证绑定已生成**

Run: `grep -n "ToggleNotification" wailsjs/go/main/App.js`
Expected: 找到包含 `ToggleNotification` 的行

**Step 4: 提交**

```bash
git add app.go wailsjs/go/main/App.js
git commit -m "feat: 添加 ToggleNotification API 方法"
```

---

## Task 3: 后端添加 CheckPushoverConfig API 方法

**Files:**
- Modify: `app.go`

**Step 1: 在 app.go 中添加 CheckPushoverConfig 方法**

在 `ToggleNotification` 方法后添加：

```go
// CheckPushoverConfig 检查 Pushover 环境变量是否已配置
// 返回配置状态，用于应用启动时的检查
func (a *App) CheckPushoverConfig() map[string]interface{} {
	token := os.Getenv("PUSHOVER_TOKEN")
	user := os.Getenv("PUSHOVER_USER")

	tokenSet := token != ""
	userSet := user != ""
	valid := tokenSet && userSet

	result := map[string]interface{}{
		"valid":     valid,
		"token_set": tokenSet,
		"user_set":  userSet,
	}

	if valid {
		logger.Info("Pushover 配置检查: 已配置")
	} else {
		logger.Warn("Pushover 配置检查: 未配置 (TOKEN=%t, USER=%t)", tokenSet, userSet)
	}

	return result
}
```

**Step 2: 运行 wails dev 重新生成绑定**

Run: `wails dev`
Expected: 服务器启动，Wails 自动生成新的 JavaScript 绑定
等待: 看到 "Server started" 后按 Ctrl+C 停止

**Step 3: 验证绑定已生成**

Run: `grep -n "CheckPushoverConfig" wailsjs/go/main/App.js`
Expected: 找到包含 `CheckPushoverConfig` 的行

**Step 4: 提交**

```bash
git add app.go wailsjs/go/main/App.js
git commit -m "feat: 添加 CheckPushoverConfig API 方法"
```

---

## Task 4: 扩展 pushoverStore 添加新方法

**Files:**
- Modify: `frontend/src/stores/pushoverStore.ts`

**Step 1: 在 pushoverStore 中添加 toggleNotification 方法**

在 `pushoverStore.ts` 的 actions 部分添加：

```typescript
// 切换通知状态（创建/删除控制文件）
async toggleNotification(projectPath: string, type: 'pushover' | 'windows'): Promise<{ success: boolean; message?: string }> {
  try {
    await ToggleNotification(projectPath, type)
    // 刷新状态
    await this.getProjectHookStatus(projectPath)
    return { success: true }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    console.error(`切换 ${type} 通知失败:`, error)
    return { success: false, message }
  }
}
```

**Step 2: 在 pushoverStore 中添加 checkPushoverConfig 方法**

```typescript
// 检查 Pushover 配置状态
async checkPushoverConfig(): Promise<{ valid: boolean; token_set: boolean; user_set: boolean }> {
  try {
    const result = await CheckPushoverConfig()
    this.configValid = result.valid
    return result
  } catch (error) {
    console.error('检查 Pushover 配置失败:', error)
    this.configValid = false
    return { valid: false, token_set: false, user_set: false }
  }
}
```

**Step 3: 在 pushoverStore state 中添加 configValid**

在 store 的 state 定义中添加：

```typescript
configValid: false as boolean,
```

**Step 4: 提交**

```bash
git add frontend/src/stores/pushoverStore.ts
git commit -m "feat: 扩展 pushoverStore 添加通知切换和配置检查方法"
```

---

## Task 5: 在 CommitPanel 中集成新组件

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 修改模板部分，替换旧组件**

找到 `<PushoverStatusCard>` 组件（约第 54-57 行），替换为：

```vue
    <!-- Pushover 单行状态组件 -->
    <PushoverStatusRow
      v-if="commitStore.projectStatus && currentProject"
      :project-path="currentProject.path"
      :status="pushoverStatus"
      :loading="pushoverStore.loading"
      @install="handleInstallHook"
      @update="handleUpdateHook"
    />
```

**Step 2: 修改 script 部分，更新导入**

在 import 部分找到（约第 238-240 行）：

```typescript
import PushoverStatusBadge from './PushoverStatusBadge.vue'
import PushoverStatusCard from './PushoverStatusCard.vue'
```

替换为：

```typescript
import PushoverStatusRow from './PushoverStatusRow.vue'
```

**Step 3: 添加处理方法**

在 script 中找到 `handleRegenerate` 函数后（约第 396 行后）添加：

```typescript
// 处理安装 Hook
async function handleInstallHook() {
  if (!currentProject.value) return
  const result = await pushoverStore.installHook(currentProject.value.path, false)
  if (!result.success) {
    alert('安装失败: ' + (result.message || '未知错误'))
  }
}

// 处理更新 Hook
async function handleUpdateHook() {
  if (!currentProject.value) return
  if (!confirm('确定要更新此项目的 Pushover Hook 吗？')) return
  const result = await pushoverStore.updateHook(currentProject.value.path)
  if (!result.success) {
    alert('更新失败: ' + (result.message || '未知错误'))
  }
}
```

**Step 4: 移除 section-header 中的 PushoverStatusBadge**

找到 section-header 中的 PushoverStatusBadge（约第 15-20 行）：

```vue
          <PushoverStatusBadge
            v-if="currentProject"
            :status="pushoverStatus"
            :loading="pushoverStore.loading"
            :compact="true"
          />
```

删除这部分代码。

**Step 5: 验证语法**

Run: `cd frontend && npm run type-check`
Expected: 无类型错误

**Step 6: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor: 在 CommitPanel 中集成 PushoverStatusRow 组件"
```

---

## Task 6: 应用启动时检查 Pushover 配置

**Files:**
- Modify: `frontend/src/App.vue` 或主入口文件

**Step 1: 找到应用初始化代码**

检查 `App.vue` 或 `main.ts` 中的应用初始化部分

**Step 2: 添加配置检查**

在应用初始化时添加：

```typescript
import { usePushoverStore } from './stores/pushoverStore'

const pushoverStore = usePushoverStore()

// 应用启动时检查 Pushover 配置
onMounted(async () => {
  await pushoverStore.checkPushoverConfig()
  if (!pushoverStore.configValid) {
    console.warn('Pushover 环境变量未配置，通知功能可能不可用')
  }
})
```

**Step 3: 提交**

```bash
git add frontend/src/App.vue
git commit -m "feat: 应用启动时检查 Pushover 配置"
```

---

## Task 7: 删除旧的组件文件

**Files:**
- Delete: `frontend/src/components/PushoverStatusBadge.vue`
- Delete: `frontend/src/components/PushoverStatusCard.vue`

**Step 1: 删除旧组件**

```bash
rm frontend/src/components/PushoverStatusBadge.vue
rm frontend/src/components/PushoverStatusCard.vue
```

**Step 2: 验证没有其他文件引用旧组件**

Run: `grep -r "PushoverStatusBadge\|PushoverStatusCard" frontend/src/`
Expected: 无结果（除了可能的注释）

**Step 3: 提交**

```bash
git add frontend/src/components/
git commit -m "refactor: 删除已弃用的 Pushover 状态组件"
```

---

## Task 8: 测试功能

**Files:**
- Test: 手动测试所有场景

**Step 1: 启动开发服务器**

Run: `wails dev`

**Step 2: 测试未安装状态**

操作：
1. 选择一个没有安装 Pushover Hook 的项目
2. 观察状态行是否显示 🔴 和 "安装 Hook" 按钮
3. 点击 "安装 Hook" 按钮
4. 验证安装是否成功

**Step 3: 测试通知切换**

操作：
1. 选择一个已安装 Pushover Hook 的项目
2. 点击 📱 图标
3. 验证项目目录中是否创建了 `.no-pushover` 文件
4. 再次点击 📱 图标
5. 验证 `.no-pushover` 文件是否被删除
6. 对 💻 图标重复相同操作

**Step 4: 测试更新功能**

操作：
1. 找到一个有新版本可用的项目
2. 观察状态行是否显示 🟡 和 "更新 Hook" 按钮
3. 点击 "更新 Hook" 按钮
4. 验证更新是否成功

**Step 5: 测试配置检查**

操作：
1. 设置 PUSHOVER_TOKEN 和 PUSHOVER_USER 环境变量
2. 重启应用
3. 检查控制台是否显示配置已验证的日志

**Step 6: 记录测试结果**

创建测试报告：
```
- 未安装状态: [PASS/FAIL]
- 通知切换 (Pushover): [PASS/FAIL]
- 通知切换 (Windows): [PASS/FAIL]
- 更新功能: [PASS/FAIL]
- 配置检查: [PASS/FAIL]
```

**Step 7: 修复发现的问题**

如果测试中发现问题，修复并重新测试。

**Step 8: 提交**

```bash
git add .
git commit -m "test: 完成 Pushover UI 重设计测试"
```

---

## Task 9: 更新文档

**Files:**
- Update: `CLAUDE.md` (如果需要)

**Step 1: 检查是否需要更新项目文档**

Review `CLAUDE.md` 中关于 Pushover Hook 的描述是否需要更新

**Step 2: 提交文档更新（如果有）**

```bash
git add CLAUDE.md
git commit -m "docs: 更新 Pushover Hook 相关文档"
```

---

## 验收标准

完成所有任务后，应该满足：

1. **UI 改进**
   - Pushover Hook 状态以单行形式显示
   - 状态图标清晰（🟢/🟡/🔴）
   - 通知图标可点击切换

2. **功能完整**
   - 点击通知图标切换状态（创建/删除控制文件）
   - 安装/更新按钮正常工作
   - 状态显示正确反映实际状态

3. **代码质量**
   - 无 TypeScript 类型错误
   - 无 ESLint 警告
   - 所有测试通过

4. **向后兼容**
   - 现有安装的 Hook 继续工作
   - 控制文件（.no-pushover/.no-windows）正确处理
