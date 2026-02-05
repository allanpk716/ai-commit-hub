# 测试模式

**分析日期:** 2026-02-05

## 测试框架

### Go 测试框架

**测试运行器：**
- 标准 `go test` 工具
- 测试命令：
  ```bash
  # 运行所有测试
  go test ./... -v

  # 运行单个包测试
  go test ./pkg/git -v

  # 运行特定测试
  go test ./pkg/git -run TestCommitChanges_Success -v
  ```

**断言库：**
- 使用 `testify/assert` 进行断言
- 常用断言方法：
  ```go
  assert.NoError(t, err)        // 错误应该为空
  assert.Equal(t, expected, actual)  // 值应该相等
  assert.True(t, condition)     // 条件应该为真
  assert.Contains(t, slice, item)  // 切片应该包含元素
  ```

### Vue 前端测试框架

**测试配置：**
- **测试运行器**: Vitest
- **测试环境**: jsdom (浏览器环境模拟)
- **测试库**: @testing-library/vue
- **配置文件**: `vitest.config.ts`

**测试命令：**
```bash
# 开发模式运行
npm run test

# 单次运行
npm run test:run

# UI 模式
npm run test:ui
```

**测试配置：**
```typescript
// vitest.config.ts
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/stores/__tests__/setup.ts'],
    include: ['**/__tests__/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        '__tests__/',
        '*.config.ts'
      ]
    }
  }
})
```

## 测试文件组织

### Go 测试结构

```
pkg/
├── git/
│   ├── git.go          # 生产代码
│   ├── commit_test.go  # 单元测试
│   └── diff_test.go    # 单元测试
tests/
└── helpers/
    ├── fixtures.go     # 测试工具函数
    ├── log_capture.go   # 日志捕获工具
    └── log_assertions.go # 日志断言工具
```

**测试文件命名：**
- 单元测试：`xxx_test.go`
- 集成测试：`xxx_integration_test.go`
- 示例：`commit_test.go`、`diff_test.go`

### Vue 测试结构

```
src/
├── components/
│   ├── ProjectList.vue
│   └── __tests__/
│       └── ProjectList.spec.ts
├── stores/
│   ├── projectStore.ts
│   └── __tests__/
│       ├── setup.ts         # 测试设置
│       ├── statusCache.spec.ts  # 状态缓存测试
│       └── errorStore.spec.ts   # 错误处理测试
└── types/
    └── index.ts           # 类型定义
```

**测试文件位置：**
- 与组件同目录下的 `__tests__` 文件夹
- 测试文件使用 `.spec.ts` 或 `.test.ts` 后缀

## 测试结构

### Go 测试结构

```go
package git

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/allanpk716/ai-commit-hub/tests/helpers"
)

func TestCommitChanges_Success(t *testing.T) {
    // 1. 设置测试环境
    repo := helpers.SetupTestRepo(t)
    repo.CreateStagedChange(t, "test.txt", "content")

    originalDir, _ := os.Getwd()
    defer os.Chdir(originalDir)
    os.Chdir(repo.Path)

    // 2. 执行测试
    err := CommitChanges(context.Background(), "test: add file")

    // 3. 验证结果
    assert.NoError(t, err)
    helpers.AssertRepoClean(t, repo)
}

func TestCommitChanges_EmptyMessage(t *testing.T) {
    // 测试空消息场景
    repo := helpers.SetupTestRepo(t)
    repo.CreateStagedChange(t, "test.txt", "content")

    originalDir, _ := os.Getwd()
    defer os.Chdir(originalDir)
    os.Chdir(repo.Path)

    err := CommitChanges(context.Background(), "")
    assert.NoError(t, err)
}
```

### Vue 测试结构

```typescript
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStatusCache } from '../statusCache'

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn()
}))

describe('StatusCache Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  describe('缓存项目状态', () => {
    it('应该正确缓存项目状态', async () => {
      const store = useStatusCache()
      const testPath = '/test/project'

      // Mock API 调用
      mockGetProjectStatus.mockResolvedValue({ branch: 'main' })

      await store.refresh(testPath, { force: true })

      const cached = store.getStatus(testPath)
      expect(cached).toBeTruthy()
      expect(cached?.gitStatus?.branch).toBe('main')
    })
  })
})
```

