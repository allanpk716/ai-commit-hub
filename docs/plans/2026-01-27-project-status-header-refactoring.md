# ProjectStatusHeader 组件重构计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 将 CommitPanel 中内联的顶部状态栏（分支信息、操作按钮组、Pushover 状态条）提取为独立的 ProjectStatusHeader 组件，符合设计文档 `2026-01-27-git-staging-ui-design.md` 的架构设计。

**架构:** 从 CommitPanel.vue 中提取状态栏相关逻辑到独立的 ProjectStatusHeader.vue 组件，通过 props 传递数据，通过 events 传递操作，保持单一职责原则。

**技术栈:**
- Vue 3 Composition API
- TypeScript
- Pinia (commitStore, pushoverStore)

---

## 前置条件

**已完成:**
- ✅ CommitPanel.vue 中已有状态栏功能实现
- ✅ PushoverStatusRow 组件已存在
- ✅ 终端菜单功能已实现
- ✅ 设计文档已完成 (`docs/plans/2026-01-27-git-staging-ui-design.md`)

---

## Task 1: 创建 ProjectStatusHeader.vue 组件

**目的:** 提取状态栏功能到独立组件

**Files:**
- Create: `frontend/src/components/ProjectStatusHeader.vue`

**Step 1: 创建组件文件**

```vue
<template>
  <div class="project-status-header">
    <!-- 分支信息和操作按钮组 -->
    <div class="status-header-top">
      <div class="branch-badge">
        <span class="icon">⑂</span>
        {{ branch }}
      </div>

      <!-- 操作按钮组 -->
      <div class="action-buttons-inline">
        <!-- 文件夹按钮 -->
        <button @click="handleOpenInExplorer" class="icon-btn" title="在文件管理器中打开">
          <span class="icon">📁</span>
        </button>

        <!-- 终端按钮：复合设计 -->
        <div class="terminal-button-wrapper">
          <button @click="handleOpenInTerminalDirectly" class="icon-btn terminal-btn-main" title="在终端中打开">
            <span class="icon">_>_</span>
          </button>
          <button @click.stop="toggleTerminalMenu" class="icon-btn terminal-btn-dropdown" title="选择终端类型">
            <span class="dropdown-arrow">▼</span>
          </button>
          <!-- 下拉菜单 -->
          <div v-if="showTerminalMenu" class="dropdown-menu terminal-menu">
            <div class="menu-header">在终端中打开</div>
            <div
              v-for="terminal in availableTerminals"
              :key="terminal.id"
              @click="handleOpenInTerminal(terminal.id)"
              class="menu-item"
            >
              <span class="menu-icon">{{ terminal.icon }}</span>
              <span>{{ terminal.name }}</span>
              <span v-if="preferredTerminal === terminal.id" class="check-mark">✓</span>
            </div>
          </div>
        </div>

        <!-- 刷新按钮 -->
        <button @click="handleRefresh" class="icon-btn" title="刷新状态">
          <span class="icon">🔄</span>
        </button>
      </div>
    </div>

    <!-- Pushover 状态条 -->
    <PushoverStatusRow
      v-if="projectPath"
      :project-path="projectPath"
      :status="pushoverStatus"
      :loading="pushoverLoading"
      @install="handleInstallPushover"
      @update="handleUpdatePushover"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import PushoverStatusRow from './PushoverStatusRow.vue'

// Props
interface Props {
  branch: string
  projectPath?: string
  pushoverStatus: any
  pushoverLoading: boolean
  availableTerminals: Array<{
    id: string
    name: string
    icon: string
  }>
  preferredTerminal: string
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  openInExplorer: []
  openInTerminal: [terminalId: string]
  openInTerminalDirectly: []
  refresh: []
  installPushover: []
  updatePushover: []
}>()

// 终端菜单状态
const showTerminalMenu = ref(false)

// 切换终端菜单
function toggleTerminalMenu() {
  showTerminalMenu.value = !showTerminalMenu.value
}

// 点击外部关闭菜单
function closeTerminalMenu() {
  showTerminalMenu.value = false
}

// 事件处理函数
function handleOpenInExplorer() {
  emit('openInExplorer')
}

function handleOpenInTerminal(terminalId: string) {
  emit('openInTerminal', terminalId)
  closeTerminalMenu()
}

function handleOpenInTerminalDirectly() {
  emit('openInTerminalDirectly')
}

function handleRefresh() {
  emit('refresh')
}

function handleInstallPushover() {
  emit('installPushover')
}

function handleUpdatePushover() {
  emit('updatePushover')
}

// 暴露关闭菜单方法供父组件调用
defineExpose({
  closeTerminalMenu
})
</script>

<style scoped>
.project-status-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  padding: var(--space-md);
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
}

.status-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
}

.branch-badge {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-sm);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.branch-badge .icon {
  font-size: 14px;
}

.action-buttons-inline {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

.icon-btn .:active {
  transform: translateY(0);
}

.icon-btn .icon {
  font-size: 16px;
}

/* 终端按钮复合样式 */
.terminal-button-wrapper {
  display: flex;
  position: relative;
}

.terminal-btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
}

.terminal-btn-dropdown {
  width: 20px;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 2px;
  padding-right: 2px;
}

.dropdown-arrow {
  font-size: 8px;
  color: var(--text-secondary);
}

/* 下拉菜单样式 */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  z-index: 100;
  min-width: 180px;
  background: var(--bg-primary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.terminal-menu {
  right: 0;
}

.menu-header {
  padding: var(--space-sm) var(--space-md);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  border-bottom: 1px solid var(--border-default);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  cursor: pointer;
  transition: background 0.2s;
}

.menu-item:hover {
  background: var(--bg-hover);
}

.menu-icon {
  font-size: 14px;
  width: 20px;
  text-align: center;
}

.check-mark {
  margin-left: auto;
  color: var(--color-primary);
  font-weight: bold;
}
</style>
```

