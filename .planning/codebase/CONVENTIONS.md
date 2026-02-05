# 编码约定

**分析日期:** 2026-02-05

## 命名模式

### Go 代码

**文件命名：**
- 使用小写，下划线分隔：`git_helper.go`、`status_cache.go`
- 测试文件添加 `_test` 后缀：`commit_test.go`、`diff_test.go`
- 接口命名以 `er` 结尾：`Repository`、`Service`

**函数命名：**
- 导出函数使用大写开头：`GetProjectStatus()`、`CommitChanges()`
- 私有函数使用小写开头：`setupTestRepo()`、`runGitCmd()`
- 测试函数使用 `Test` 前缀：`TestCommitChanges_Success()`

**变量命名：**
- 使用驼峰命名：`gitProjectRepo`、`commitHistoryRepo`
- 布尔变量使用 `is/has/can` 前缀：`isLoading`、`hasChanges`
- 错误变量统一为 `err`

**类型定义：**
- 结构体使用名词形式：`App`、`GitProject`
- 接口使用形容词或名词：`Repository`、`Service`
- 扩展类型使用 `With` 前缀：`GitProjectWithStatus`

### TypeScript/Vue 代码

**文件命名：**
- 组件使用 PascalCase：`ProjectList.vue`、`CommitPanel.vue`
- 工具文件使用 kebab-case：`status-cache.ts`、`project-store.ts`
- 测试文件添加 `.spec.ts` 或 `.test.ts` 后缀

**函数命名：**
- 使用驼峰命名：`loadProjects()`、`handleGenerate()`
- 事件处理函数使用 `handle` 前缀：`handleRefresh()`
- 计算属性使用 `get` 前缀或直接属性名

**变量命名：**
- 使用 camelCase：`selectedPath`、`loadingState`
- 响应式变量使用 `ref()` 和 `computed()`
常量使用大写：`API_BASE_URL`

## 代码风格

### Go 代码

**格式化：**
- 使用 `go fmt` 自动格式化
- 导入分组：标准库、第三方库、本地包
- 错误处理：使用 `if err != nil` 模式

```go
// 正确的导入组织
import (
    "context"
    "fmt"
    "os"

    "github.com/WQGroup/logger"
    "github.com/allanpk716/ai-commit-hub/pkg/git"
    "github.com/allanpk716/ai-commit-hub/pkg/repository"
)

// 错误处理模式
func doSomething() error {
    err := doAnotherThing()
    if err != nil {
        logger.Errorf("操作失败: %v", err)
        return err
    }
    return nil
}
```

**注释规范：**
- 包级别注释说明主要功能
- 导出函数添加注释说明用途和参数
- 复杂逻辑添加行内注释

```go
// SetupTestRepo 创建测试 Git 仓库
func SetupTestRepo(t *testing.T) *TestRepo {
    t.Helper()
    // 创建临时目录
    tempDir := t.TempDir()
    // 初始化 Git 仓库
    RunGitCmd(t, tempDir, "init")
    return &TestRepo{Path: tempDir}
}
```

### TypeScript/Vue 代码

**格式化：**
- 使用 Prettier 进行代码格式化
- 使用 ESLint 进行代码检查
- Vue 组件使用 `<script setup>` 语法

**组件结构：**
```vue
<template>
  <div class="component-name">
    <!-- 模板内容 -->
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

// 导入的组件
import ComponentName from './ComponentName.vue'

// 响应式状态
const count = ref(0)

// 计算属性
const doubled = computed(() => count.value * 2)

// 方法定义
function increment() {
  count.value++
}
</script>

<style scoped>
.component-name {
  /* 样式定义 */
}
</style>
```

**TypeScript 类型：**
- 使用接口定义对象类型
- 泛型用于类型安全的函数
- 联合类型处理多种情况

```typescript
interface ProjectStatus {
  branch: string
  hasChanges: boolean
  stagedCount: number
}

type ProviderName = 'openai' | 'anthropic' | 'deepseek'

function getStatus(path: string): ProjectStatus | null {
  // 实现
}
```

## 导入组织

### Go 导入顺序
```go
import (
    // 标准库
    "context"
    "fmt"
    "os"
    "time"

    // 第三方库
    "github.com/WQGroup/logger"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "gorm.io/gorm"

    // 本地包
    "github.com/allanpk716/ai-commit-hub/pkg/git"
    "github.com/allanpk716/ai-commit-hub/pkg/repository"
    "github.com/allanpk716/ai-commit-hub/pkg/service"
)
```

### TypeScript 导入顺序
```typescript
// 1. Vue 相关
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

// 2. 组件导入
import ProjectList from './ProjectList.vue'
import CommitPanel from './CommitPanel.vue'

// 3. 工具和类型
import type { GitProject } from '../types'
import { useStatusCache } from './status-cache'

// 4. Wails 绑定
import { GetAllProjects, AddProject } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
```

