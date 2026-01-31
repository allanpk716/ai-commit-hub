# 分支状态和终端按钮改进实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 改进分支同步状态徽章显示和终端按钮可见性

**Architecture:** 前端 Vue 组件样式修改，利用现有数据流无需后端变更

**Tech Stack:** Vue 3, TypeScript, Scoped CSS

---

### Task 1: 修改终端按钮边框颜色

**Files:**
- Modify: `frontend/src/components/ProjectStatusHeader.vue:324-347`

**Step 1: 添加 cyan 边框到终端按钮**

在 `.terminal-btn-main` 和 `.terminal-btn-dropdown` 样式中添加边框颜色：

```css
.terminal-btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  border-color: rgba(6, 182, 212, 0.4); /* 新增：cyan 边框 */
}

.terminal-btn-dropdown {
  width: 18px;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 2px;
  padding-right: 2px;
  border-color: rgba(6, 182, 212, 0.4); /* 新增：cyan 边框 */
}
```

**Step 2: 运行开发服务器验证**

Run: `wails dev`
Expected: 终端按钮显示 cyan 边框

**Step 3: 添加 hover 状态样式**

在 hover 样式中增强背景色：

```css
.icon-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

/* 新增：终端按钮 hover 特殊样式 */
.terminal-btn-main:hover,
.terminal-btn-dropdown:hover {
  background: rgba(6, 182, 212, 0.15);
  border-color: var(--accent-primary);
}
```

**Step 4: 添加图标颜色样式**

在 `.icon-btn .icon` 后添加：

```css
/* 终端按钮图标颜色 */
.terminal-btn-main .icon {
  color: var(--accent-primary);
}

.terminal-btn-dropdown .dropdown-arrow {
  color: var(--accent-primary);
}
```

**Step 5: 运行开发服务器验证**

Run: `wails dev`
Expected: hover 时按钮显示 cyan 背景，图标为 cyan 色

**Step 6: 提交**

```bash
git add frontend/src/components/ProjectStatusHeader.vue
git commit -m "feat(ui): 增强终端按钮可见性，添加 cyan 主题色边框和图标"
```

---

### Task 2: 验证分支状态徽章合并（无需修改）

**Files:**
- Verify: `frontend/src/components/ProjectStatusHeader.vue:116-148`

**Step 1: 检查现有实现**

确认 `syncStatusText` 计算属性已正确处理合并：

```typescript
const syncStatusText = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  let text = ''
  if (ahead > 0) text += `↑${ahead}`
  if (behind > 0) text += (text ? ' ' : '') + `↓${behind}`
  return text  // 输出示例: "↑3 ↓2"
})
```

确认 `syncStatusClass` 已正确处理分歧状态：

```typescript
const syncStatusClass = computed(() => {
  if (!syncStatus.value) return ''
  const { ahead, behind } = syncStatus.value
  if (ahead > 0 && behind === 0) return 'status-ahead'
  if (behind > 0 && ahead === 0) return 'status-behind'
  return 'status-diverged'  // ahead > 0 && behind > 0
})
```

**Step 2: 确认样式已支持**

确认 `.sync-status-badge.status-diverged` 样式存在：

```css
.sync-status-badge.status-diverged {
  background: rgba(239, 68, 68, 0.2);
  color: var(--accent-error);
  border: 1px solid rgba(239, 68, 68, 0.3);
}
```

**Step 3: 功能验证**

由于现有代码已支持，无需修改。确认即可。

---

### Task 3: 手动测试

**Files:**
- None (手动验证)

**Step 1: 测试分支状态显示**

1. 创建测试分支并进行推送
2. 本地提交新更改（领先状态）
3. 其他分支有新提交（落后状态）
4. 分歧状态（本地和远程都有新提交）

验证：
- 领先 → 绿色徽章 `↑N`
- 落后 → 橙色徽章 `↓N`
- 分歧 → 红色徽章 `↑N ↓M`
- 同步 → 无徽章

**Step 2: 测试终端按钮可见性**

1. 启动应用
2. 查看终端按钮在默认状态下的显示
3. hover 按钮查看颜色变化

验证：
- 边框显示为 cyan 色
- hover 时背景变为 cyan 半透明
- 图标颜色为 cyan

**Step 3: 提交验证完成**

```bash
git add docs/plans/2025-01-31-branch-status-terminal-button-implementation.md
git commit -m "docs: 添加分支状态和终端按钮改进实现计划"
```
