# 上次生成历史记录只读化设计

**日期**: 2026-01-26
**状态**: 设计阶段
**作者**: Claude

## 概述

将"上次生成"区域从可编辑改为纯展示，移除所有交互按钮。用户可以查看历史记录，但不能直接加载到编辑区。

## 需求

### 功能变更
- 移除"加载并编辑"按钮
- 移除 `loadHistory()` 方法
- 保持元数据显示（Provider、语言、时间戳）

### 目标
- 简化界面，减少视觉干扰
- 明确区分"历史记录"（只读）和"生成结果"（可操作）
- 避免用户混淆不同区域的用途

## 设计

### 前端修改

**文件**: `frontend/src/components/CommitPanel.vue`

**删除的模板代码**（第 227-232 行）:
```vue
<div class="history-actions">
  <button @click="loadHistory(lastHistoryItem)" class="btn-action btn-secondary btn-sm">
    <span class="icon">📝</span>
    加载并编辑
  </button>
</div>
```

**删除的 JavaScript 方法**（第 337-339 行）:
```javascript
function loadHistory(item: CommitHistory) {
  commitStore.generatedMessage = item.message
}
```

### 样式清理

**可选删除**（如果没有其他地方使用）:
- `.history-actions` (第 1286-1290 行)
- `.btn-sm` (第 1303-1306 行)

## 数据流

### 修改前
```
历史记录 → 点击按钮 → loadHistory() → 写入 generatedMessage → 显示在生成结果区
```

### 修改后
```
选择项目 → loadHistoryForProject() → 获取历史记录 → lastHistoryItem → 渲染只读卡片
```

数据流变为单向，不再有从历史记录回到生成结果的反向流。

## 边界情况

| 场景 | 处理方式 |
|------|----------|
| 空历史记录 | `v-if="lastHistoryItem"` 控制不显示区域 |
| 加载失败 | 现有 try-catch 捕获，打印控制台 |
| 长消息 | `white-space: pre-wrap` 和 `word-break` 自动换行 |

## 测试

### 功能测试
1. 选择有历史的项目，确认区域显示
2. 选择无历史的项目，确认区域隐藏
3. 切换项目，确认记录更新

### UI 测试
1. 确认按钮已移除
2. 确认布局无错乱

### 回归测试
1. 生成并提交新 commit，确认历史保存
2. 确认其他功能不受影响

## 影响评估

### 优点
- 界面更简洁
- 职责划分更清晰
- 减少用户混淆

### 缺点
- 用户需要手动复制历史消息
- 增加一步操作（复制 → 粘贴）

### 用户影响
- 符合典型使用流程：查看历史 → 手动复制 → 重新生成
- 避免误操作将历史内容覆盖当前生成

## 实施

### 步骤
1. 删除模板中的按钮区块
2. 删除 `loadHistory()` 方法
3. 删除不再使用的样式
4. 测试验证

### 风险
- 低风险：仅移除功能，不修改核心逻辑
- 可快速回滚：删除的代码量少

## 后续考虑

如需恢复加载功能，可添加更明确的交互：
- 右键菜单"复制并加载"
- 独立的"使用历史模板"按钮
- 历史记录侧边栏，支持拖拽到生成区
