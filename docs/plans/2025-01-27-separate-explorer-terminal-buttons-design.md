# 分离文件管理器和终端功能设计文档

**日期**: 2025-01-27
**作者**: Claude Code
**状态**: 设计完成

---

## 一、需求概述

### 当前问题
在 `CommitPanel.vue` 中点击"📁"文件夹图标会弹出下拉菜单，同时包含"在文件管理器中打开"和"在终端中打开"两个功能，导致功能混杂。

### 设计目标
将文件管理器和终端功能分离为两个独立的按钮：
1. **文件夹按钮** → 只打开文件管理器
2. **终端按钮** → 复合设计，主体直接打开，右侧箭头选择终端类型

---

## 二、UI 设计

### 布局结构

```
┌─────────────────────────────────────────────────┐
│ 当前状态                     [分支⑂main] [📁][_>_|[🔄] │
└─────────────────────────────────────────────────┘
                                   ↑    ↑     ↑
                                   文件夹 终端 刷新
                                         ↓
                                    主体+下拉
```

### 终端按钮详细结构

```
┌──────────────┬──┐
│   _>_        │ ▼│  ← 下拉箭头
└──────────────┴──┘
     ↑
  主体区域（点击直接打开上次选择的终端）
```

---

## 三、技术实现

### 1. 组件模板

```vue
<!-- 操作按钮组 -->
<div class="action-buttons-inline">
  <!-- 文件夹按钮：只打开文件管理器 -->
  <button @click="openInExplorer" class="icon-btn" title="在文件管理器中打开">
    <span class="icon">📁</span>
  </button>

  <!-- 终端按钮：复合设计 -->
  <div class="terminal-button-wrapper">
    <button @click="openInTerminalDirectly" class="icon-btn terminal-btn-main" title="在终端中打开">
      <span class="icon">_>_</span>
    </button>
    <button @click.stop="toggleTerminalMenu" class="icon-btn terminal-btn-dropdown" title="选择终端类型">
      <span class="dropdown-arrow">▼</span>
    </button>
    <!-- 下拉菜单 -->
    <div v-if="showTerminalMenu" class="dropdown-menu terminal-menu">
      <div class="menu-header">在终端中打开</div>
      <div v-for="terminal in availableTerminals" :key="terminal.id"
           @click="openInTerminal(terminal.id)" class="menu-item">
        <span class="menu-icon">{{ terminal.icon }}</span>
        <span>{{ terminal.name }}</span>
        <span v-if="preferredTerminal === terminal.id" class="check-mark">✓</span>
      </div>
    </div>
  </div>

  <!-- 刷新按钮 -->
  <button @click.stop="handleRefresh" class="icon-btn" title="刷新状态">
    <span class="icon">🔄</span>
  </button>
</div>
```

### 2. 核心逻辑

#### 直接打开终端（新增）

```typescript
async function openInTerminalDirectly() {
  if (!currentProjectPath.value) return

  const terminalId = preferredTerminal.value || 'powershell'

  try {
    await OpenInTerminal(currentProjectPath.value, terminalId)
    showToast('success', '已在终端中打开')
  } catch (e) {
    const message = e instanceof Error ? e.message : '打开失败'
    showToast('error', message)
  }
}
```

#### 从菜单选择终端（修改）

```typescript
async function openInTerminal(terminalId: string) {
  if (!currentProjectPath.value) return

  try {
    await OpenInTerminal(currentProjectPath.value, terminalId)
    savePreferredTerminal(terminalId)
    showToast('success', '已在终端中打开')
    showTerminalMenu.value = false
  } catch (e) {
    const message = e instanceof Error ? e.message : '打开失败'
    showToast('error', message)
  }
}
```

### 3. 样式设计

```css
.terminal-button-wrapper {
  display: flex;
  position: relative;
}

.terminal-btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  padding-right: 6px;
}

.terminal-btn-dropdown {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding-left: 6px;
  padding-right: 6px;
  font-size: 12px;
}

.terminal-btn-main:hover {
  background: var(--bg-elevated);
  color: var(--text-primary);
}

.terminal-btn-dropdown:hover {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-primary);
}

.terminal-menu {
  right: 0;
  top: calc(100% + 4px);
  min-width: 180px;
}
```

---

## 四、数据流设计

### 用户偏好存储流程

```
用户选择终端 → savePreferredTerminal(id) → localStorage → 更新响应式变量
```

### 终端打开流程

**直接打开**: 点击主体 → 读取偏好 → OpenInTerminal()
**菜单选择**: 点击下拉 → 显示菜单 → 选择 → 保存偏好 → OpenInTerminal()

---

## 五、错误处理

| 错误场景 | 处理方式 |
|---------|---------|
| 未选择项目 | showToast "请先选择项目" |
| 终端未安装 | showToast 显示具体错误 |
| 权限不足 | showToast "权限不足" |
| 其他异常 | showToast "打开失败: {详情}" |

---

## 六、测试验证

### 功能测试
- [ ] 文件夹按钮只打开文件管理器
- [ ] 终端主体直接打开上次选择的终端
- [ ] 下拉箭头显示终端菜单
- [ ] 选择终端后保存偏好
- [ ] 已选终端显示 ✓

### 交互测试
- [ ] 点击外部关闭菜单
- [ ] 悬停效果正确

### 边界测试
- [ ] 未选择项目时提示错误
- [ ] localStorage 清空后使用默认

---

## 七、实现文件

| 文件 | 修改内容 |
|------|---------|
| `frontend/src/components/CommitPanel.vue` | 拆分按钮、新增逻辑、样式调整 |