## Mocking 框架

### Go Mocking

**依赖注入测试：**
```go
type MockGitRepository struct {
    projects []*models.GitProject
    err      error
}

func (m *MockGitRepository) GetAll() ([]*models.GitProject, error) {
    return m.projects, m.err
}

func TestGetAllProjects_Empty(t *testing.T) {
    mockRepo := &MockGitRepository{
        projects: []*models.GitProject{},
        err:      nil,
    }

    app := NewApp(mockRepo)
    projects, err := app.GetAllProjects()

    assert.NoError(t, err)
    assert.Empty(t, projects)
}
```

**命令执行测试：**
```go
func TestGitCommandMock(t *testing.T) {
    // 测试 git 命令执行
    repo := helpers.SetupTestRepo(t)

    // 使用测试工具
    repo.CreateStagedChange(t, "test.txt", "content")
    helpers.AssertHasStagedChanges(t, repo)
}
```

### Vue Mocking

**Wails API Mocking：**
```typescript
// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn()
}))

// Mock App API
const mockGetProjectStatus = vi.fn()
const mockGetStagingStatus = vi.fn()

vi.mock('../../../wailsjs/go/main/App', () => ({
  GetProjectStatus: () => mockGetProjectStatus(),
  GetStagingStatus: () => mockGetStagingStatus(),
}))

describe('项目状态获取', () => {
  it('应该正确获取项目状态', () => {
    mockGetProjectStatus.mockResolvedValue({ branch: 'main' })

    // 测试代码
    expect(mockGetProjectStatus).toHaveBeenCalled()
  })
})
```

**组件 Mocking：**
```typescript
// Mock 组件
vi.mock('./ProjectList.vue', () => ({
  default: {
    template: '<div>Mocked ProjectList</div>'
  }
}))

// Mock window 属性
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
  })),
})
```

## Fixtures 和工厂

### Go 测试 Fixtures

```go
// tests/helpers/fixtures.go
package helpers

import (
    "os"
    "path/filepath"
    "testing"
)

type TestRepo struct {
    Name string
    Path string
}

func SetupTestRepo(t *testing.T) *TestRepo {
    t.Helper()
    tempDir := t.TempDir()
    repoPath := filepath.Join(tempDir, "test-repo")

    // 创建测试仓库
    os.Mkdir(repoPath, 0755)
    RunGitCmd(t, repoPath, "init")
    RunGitCmd(t, repoPath, "config", "user.name", "Test User")
    RunGitCmd(t, repoPath, "config", "user.email", "test@example.com")

    return &TestRepo{
        Name: "test-repo",
        Path: repoPath,
    }
}

func (tr *TestRepo) CreateStagedChange(t *testing.T, filename, content string) {
    t.Helper()
    WriteFile(t, tr.Path, filename, content)
    RunGitCmd(t, tr.Path, "add", filename)
}
```

### Vue 测试 Fixtures

```typescript
// 测试数据工厂
const createMockProject = (overrides: Partial<Project> = {}): Project => ({
  id: 1,
  path: '/test/project',
  name: 'Test Project',
  sort_order: 1,
  ...overrides
})

const createMockStatus = (overrides: Partial<Status> = {}): Status => ({
  branch: 'main',
  hasChanges: true,
  stagedCount: 2,
  ...overrides
})

describe('项目数据处理', () => {
  it('应该正确处理项目数据', () => {
    const mockProject = createMockProject({
      name: 'Custom Project'
    })

    expect(mockProject.name).toBe('Custom Project')
  })
})
```

## 覆盖率

### Go 覆盖率

**运行覆盖率测试：**
```bash
# 生成覆盖率报告
go test -cover ./...

# 生成覆盖率 HTML 报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 显示覆盖率百分比
go test -covermode=count -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**覆盖率配置：**
```bash
# .gitignore 中排除覆盖率文件
coverage.out
coverage.html
```

### Vue 覆盖率

**运行覆盖率测试：**
```bash
# 运行测试并生成覆盖率报告
npm run test:run -- --coverage

