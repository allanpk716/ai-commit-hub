# 分支状态和终端按钮改进设计

## 概述

两个 UI 改进：
1. 分支同步状态徽章合并 - 分歧状态下合并为单一徽章
2. 终端按钮色彩增强 - 添加 cyan 主题色提高可见性

## 现状

**分支状态徽章**：
- `ProjectStatusHeader.vue` 已实现同步状态显示
- 当 `ahead > 0 && behind > 0` 时显示两个独立徽章

**终端按钮**：
- 使用 `var(--bg-tertiary)` 背景
- 颜色较深，不够醒目

## 改进方案

### 1. 分支状态徽章合并

**行为**：
- 分歧状态合并为单一徽章，格式 `↑3 ↓2`
- 颜色编码：绿色（领先）、橙色（落后）、红色（分歧）
- 同步状态不显示徽章

**实现**：
```typescript
// syncStatusText 已支持合并格式
if (ahead > 0) text += `↑${ahead}`
if (behind > 0) text += (text ? ' ' : '') + `↓${behind}`
```

### 2. 终端按钮色彩增强

**行为**：
- 主按钮和下拉按钮添加 cyan 边框
- hover 时使用 cyan 半透明背景
- 图标使用 `var(--accent-primary)` 颜色

**实现**：
```css
.terminal-btn-main,
.terminal-btn-dropdown {
  border: 1px solid rgba(6, 182, 212, 0.4);
}

.terminal-btn-main .icon,
.terminal-btn-dropdown .dropdown-arrow {
  color: var(--accent-primary);
}
```

## 测试

**分支状态**：
- 领先 → 绿色 `↑N`
- 落后 → 橙色 `↓N`
- 分歧 → 红色 `↑N ↓M`
- 同步 → 无徽章

**视觉**：
- 按钮在不同背景下的可见性
- hover 状态颜色过渡
