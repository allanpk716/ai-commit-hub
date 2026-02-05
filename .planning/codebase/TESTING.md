# Testing Patterns

**Analysis Date:** 2026-02-05

## Test Framework

### Go
**Runner:**
- `go test ./... -v`
- Test files: `*_test.go`
- Assertion library: `github.com/stretchr/testify`

**Run Commands:**
```bash
go test ./... -v              # Run all tests
go test -run TestMockAIClient  # Run specific test
go test -cover                 # Show coverage
```

### TypeScript
**Runner:**
- Framework: `vitest`
- Config: `frontend/vitest.config.*`
- Assertion library: `@testing-library/vue` with `vi` for mocking

**Run Commands:**
```bash
npm run test          # Run tests (vitest)
npm run test:ui       # Run with UI
npm run test:run      # Run tests without watch mode
```

## Test File Organization

### Go
**Location:**
- Service tests: `pkg/service/*_test.go`
- Git tests: `pkg/git/*_test.go`
- Repository tests: `pkg/repository/*_test.go`
- Helper tests: `tests/helpers/*_test.go`

**Structure:**
```
tests/
├── helpers/
│   ├── fixtures.go         # Test data factories
│   ├── fixtures_test.go    # Helper tests
│   ├── log_capture.go      # Logging utilities
│   └── log_assertions.go  # Logging assertions
```

### TypeScript
**Location:**
- Store tests: `frontend/src/stores/__tests__/*.spec.ts`
- Component tests: `frontend/src/components/__tests__/*.spec.ts`
- Helper tests: `frontend/src/utils/*_test.ts` (if any)

**Structure:**
```
frontend/src/
├── stores/
│   ├── __tests__/
│   │   ├── statusCache.spec.ts
│   │   ├── errorStore.spec.ts
│   │   └── setup.ts
└── components/
    └── __tests__/
        └── ErrorToast.spec.ts
```

## Test Structure

### Go Pattern
```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/allanpk716/ai-commit-hub/tests/helpers"
)

func TestMockAIClient_Basic(t *testing.T) {
    client := NewMockAIClient("test response", nil)

    response, err := client.GetCommitMessage(context.Background(), "test prompt")

    assert.NoError(t, err)
    assert.Equal(t, "test response", response)
}
```

### TypeScript Pattern
```typescript
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStatusCache } from '../statusCache'

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

  it('应该正确缓存项目状态', async () => {
    const store = useStatusCache()
    const testPath = '/test/project'

    // Mock API calls
    mockGetProjectStatus.mockResolvedValue(mockStatus.gitStatus)
    mockGetStagingStatus.mockResolvedValue(mockStatus.stagingStatus)

    await store.refresh(testPath, { force: true })
    await vi.runAllTimersAsync()

    const cached = store.getStatus(testPath)
    expect(cached).toBeTruthy()
    expect(cached?.gitStatus?.branch).toBe('main')
  })
})
```

## Mocking

### Go Mocking
**Framework:** Manual mocks with interface implementations
```go
// Mock implementation
type MockAIClient struct {
    response string
    err      error
    deltas   []string
}

func (m *MockAIClient) GetCommitMessage(ctx context.Context, prompt string) (string, error) {
    return m.response, m.err
}

func (m *MockAIClient) StreamCommitMessage(ctx context.Context, prompt string, deltaFunc func(string)) (string, error) {
    for _, delta := range m.deltas {
        deltaFunc(delta)
    }
    return m.response, m.err
}
```

### TypeScript Mocking
**Framework:** `vi` (vitest)
```typescript
// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn()
}))

// Mock App API
const mockGetProjectStatus = vi.fn()
const mockGetStagingStatus = vi.fn()

vi.mock('../../../wailsjs/go/main/App', () => ({
  GetProjectStatus: () => mockGetProjectStatus(),
  GetStagingStatus: () => mockGetStagingStatus()
}))
```

## Fixtures and Factories

### Go Test Data
```go
// tests/helpers/fixtures.go
type TestRepo struct {
    Name string
    Path string
}

func SetupTestRepo(t *testing.T) *TestRepo {
    t.Helper()
    tempDir := t.TempDir()
    repoPath := filepath.Join(tempDir, "test-repo")

    // Create directory
    if err := os.Mkdir(repoPath, 0755); err != nil {
        t.Fatalf("创建目录失败: %v", err)
    }

    // Initialize git repo
    RunGitCmd(t, repoPath, "init")
    RunGitCmd(t, repoPath, "config", "user.name", "Test User")
    RunGitCmd(t, repoPath, "config", "user.email", "test@example.com")

    return &TestRepo{Name: "test-repo", Path: repoPath}
}
```

### TypeScript Test Data
```typescript
// Test data in test files
const mockStatus: ProjectStatusCache = {
  gitStatus: { branch: 'main' },
  stagingStatus: { hasChanges: true, stagedCount: 2 },
  untrackedCount: 3,
  pushoverStatus: { enabled: true, version: '1.0.0' },
  lastUpdated: Date.now(),
  loading: false,
  error: null,
  stale: false
}
```

## Coverage

### Requirements
- **Go**: No enforced target visible in configuration
- **TypeScript**: No enforced target visible in configuration

### View Coverage
```bash
# Go
go test -cover

# TypeScript
npm run test:run -- --coverage
```

## Test Types

### Unit Tests
- **Go**: Focus on individual functions and methods
- **TypeScript**: Focus on store logic, component rendering, utility functions

### Integration Tests
- **Go**: Found in `tests/integration/`
  - API testing with real Git operations
  - Service layer integration
- **TypeScript**: Component interaction tests with real DOM

### E2E Tests
- **Not detected** in current codebase

## Common Patterns

### Async Testing
```go
// Go async testing
func TestSomethingAsync(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := someAsyncOperation(ctx)
    assert.NoError(t, err)
}
```

```typescript
// TypeScript async testing
it('应该处理异步操作', async () => {
    const promise = someAsyncOperation()

    await expect(promise).resolves.toBeUndefined()
})
```

### Error Testing
```go
// Go error testing
func TestErrorConditions(t *testing.T) {
    expectedErr := errors.New("API error")
    client := NewMockAIClientWithError(expectedErr)

    _, err := client.GetCommitMessage(context.Background(), "test prompt")
    assert.Error(t, err)
    assert.Equal(t, expectedErr, err)
}
```

```typescript
// TypeScript error testing
it('应该正确处理 API 错误', async () => {
    mockGetProjectStatus.mockRejectedValue(new Error('Network error'))

    await store.refresh(testPath, { force: true })
    await vi.runAllTimersAsync()

    const status = store.getStatus(testPath)
    expect(status?.error).toContain('Network error')
})
```

### Test Setup/Teardown
```go
// Go setup/teardown
func TestSuite(t *testing.T) {
    // Setup
    t.Setenv("TEST_ENV", "true")

    // Cleanup
    t.Cleanup(func() {
        // Teardown code
    })
}
```

```typescript
// TypeScript setup/teardown
beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
})

afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
})
```

---

*Testing analysis: 2026-02-05*