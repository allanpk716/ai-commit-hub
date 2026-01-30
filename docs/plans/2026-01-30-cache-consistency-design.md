# StatusCache 缓存一致性增强设计

## 问题描述

未暂存的文件添加到暂存区后，偶发显示为未跟踪状态。

**根本原因**：StatusCache 中 `untrackedCount` 与 `stagingStatus.untracked.length` 存在不一致。

**触发场景**：
1. 乐观更新未同步更新 `untrackedCount`
2. 并发刷新操作导致状态覆盖
3. 网络请求部分失败时缓存状态不完整

## 解决方案

在 StatusCache 中建立多层次一致性保障机制，自动检测并修正数据不一致。

## 核心组件

### 1. 数据验证层

```typescript
// 验证缓存一致性
function isConsistent(cache: ProjectStatusCache): boolean {
  if (!cache) return true
  const expectedUntrackedCount = cache.stagingStatus?.untracked?.length ?? 0
  return cache.untrackedCount === expectedUntrackedCount
}

// 规范化缓存（自动修正）
function normalizeCache(cache: ProjectStatusCache): ProjectStatusCache {
  if (!cache) return cache
  const normalized = { ...cache }
  if (normalized.stagingStatus?.untracked) {
    normalized.untrackedCount = normalized.stagingStatus.untracked.length
  }
  return normalized
}

// 验证并修复指定缓存
function validateAndFix(path: string): void {
  const cached = cache.value[path]
  if (cached && !isConsistent(cached)) {
    cache.value[path] = normalizeCache(cached)
  }
}
```

### 2. 集成点

| 方法 | 集成位置 | 操作 |
|------|----------|------|
| `updateCache()` | 更新后 | 调用 `validateAndFix()` |
| `refresh()` | 获取数据完成后 | 调用 `validateAndFix()` |
| `updateOptimistic()` | 乐观更新后 | 调用 `validateAndFix()` |
| `preload()` | 填充缓存前 | 使用 `normalizeCache()` |

### 3. 辅助方法

```typescript
// 批量验证修复所有缓存
function validateAndFixAll(): void

// 获取缓存健康状态（调试用）
function getHealthStatus(): {
  total: number
  consistent: number
  inconsistent: string[]
}
```

## 数据流

```
用户操作
    ↓
updateOptimistic()
    ├─ 验证当前状态
    ├─ 应用乐观更新
    └─ 验证并修正
    ↓
UI 立即更新
    ↓
后端操作
    ↓
refresh()
    ├─ 并发获取状态
    └─ 验证并修正
    ↓
UI 最终同步
```

## 边界情况处理

| 场景 | 处理方式 |
|------|----------|
| 并发更新冲突 | 使用 `pendingRequests` 防重复，以后端结果为准 |
| 网络部分失败 | 保留现有缓存 + 验证修正 |
| 乐观更新回滚 | 回滚后也验证一致性 |
| 初始化时序 | 使用 `Promise.allSettled` 避免单点失败 |

## 错误处理

**策略**：自动修正 + 静默处理

- 不抛出异常，不显示错误提示
- 自动修正数据，仅开发环境输出 `console.debug`
- 极端情况兜底：`stagingStatus` 为 null 时设置 `untrackedCount = 0`

## 测试策略

### 单元测试

- `isConsistent()` 一致/不一致场景
- `normalizeCache()` 各种数据组合
- `validateAndFix()` 修正是否生效

### 集成测试

- 并发暂存操作
- 网络错误模拟
- 乐观更新回滚

### 手动测试清单

- [ ] 单个文件暂存后，未跟踪数量正确
- [ ] 批量暂存后，未跟踪数量正确
- [ ] 快速连续暂存，数量始终一致
- [ ] 暂存后取消，数量正确恢复
- [ ] 切换项目返回，状态一致
- [ ] 刷新页面，预加载数据一致

## 实施步骤

1. **核心验证逻辑**：添加验证方法，修改 `updateCache()`
2. **增强现有方法**：集成到 `refresh()`、`updateOptimistic()`、`preload()`
3. **调试支持**：添加 `validateAndFixAll()`、`getHealthStatus()`
4. **测试验证**：运行单元测试和手动测试清单

## 文件变更

| 文件 | 变更类型 |
|------|----------|
| `frontend/src/stores/statusCache.ts` | 修改 |
| `frontend/src/stores/__tests__/statusCache.spec.ts` | 修改/新增 |

## 兼容性

- 不破坏现有 API
- 不影响现有功能
- 纯内部增强，对外透明

## 风险评估

**风险等级**：低

- 验证逻辑为防御性编程
- 失败不影响现有行为
- 无破坏性变更
