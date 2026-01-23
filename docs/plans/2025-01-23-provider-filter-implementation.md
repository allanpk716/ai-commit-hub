# Provider 过滤与错误处理改进实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 实现前端 provider 下拉框动态过滤（显示所有但禁用未配置的）和完善的错误提示机制。

**架构:** 后端新增 `GetConfiguredProviders()` API 返回所有 provider 的配置状态；前端 commitStore 加载配置信息，CommitPanel 组件动态渲染下拉选项。

**技术栈:** Go 1.21+ + Wails v2（后端），Vue 3 + TypeScript + Pinia（前端），SQLite + GORM（数据库）

---

## 前置准备

### Task 0: 验证开发环境

**检查项目结构**
- Run: `ls -la`
- Expected: 看到 `app.go`, `frontend/`, `pkg/` 等目录

**启动开发服务器（可选，用于测试）**
- Run: `wails dev`
- Expected: 开发服务器启动，无错误输出

---

## 第一阶段：后端实现

### Task 1: 添加 ProviderInfo 类型定义

**Files:**
- Create: `pkg/models/provider.go`

**Step 1: 创建 ProviderInfo 模型文件**

```go
package models

// ProviderInfo 表示 provider 的配置状态
type ProviderInfo struct {
    Name      string `json:"name"`
    Configured bool   `json:"configured"`
    Reason     string `json:"reason,omitempty"` // 未配置的原因（如"缺少 API Key"）
}
```

**Step 2: 提交**

```bash
git add pkg/models/provider.go
git commit -m "feat: 添加 ProviderInfo 模型类型

- 新增 ProviderInfo 结构体，用于表示 provider 的配置状态
- 包含名称、是否已配置、未配置原因等字段

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 2: 在 ConfigService 添加 GetConfiguredProviders 方法

**Files:**
- Modify: `pkg/service/config_service.go`

**Step 1: 添加导入**

在文件开头添加：
```go
import (
    // ... 现有导入 ...
    "github.com/allanpk716/ai-commit-hub/pkg/models"
)
```

**Step 2: 实现 GetConfiguredProviders 方法**

在 `config_service.go` 末尾添加：

```go
// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (s *ConfigService) GetConfiguredProviders(cfg *config.Config) []models.ProviderInfo {
    // 获取所有已注册的 providers
    registeredProviders := registry.Names()

    result := make([]models.ProviderInfo, 0, len(registeredProviders))

    for _, name := range registeredProviders {
        info := models.ProviderInfo{
            Name: name,
        }

        // 检查该 provider 是否在 config 中配置
        if cfg.Providers == nil {
            info.Configured = false
            info.Reason = "未在配置文件中添加"
            result = append(result, info)
            continue
        }

        providerSettings, exists := cfg.Providers[name]
        if !exists {
            info.Configured = false
            info.Reason = "未在配置文件中添加"
            result = append(result, info)
            continue
        }

        // 检查是否需要 API Key
        requiresKey := registry.RequiresAPIKey(name)

        // 验证配置完整性
        var reason string
        configured := true

        if requiresKey && providerSettings.APIKey == "" {
            configured = false
            reason = "缺少 API Key"
        } else if providerSettings.BaseURL == "" && name != "openai" && name != "anthropic" {
            // 某些 providers 有默认 BaseURL，不需要检查
            if name == "ollama" || name == "deepseek" || name == "google" || name == "phind" {
                if providerSettings.BaseURL == "" {
                    configured = false
                    reason = "缺少 BaseURL"
                }
            }
        }

        info.Configured = configured
        if !configured {
            info.Reason = reason
        }

        result = append(result, info)
    }

    return result
}
```

**Step 3: 提交**

```bash
git add pkg/service/config_service.go
git commit -m "feat: ConfigService 添加 GetConfiguredProviders 方法

- 新增方法返回所有 providers 的配置状态
- 检查 provider 是否在 config.yaml 中配置
- 检查 API Key 和 BaseURL 是否完整
- 返回配置状态和未配置原因

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 3: 在 App 添加 GetConfiguredProviders API 方法

**Files:**
- Modify: `app.go`

**Step 1: 在 app.go 末尾添加方法**

```go
// GetConfiguredProviders 返回所有支持的 providers 及其配置状态
func (a *App) GetConfiguredProviders() ([]models.ProviderInfo, error) {
    if a.initError != nil {
        return nil, a.initError
    }

    cfg, err := a.configService.LoadConfig(a.ctx)
    if err != nil {
        return nil, fmt.Errorf("加载配置失败: %w", err)
    }

    providers := a.configService.GetConfiguredProviders(cfg)
    return providers, nil
}
```