**Step 2: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 3: 提交**

```bash
git add frontend/src/components/ProjectStatusHeader.vue
git commit -m "feat(component): 创建 ProjectStatusHeader 组件"
```

---

## Task 2: 重构 CommitPanel.vue 使用新组件

**目的:** 用 ProjectStatusHeader 替换内联的状态栏代码

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue:4-55`

**Step 1: 移除内联的状态栏代码**

删除第 4-55 行的整个 section-header 和相关代码，保留 StagingArea。

**Step 2: 导入 ProjectStatusHeader 组件**

在 `<script setup>` 部分添加导入：

```vue
<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useCommitStore } from '../stores/commitStore'
import { useProjectStore } from '../stores/projectStore'
import { usePushoverStore } from '../stores/pushoverStore'
import ProjectStatusHeader from './ProjectStatusHeader.vue'
import StagingArea from './StagingArea.vue'

// ... 其余代码保持不变
```

**Step 3: 替换模板中的状态栏**

将原来的状态栏部分替换为：

```vue
    <!-- Project Info Section -->
    <section class="panel-section staging-section" v-if="commitStore.projectStatus">
      <!-- Project Status Header -->
      <ProjectStatusHeader
        :branch="commitStore.projectStatus.branch"
        :project-path="currentProject?.path"
        :pushover-status="pushoverStatus"
        :pushover-loading="pushoverStore.loading"
        :available-terminals="availableTerminals"
        :preferred-terminal="preferredTerminal"
        @open-in-explorer="openInExplorer"
        @open-in-terminal="openInTerminal"
        @open-in-terminal-directly="openInTerminalDirectly"
        @refresh="handleRefresh"
        @install-pushover="handleInstallPushover"
        @update-pushover="handleUpdatePushover"
      />

      <!-- Staging Area -->
      <StagingArea v-if="commitStore.stagingStatus" />
    </section>