## 错误处理

### Go 错误处理
```go
// 1. 使用统一的日志库
import "github.com/WQGroup/logger"

// 2. 错误返回模式
func (a *App) LoadProjects() error {
    if a.initError != nil {
        return a.initError
    }

    projects, err := a.gitProjectRepo.GetAll()
    if err != nil {
        logger.Errorf("加载项目失败: %v", err)
        return fmt.Errorf("加载项目失败: %w", err)
    }

    return nil
}

// 3. 测试中的错误处理
func TestSomething(t *testing.T) {
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("测试发生恐慌: %v", r)
        }
    }()
}
```

### TypeScript 错误处理
```typescript
// 1. 错误捕获和处理
async function loadProjects() {
    try {
        loading.value = true
        const result = await GetAllProjects() as GitProject[]
        projects.value = result
    } catch (error: unknown) {
        const message = error instanceof Error ? error.message : '加载失败'
        error.value = message
        console.error('Failed to load projects:', error)
    } finally {
        loading.value = false
    }
}

// 2. 自定义错误类
class ApiError extends Error {
    constructor(
        message: string,
        public statusCode: number,
        public details?: any
    ) {
        super(message)
        this.name = 'ApiError'
    }
}
```

## 日志规范

### Go 日志规范
```go
import "github.com/WQGroup/logger"

// 使用不同级别
logger.Info("应用启动")
logger.Warn("配置文件不存在，使用默认配置")
logger.Errorf("数据库连接失败: %v", err)

// 格式化日志
logger.Infof("处理项目 %s，状态: %s", projectPath, status)
```

### TypeScript 日志规范
```typescript
// 使用 console 进行调试
console.log('应用启动')
console.warn('配置文件不存在，使用默认配置')
console.error('加载失败:', error)

// 开发环境使用 debug
if (import.meta.env.DEV) {
    console.debug('调试信息:', data)
}
```

## 函数设计

### Go 函数设计
```go
// 单一职责原则
func (a *App) GetProjectStatus(ctx context.Context, path string) (*ProjectStatus, error) {
    // 获取 Git 状态
    // 获取暂存区状态
    // 合并结果
    // 返回统一状态
}

// 参数验证
func ValidateProjectPath(path string) error {
    if path == "" {
        return errors.New("项目路径不能为空")
    }
    return nil
}
```

### TypeScript 函数设计
```typescript
// 异步函数
async function generateCommitMessage(projectPath: string): Promise<string> {
    // 实现
}

// 纯函数
function calculateStatus(status: StagingStatus): boolean {
    return status.hasChanges && status.stagedCount > 0
}

// 防抖函数
const debouncedSearch = useDebounceFn(async (query: string) => {
    await searchProjects(query)
}, 300)
```

## 模块设计

### Go 模块设计
```go
// 清晰的接口定义
type Repository interface {
    Get(id uint) (*Model, error)
    Create(model *Model) error
    Update(model *Model) error
    Delete(id uint) error
}

// 依赖注入
func NewApp(
    ctx context.Context,
    repo *GitProjectRepository,
    service *ConfigService,
) *App {
    return &App{
        ctx: ctx,
        gitProjectRepo: repo,
        configService: service,
    }
}
```

### TypeScript 模块设计
```typescript
// Pinia Store 设计
export const useProjectStore = defineStore('project', () => {
    // 状态
    const projects = ref<GitProject[]>([])
    const loading = ref(false)

    // 操作
    async function loadProjects() {
        // 实现
    }

    // 计算属性
    const selectedProject = computed(() => {
        return projects.value.find(p => p.path === selectedPath.value)
    })

    return {
        projects,
        loading,
        selectedProject,
        loadProjects
    }
})

// 工具函数
export function formatDate(date: Date): string {
    return date.toLocaleString('zh-CN')
}
```

## Vue 组件规范

### 组件结构
```vue
<template>
  <div class="component-name">
    <!-- Props 和事件 -->
    <slot :name="slotName" :data="slotData" />

    <!-- 组件逻辑 -->
    <ChildComponent
      :prop="value"
      @event="handleEvent"
    />
  </div>
</template>

<script setup lang="ts">
// Props 定义
const props = defineProps<{
  title: string
  visible?: boolean
}>()

// Emits 定义
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update', value: string): void
}>()

// 组件逻辑
const count = ref(0)

function handleClick() {
  emit('update', count.value.toString())
}
</script>

<style scoped>
.component-name {
  /* 样式定义 */
}
</style>
```

### 组件命名约定
- 使用 PascalCase：`ProjectList`、`CommitPanel`
- 文件名与组件名一致：`ProjectList.vue`
- 测试文件：`ProjectList.spec.ts`

---

*编码约定分析: 2026-02-05*