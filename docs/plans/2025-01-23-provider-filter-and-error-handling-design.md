# Provider 过滤与错误处理改进设计

**日期**: 2025-01-23
**状态**: 设计已完成，待实现

## 问题概述

### 问题 1：Provider 过滤
当前在 `config.yaml` 中只配置了一个服务商，但前端界面所有支持的服务商都可以选择。用户选择未配置的 provider 后会导致调用失败。

### 问题 2：错误提示不完善
使用出错时需要在软件界面上有清晰的弹出提示，让用户知道具体出了什么问题。

## 设计目标

1. **Provider 过滤**：前端显示所有支持的 provider，但对未配置的进行禁用处理并提示原因
2. **错误提示改进**：确保所有错误都能通过界面上的 error-banner 正确显示

## 架构设计

### 后端改动

#### 新增 API 方法：`GetConfiguredProviders()`

**位置**: `app.go`

```go
// ProviderInfo 表示 provider 的配置状态
type ProviderInfo struct {
    Name      string `json:"name"`
    Configured bool   `json:"configured"`
    Reason     string `json:"reason,omitempty"` // 未配置的原因
}

// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (a *App) GetConfiguredProviders() ([]ProviderInfo, error)
```

**实现逻辑**:
1. 获取 config.yaml 中的 `Providers` map
2. 获取 registry 中所有已注册的 providers
3. 检查每个 provider 是否有有效配置（API Key 或 BaseURL）
4. 返回所有 providers 的配置状态

#### ConfigService 辅助方法

**位置**: `pkg/service/config_service.go`

```go
// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (s *ConfigService) GetConfiguredProviders(cfg *config.Config) []ProviderInfo
```

#### CommitService 错误处理增强

**位置**: `pkg/service/commit_service.go`

在 `GenerateCommit` 方法中添加配置检查：
- 检查 provider 是否已配置
- 检查 API Key 是否存在
- 通过 `commit-error` 事件发送友好的错误信息

### 前端改动

#### 类型定义

**位置**: `frontend/src/types/index.ts`

```typescript
export interface ProviderInfo {
  name: string           // provider 名称，如 'openai'
  configured: boolean    // 是否已配置
  reason?: string        // 未配置的原因
}
```

#### commitStore 扩展

**位置**: `frontend/src/stores/commitStore.ts`

```typescript
// 新增 state
const availableProviders = ref<ProviderInfo[]>([])

// 新增方法
async function loadAvailableProviders() {
  try {
    const result = await GetConfiguredProviders()
    availableProviders.value = result as ProviderInfo[]
  } catch (e) {
    console.error('Failed to load providers:', e)
  }
}
```

#### CommitPanel.vue 改造

**位置**: `frontend/src/components/CommitPanel.vue`

将硬编码的 provider options（第 88-100 行）改为动态渲染：

```vue
<select v-model="commitStore.provider" class="setting-select" ...>
  <option
    v-for="p in commitStore.availableProviders"
    :key="p.name"
    :value="p.name"
    :disabled="!p.configured"
  >
    {{ getProviderDisplayName(p.name) }}
    {{ !p.configured ? `(未配置: ${p.reason})` : '' }}
  </option>
</select>
```

添加显示名称映射和辅助函数：

```typescript
const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  ollama: 'Ollama',
  google: 'Google',
  phind: 'Phind'
}

function getProviderDisplayName(name: string): string {
  return PROVIDER_DISPLAY_NAMES[name] || name
}
```

在组件挂载时加载 provider 列表：

```typescript
onMounted(() => {
  commitStore.loadAvailableProviders()
})
```

## 数据流

```
Frontend                Backend
   |                       |
   |   GetConfiguredProviders()
   |---------------------->|
   |                       1. Load config.yaml
   |                       2. Get registered providers
   |                       3. Check each provider config
   |                       4. Return ProviderInfo[]
   |<----------------------|
   |                       |
Update availableProviders |
Render select options     |
with disabled state       |
```

## 错误处理流程

```
User Action              Backend
   |                       |
Generate Commit           |
   |---------------------->|
   |                       Check provider config
   |                       |
   |                 Not configured?
   |<----------------------| commit-error event
   |                       |
Show error banner         |
```

## 实现清单

### 后端任务
- [ ] 在 `app.go` 添加 `GetConfiguredProviders()` 方法
- [ ] 在 `ConfigService` 添加 `GetConfiguredProviders()` 辅助方法
- [ ] 在 `CommitService` 增强错误处理，检查 provider 配置状态
- [ ] 确保 AI 调用错误通过 `commit-error` 事件发送

### 前端任务
- [ ] 在 `types/index.ts` 添加 `ProviderInfo` 类型
- [ ] 在 `commitStore.ts` 添加 `availableProviders` state 和加载方法
- [ ] 在 `CommitPanel.vue` 将硬编码的 options 改为动态渲染
- [ ] 为禁用的 options 添加样式提示
- [ ] 在组件 onMounted 时加载 provider 列表

### 测试任务
- [ ] 测试只配置一个 provider 时，下拉框显示正确
- [ ] 测试选择未配置 provider 时的禁用状态
- [ ] 测试各种错误场景下的提示显示
- [ ] 测试 provider 配置变更后界面更新

## 设计决策

### 为什么选择禁用而不是隐藏？

1. **透明性**：用户能看到所有支持的 providers，了解产品功能
2. **引导性**：未配置的 provider 显示"未配置"提示，引导用户去配置
3. **可扩展性**：当用户添加新配置后，无需刷新即可使用

### 为什么使用 error-banner 而不是 alert？

1. **非阻塞**：错误横幅不会打断用户操作流程
2. **美观性**：与现有 UI 风格一致
3. **信息量**：可以显示更详细的错误信息

## 参考资料

- 现有配置结构：`pkg/config/config.go`
- Provider Registry：`pkg/provider/registry/registry.go`
- 前端组件：`frontend/src/components/CommitPanel.vue`