**Step 2: 提交**

```bash
git add app.go
git commit -m "feat: 添加 GetConfiguredProviders API 方法

- 新增导出方法供前端调用
- 返回所有 providers 的配置状态
- 包含错误处理

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**Step 3: 重新生成 Wails 绑定**

运行：`wails dev` 并等待绑定生成完成
Expected: 看到 `wailsbindings.go` 或 `bindings.go` 生成

---

### Task 4: 增强 CommitService 错误处理

**Files:**
- Modify: `pkg/service/commit_service.go`

**Step 1: 在 GenerateCommit 方法开始处添加配置检查**

找到 `GenerateCommit` 方法，在创建 AI client 之前添加检查：

```go
func (s *CommitService) GenerateCommit(projectPath, provider, language string) error {
    // ... 现有代码 ...

    // 加载配置检查 provider 是否已配置
    cfg, err := s.configService.LoadConfig(s.ctx)
    if err != nil {
        runtime.EventsEmit(s.ctx, "commit-error", fmt.Sprintf("加载配置失败: %v", err))
        return fmt.Errorf("加载配置失败: %w", err)
    }

    // 检查 provider 是否已配置
    providers := s.configService.GetConfiguredProviders(cfg)
    providerConfigured := false
    for _, p := range providers {
        if p.Name == provider && p.Configured {
            providerConfigured = true
            break
        }
    }

    if !providerConfigured {
        errMsg := fmt.Sprintf("Provider '%s' 未配置，请先在配置文件中添加", provider)
        runtime.EventsEmit(s.ctx, "commit-error", errMsg)
        return fmt.Errorf("provider not configured: %s", provider)
    }

    // ... 继续现有代码 ...
}
```

**Step 2: 提交**

```bash
git add pkg/service/commit_service.go
git commit -m "feat: 增强 CommitService 错误处理

- 在生成 commit 前检查 provider 是否已配置
- 未配置时发送友好的错误提示
- 通过 commit-error 事件通知前端

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 第二阶段：前端实现

### Task 5: 添加 ProviderInfo 类型定义

**Files:**
- Modify: `frontend/src/types/index.ts`

**Step 1: 在 types/index.ts 末尾添加类型**

```typescript
// Provider 配置信息
export interface ProviderInfo {
  name: string           // provider 名称，如 'openai'
  configured: boolean    // 是否已配置
  reason?: string        // 未配置的原因
}
```

**Step 2: 提交**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(types): 添加 ProviderInfo 类型定义

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 6: 在 commitStore 添加 provider 列表加载

**Files:**
- Modify: `frontend/src/stores/commitStore.ts`

**Step 1: 导入新类型**

在文件开头的 import 部分：
```typescript
import type { ProjectStatus, ProjectAIConfig, ProviderInfo } from '../types'
```

**Step 2: 导入 GetConfiguredProviders API**

在 import 列表中添加：
```typescript
import {
  GetProjectStatus,
  GenerateCommit,
  GetProjectAIConfig,
  UpdateProjectAIConfig,
  ValidateProjectConfig,
  ConfirmResetProjectConfig,
  GetConfiguredProviders  // 新增
} from '../../wailsjs/go/main/App'
```

**Step 3: 添加 state**

在 `useCommitStore` 定义中，`const error = ...` 之后添加：
```typescript
const error = ref<string | null>(null)

// Provider 列表
const availableProviders = ref<ProviderInfo[]>([])
```

**Step 4: 添加加载方法**

在 `loadProjectAIConfig` 函数之后添加：
```typescript
async function loadAvailableProviders() {
  try {
    const result = await GetConfiguredProviders()
    availableProviders.value = result as ProviderInfo[]
  } catch (e) {
    console.error('Failed to load providers:', e)
    // 失败时使用空数组，避免界面崩溃
    availableProviders.value = []
  }
}
```

**Step 5: 在 return 中导出**

在 return 语句中添加：
```typescript
return {
  // ... 现有导出 ...
  availableProviders,
  loadAvailableProviders
}
```

**Step 6: 提交**

```bash
git add frontend/src/stores/commitStore.ts
git commit -m "feat(commitStore): 添加 provider 列表加载功能

- 新增 availableProviders state
- 新增 loadAvailableProviders 方法
- 导出新的 state 和方法供组件使用

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 7: 改造 CommitPanel 组件

**Files:**
- Modify: `frontend/src/components/CommitPanel.vue`

**Step 1: 添加 onMounted 导入**

在 `<script setup lang="ts">` 部分的 import 语句中添加：
```typescript
import { ref, watch, onMounted } from 'vue'
```

**Step 2: 添加 provider 显示名称映射**

在 `const DAY = ...` 之后添加：
```typescript
// Provider 显示名称映射
const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  ollama: 'Ollama',
  google: 'Google',
  phind: 'Phind'
}