# 查看覆盖率报告
npm run test:run -- --coverage --reporter=verbose
```

**覆盖率配置：**
```typescript
// vitest.config.ts
coverage: {
  provider: 'v8',
  reporter: ['text', 'json', 'html'],
  exclude: [
    'node_modules/',
    '__tests__/',
    '*.config.ts',
    'src/main.ts',
    'src/App.vue'
  ],
  thresholds: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
}
```

## 测试类型

### 单元测试

**Go 单元测试：**
```go
func TestGetCurrentBranch_Success(t *testing.T) {
    repo := helpers.SetupTestRepo(t)
    branch, err := GetCurrentBranch(context.Background())

    assert.NoError(t, err)
    assert.NotEmpty(t, branch)
    assert.Contains(t, []string{"master", "main"}, branch)
}
```

**Vue 单元测试：**
```typescript
describe('状态缓存计算属性', () => {
  it('应该正确计算已缓存路径', () => {
    const store = useStatusCache()

    store.initCache('/project1')
    store.initCache('/project2')

    const paths = store.cachedPaths
    expect(paths).toHaveLength(2)
    expect(paths).toContain('/project1')
  })
})
```

### 集成测试

**Go 集成测试：**
```go
func TestGitWorkflow_Integration(t *testing.T) {
    repo := helpers.SetupTestRepo(t)

    // 1. 创建文件
    repo.CreateStagedChange(t, "feature.txt", "new feature")

    // 2. 提交
    err := CommitChanges(context.Background(), "feat: add feature")
    assert.NoError(t, err)

    // 3. 验证提交
    msg, err := GetHeadCommitMessage(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, "feat: add feature", msg)
}
```

**Vue 集成测试：**
```typescript
describe('完整工作流测试', () => {
  it('应该正确处理项目状态更新', async () => {
    const store = useStatusCache()
    const projectPath = '/test/project'

    // 模拟项目选择
    store.initCache(projectPath)

    // 模拟状态更新
    store.updateCache(projectPath, {
      gitStatus: { branch: 'main' },
      stagingStatus: { hasChanges: true, stagedCount: 1 },
      lastUpdated: Date.now()
    })

    const status = store.getStatus(projectPath)
    expect(status?.gitStatus?.branch).toBe('main')
    expect(status?.stagingStatus?.hasChanges).toBe(true)
  })
})
```

### 错误处理测试

**Go 错误测试：**
```go
func TestCommitChanges_RepoError(t *testing.T) {
    repo := helpers.SetupTestRepo(t)
    repo.CreateStagedChange(t, "test.txt", "content")

    // 模拟磁盘错误
    originalChdir := os.Chdir
    os.Chdir = func(dir string) error {
        return fmt.Errorf("磁盘错误")
    }
    defer func() {
        os.Chdir = originalChdir
    }()

    err := CommitChanges(context.Background(), "test: add file")
    assert.Error(t, err)
}
```

**Vue 错误测试：**
```typescript
describe('错误处理', () => {
  it('应该正确处理 API 错误', async () => {
    const store = useStatusCache()
    const testPath = '/test/project'

    // Mock API 错误
    mockGetProjectStatus.mockRejectedValue(new Error('网络错误'))

    await store.refresh(testPath, { force: true })

    const status = store.getStatus(testPath)
    expect(status?.error).toContain('网络错误')
  })
})
```

## 测试最佳实践

### Go 测试最佳实践

1. **使用测试辅助函数**
```go
// 避免重复代码
func TestSomething(t *testing.T) {
    repo := helpers.SetupTestRepo(t)
    // 测试逻辑
}
```

2. **使用 defer 清理**
```go
func TestSomething(t *testing.T) {
    originalDir, _ := os.Getwd()
    defer os.Chdir(originalDir)
    os.Chdir(tempDir)
    // 测试逻辑
}
```

3. **并发安全测试**
```go
func TestConcurrentAccess(t *testing.T) {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // 测试逻辑
        }(i)
    }

    wg.Wait()
}
```

### Vue 测试最佳实践

1. **使用测试工具**
```typescript
// 使用 vi.useFakeTimers() 处理异步
await vi.runAllTimersAsync()
```

2. **清理 Mock**
```typescript
beforeEach(() => {
    vi.clearAllMocks()
})
```

3. **隔离测试**
```typescript
// 每个测试都应该是独立的
it('应该正确处理状态更新', () => {
    // 不依赖其他测试的状态
})
```

---

*测试模式分析: 2026-02-05*