```

**Step 4: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 5: 浏览器测试功能**

Run: `wails dev`
Expected:
- 分支信息正确显示
- 文件夹按钮可以打开文件管理器
- 终端按钮和菜单功能正常
- 刷新按钮可以刷新状态
- Pushover 状态条正常显示和操作

**Step 6: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(panel): 使用 ProjectStatusHeader 组件替换内联代码"
```

---

## Task 3: 清理未使用的代码

**目的:** 移除 CommitPanel 中不再需要的终端菜单相关状态和方法

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 移除 showTerminalMenu 状态**

找到 `const showTerminalMenu = ref(false)` 并删除

**Step 2: 移除 closeTerminalMenu 方法**

找到 `function closeTerminalMenu()` 并删除

**Step 3: 移除 toggleTerminalMenu 方法**

找到 `function toggleTerminalMenu()` 并删除

**Step 4: 移除点击外部关闭菜单的事件处理**

如果有相关的点击事件监听器，一并删除

**Step 5: 验证组件编译**

Run: `cd frontend && npm run type-check`
Expected: No type errors

**Step 6: 浏览器测试功能**

Run: `wails dev`
Expected: 所有功能正常工作

**Step 7: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "refactor(panel): 清理未使用的终端菜单状态和方法"
```

---

## Task 4: 优化组件样式

**目的:** 确保 ProjectStatusHeader 与整体设计一致

**Files:**
- Modify: `frontend/src/components/ProjectStatusHeader.vue`

**Step 1: 添加 CSS 变量支持**

确保组件使用项目定义的 CSS 变量

**Step 2: 调整间距和布局**

根据实际显示效果微调样式

**Step 3: 响应式设计测试**

测试不同窗口大小下的显示效果

**Step 4: 浏览器测试**

Run: `wails dev`
Expected: 样式美观，响应式正常

**Step 5: 提交**

```bash
git add frontend/src/components/ProjectStatusHeader.vue
git commit -m "style(header): 优化 ProjectStatusHeader 样式"
```

---

## Task 5: 端到端测试

**目的:** 验证重构后所有功能正常

**Step 1: 启动应用**

Run: `wails dev`
Expected: 应用启动成功

**Step 2: 测试分支显示**

1. 选择一个项目
2. 验证分支信息正确显示

**Step 3: 测试文件夹按钮**

1. 点击文件夹按钮
2. 验证文件管理器正确打开到项目目录

**Step 4: 测试终端按钮**

1. 点击终端主按钮
2. 验证终端打开
3. 点击下拉箭头
4. 验证终端菜单显示
5. 选择不同终端
6. 验证选择的终端被记住

**Step 5: 测试刷新按钮**

1. 修改项目文件
2. 点击刷新按钮
3. 验证状态更新

**Step 6: 测试 Pushover 状态条**

1. 验证 Pushover 状态正确显示
2. 测试安装功能
3. 测试更新功能

**Step 7: 测试与暂存区集成**

1. 验证 StagingArea 正常显示
2. 验证状态刷新时暂存区同步更新

**Step 8: 记录问题**

记录测试中发现的所有问题

**Step 9: 最终提交**

```bash
git add -A
git commit -m "test: 完成 ProjectStatusHeader 重构测试"
```

---

## 验收标准

- [ ] ProjectStatusHeader 组件创建成功
- [ ] CommitPanel 使用新组件重构完成
- [ ] 所有原有功能正常工作（分支显示、文件夹、终端、刷新、Pushover）
- [ ] 组件代码清晰，职责单一
- [ ] 样式与整体设计一致
- [ ] 响应式布局正常
- [ ] 无控制台错误或警告
- [ ] 通过端到端测试

---

## 参考资料

- 设计文档: `docs/plans/2026-01-27-git-staging-ui-design.md`
- 实施计划: `docs/plans/2026-01-27-git-staging-implementation.md`
- Wails 文档: https://wails.io/docs/next/introduction