// 获取 provider 显示名称
function getProviderDisplayName(name: string): string {
  return PROVIDER_DISPLAY_NAMES[name] || name
}
```

**Step 3: 替换硬编码的 provider options**

找到第 88-100 行的 `<select>` 元素，替换为：
```vue
<select
  v-model="commitStore.provider"
  class="setting-select"
  @change="handleConfigChange"
  :disabled="commitStore.isSavingConfig"
>
  <option
    v-for="p in commitStore.availableProviders"
    :key="p.name"
    :value="p.name"
    :disabled="!p.configured"
  >
    {{ getProviderDisplayName(p.name) }}
    <template v-if="!p.configured"> (未配置: {{ p.reason }})</template>
  </option>
</select>
```

**Step 4: 添加组件挂载时加载 provider 列表**

在 `handleRegenerate` 函数之后添加：
```typescript
// 组件挂载时加载 provider 列表
onMounted(() => {
  commitStore.loadAvailableProviders()
})
```

**Step 5: 为禁用的选项添加样式**

在 `<style scoped>` 部分的 `.setting-select option` 样式之后添加：
```css
.setting-select option:disabled {
  color: var(--text-muted);
  background: var(--bg-tertiary);
  opacity: 0.6;
}
```

**Step 6: 提交**

```bash
git add frontend/src/components/CommitPanel.vue
git commit -m "feat(CommitPanel): 改造 provider 下拉框为动态渲染

- 将硬编码的 options 改为从 commitStore 动态加载
- 未配置的 provider 显示为禁用状态
- 添加 provider 显示名称映射
- 组件挂载时自动加载 provider 列表
- 为禁用选项添加样式

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 第三阶段：测试

### Task 8: 手动测试

**Step 1: 启动开发服务器**

Run: `wails dev`
Expected: 应用启动，无错误

**Step 2: 测试 provider 过滤功能**

1. 打开配置文件 `~/.ai-commit-hub/config.yaml`
2. 修改 `providers` 部分，只保留一个 provider（如 openai）：
```yaml
providers:
  openai:
    apiKey: "sk-test"
    model: "gpt-4"
    baseURL: "https://api.openai.com/v1"
```
3. 保存文件，重启应用
4. 检查 provider 下拉框：
   - OpenAI 应该可选中
   - 其他 providers 应该显示但禁用，提示"未配置: 未在配置文件中添加"

**Step 3: 测试错误提示功能**

1. 尝试生成 commit（需要先 git add 一些文件）
2. 如果选择未配置的 provider（如果 UI 允许），应该看到错误横幅
3. 如果配置了 API Key 无效，应该看到 API 错误提示

**Step 4: 测试配置动态加载**

1. 添加一个新的 provider 到 config.yaml
2. 重启应用
3. 检查新 provider 是否变为可选状态

---

## 第四阶段：文档更新

### Task 9: 更新功能文档

**Files:**
- Modify: `README.md` 或相关文档

**Step 1: 添加功能说明**

在 README 或功能文档中添加：
```markdown
## Project AI 配置

每个项目可以单独配置 AI Provider：
- 在界面上选择项目后，可以自定义其 provider 和语言
- 只有在 `config.yaml` 中配置过的 providers 才能选择
- 未配置的 providers 会显示但禁用，提示未配置原因
```

**Step 2: 提交**

```bash
git add README.md
git commit -m "docs: 更新项目 AI 配置功能说明

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 完成检查清单

- [ ] ProviderInfo 模型已创建
- [ ] ConfigService.GetConfiguredProviders 已实现
- [ ] App.GetConfiguredProviders API 已添加
- [ ] CommitService 错误处理已增强
- [ ] 前端 ProviderInfo 类型已定义
- [ ] commitStore 已添加 provider 列表加载
- [ ] CommitPanel 下拉框已改为动态渲染
- [ ] 所有改动已提交
- [ ] 手动测试通过

---

## 相关资源

- 设计文档: `docs/plans/2025-01-23-provider-filter-and-error-handling-design.md`
- Provider Registry: `pkg/provider/registry/registry.go`
- 配置结构: `pkg/config/config.go`
- 前端组件: `frontend/src/components/CommitPanel.vue`
