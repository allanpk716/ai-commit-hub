# 项目级别 AI 配置功能设计文档

**日期**: 2025-01-23
**作者**: AI Commit Hub Team

## 1. 概述

### 1.1 功能目标

允许为不同的 Git 项目配置独立的 AI Provider 和语言设置，同时保持配置文件作为全局默认值。当项目配置与配置文件不一致时，提供用户友好的重置机制。

### 1.2 核心需求

- **固定项目配置**: 每个项目可以有自己的 AI 配置
- **默认值继承**: 新项目默认使用配置文件的全局设置
- **配置验证**: 检测项目配置是否与配置文件一致
- **自动重置**: 配置不一致时，用户确认后自动重置为默认值
- **立即保存**: 修改配置后立即保存到数据库

### 1.3 架构方案

采用**数据库优先 + 配置文件兜底**的架构：

1. 项目配置存储在数据库的 `git_projects` 表中
2. 读取时优先使用数据库配置，为空则使用配置文件默认值
3. 配置验证在切换项目时触发
4. 不一致时显示警告，用户确认后自动重置

---

## 2. 数据结构设计

### 2.1 数据库模型

```go
// pkg/models/git_project.go
type GitProject struct {
    ID        uint   `gorm:"primaryKey" json:"id"`
    Path      string `gorm:"not null;uniqueIndex" json:"path"`
    Name      string `json:"name"`
    SortOrder int    `gorm:"index" json:"sort_order"`

    // 项目级别 AI 配置（可选）
    Provider   *string `json:"provider,omitempty"`    // nil 表示使用默认
    Language   *string `json:"language,omitempty"`    // nil 表示使用默认
    Model      *string `json:"model,omitempty"`       // nil 表示使用默认
    UseDefault bool    `gorm:"default:true" json:"use_default"` // true=使用默认配置
}
```

### 2.2 前端类型

```typescript
// frontend/src/types/index.ts
export interface GitProject {
  id: number
  path: string
  name: string
  sort_order: number
  created_at?: string
  updated_at?: string

  // 项目 AI 配置（可选）
  provider?: string | null
  language?: string | null
  model?: string | null
  use_default?: boolean
}

export interface ProjectAIConfig {
  provider: string
  language: string
  model?: string
  isDefault: boolean
}
```

---

## 3. 后端 API 设计

### 3.1 新增 API 方法

| 方法 | 参数 | 返回值 | 描述 |
|------|------|--------|------|
| `GetProjectAIConfig(projectID int)` | 项目 ID | `ProjectAIConfig` | 获取项目的有效 AI 配置 |
| `UpdateProjectAIConfig(projectID int, provider, language, model string, useDefault bool)` | 配置参数 | `error` | 更新项目的 AI 配置 |
| `ValidateProjectConfig(projectID int)` | 项目 ID | `valid bool, resetFields []string, suggestedConfig map, error` | 验证项目配置 |
| `ConfirmResetProjectConfig(projectID int)` | 项目 ID | `error` | 确认并重置为默认 |

### 3.2 配置解析逻辑

```
1. 从数据库读取项目配置
2. 如果 UseDefault = true 或配置字段为 nil
   → 使用配置文件默认值
3. 否则
   → 使用数据库中的配置值
4. 返回合并后的配置
```

### 3.3 配置验证逻辑

```
1. 检查 Provider 是否在配置文件中存在
2. 检查 Language 是否有效（zh/en/chinese/english）
3. 如果有任何无效项
   → 返回 valid=false 和建议的默认配置
4. 否则
   → 返回 valid=true
```

---

## 4. 前端交互设计

### 4.1 UI 改进

**CommitPanel 组件新增元素**：

1. **自定义配置标记**: 当项目使用自定义配置时显示
2. **恢复默认按钮**: 一键恢复为默认配置
3. **配置不一致警告**: 检测到无效配置时显示警告横幅
4. **立即保存**: 修改下拉框后立即保存

### 4.2 用户流程

```
用户切换项目
    ↓
加载项目配置
    ↓
验证配置有效性
    ↓
    ├─ 有效 → 显示配置
    └─ 无效 → 显示警告 + 确认重置按钮
              ↓
           用户点击确认
              ↓
           重置为默认配置
```

### 4.3 状态管理

**commitStore 新增状态**：

```typescript
const selectedProjectId = ref<number>(0)
const isDefaultConfig = ref(true)
const configValidation = ref<{
  valid: boolean
  resetFields: string[]
  suggestedConfig?: ProjectAIConfig} | null>(null)
```

**commitStore 新增方法**：

- `loadProjectAIConfig(projectId: number)` - 加载项目配置
- `saveProjectConfig(projectId: number)` - 立即保存配置
- `confirmResetConfig(projectId: number)` - 确认重置

---

## 5. 数据迁移

### 5.1 数据库迁移

```go
func MigrateAddProjectAIConfig(db *gorm.DB) error {
    return db.AutoMigrate(&models.GitProject{})
}
```

GORM AutoMigrate 会自动添加新字段，现有项目会使用默认值。

### 5.2 向后兼容

- 现有项目自动设置 `UseDefault = true`
- 前端可选字段兼容旧数据
- API 保留现有行为

---

## 6. 测试策略

### 6.1 单元测试

- **服务层**: 配置解析、验证逻辑
- **Repository**: CRUD 操作
- **API**: 各个端点的输入输出

### 6.2 集成测试

- 完整的配置读取流程
- 配置不一致的检测和重置
- 多项目配置切换

### 6.3 端到端测试

- 添加新项目使用默认配置
- 修改项目配置并验证保存
- 配置不一致时的警告和重置

---

## 7. 风险与缓解

| 风险 | 缓解措施 |
|------|---------|
| 数据库迁移失败 | 使用事务，测试环境先验证 |
| 配置文件损坏 | 硬编码默认值作为后备 |
| 并发保存冲突 | 前端防抖，限制保存频率 |
| 前后端类型不同步 | TypeScript 严格模式 |

---

## 8. 成功指标

- [ ] 所有单元测试通过
- [ ] 所有集成测试通过
- [ ] 新项目默认使用配置文件设置
- [ ] 可以为项目设置自定义配置
- [ ] 切换项目时正确加载对应配置
- [ ] 配置不一致时显示警告并可以重置
- [ ] 修改配置后立即保存到数据库
- [ ] 恢复默认功能正常工作
