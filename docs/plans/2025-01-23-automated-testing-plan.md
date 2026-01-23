# AI Commit Hub 自动化测试计划

## 文档信息

- **创建日期**: 2025-01-23
- **版本**: 1.0
- **状态**: 设计阶段

## 1. 概述

本文档为 AI Commit Hub 项目制定完整的自动化测试计划，覆盖从单元测试到端到端测试的所有层级，包含前端点击测试和后端日志分析。

### 1.1 测试目标

- 验证核心功能正确性
- 确保前后端通信可靠
- 模拟真实用户操作场景
- 捕获和分析运行时日志
- 实现持续集成自动化

### 1.2 测试范围

- **后端**: Go 代码 (Repository, Service, Git 操作, AI 集成)
- **前端**: Vue3 组件和 Pinia Store
- **集成**: Wails API 绑定和事件流
- **E2E**: 完整用户工作流
- **日志**: 结构化日志捕获和分析

## 2. 测试金字塔架构

```
                    E2E Tests
                   (少量)
                  ┌─────────┐
                  │  Wails  │
                  │   E2E   │
                 ─┴─────────┴─
                Integration Tests
              (中等数量)
             ┌──────────────────┐
             │  API Binding     │
             │  Wails Events    │
             │  Service Layer   │
            ─┴──────────────────┴─
           Unit Tests
         (大量)
       ┌──────────────────────────────┐
       │ Repository (GORM/SQLite)    │
       │ Service (AI Provider, Git)   │
       │ Vue Components (Vitest)      │
       │ Pinia Stores                 │
      ─┴──────────────────────────────┴─
```

## 3. 测试目录结构

```
ai-commit-hub/
├── tests/                          # 新增测试目录
│   ├── e2e/                        # E2E 测试
│   │   ├── app_test.go            # Wails 应用 E2E
│   │   ├── fixtures/              # 测试固件
│   │   │   ├── test-projects/     # 测试用 Git 仓库
│   │   │   └── config/            # 测试配置
│   │   └── helpers/               # E2E 辅助函数
│   ├── integration/               # 集成测试
│   │   ├── api_test.go           # Wails API 测试
│   │   └── events_test.go        # Wails Events 测试
│   └── setup/                     # 测试设置
│       └── setup.go               # 测试环境初始化
├── pkg/
│   ├── repository/                # 现有
│   │   ├── git_project_repository_test.go  # 已存在
│   │   └── commit_history_repository_test.go
│   ├── service/
│   │   └── service_test.go       # 新增 Service 测试
│   ├── git/
│   │   └── git_test.go           # 新增 Git 操作测试
│   └── ai/
│       └── ai_test.go            # 新增 AI Client 测试
└── frontend/
    └── tests/                     # 前端测试
        ├── unit/                  # 单元测试
        │   ├── components/        # Vue 组件测试
        │   └── stores/            # Pinia Store 测试
        └── integration/           # 前端集成测试
```

## 4. 单元测试层

### 4.1 Repository 层测试

**目标**: 验证数据访问逻辑正确性

**测试用例**:

```go
// GitProjectRepository 测试
func TestGitProjectRepository_Create(t *testing.T)
func TestGitProjectRepository_GetAll(t *testing.T)
func TestGitProjectRepository_Delete(t *testing.T)
func TestGitProjectRepository_Update(t *testing.T)
func TestGitProjectRepository_GetMaxSortOrder(t *testing.T)

// CommitHistoryRepository 测试
func TestCommitHistoryRepository_Create(t *testing.T)
func TestCommitHistoryRepository_GetByProjectID(t *testing.T)
func TestCommitHistoryRepository_CascadeDelete(t *testing.T)
```

### 4.2 Service 层测试

**目标**: 验证业务逻辑，使用 Mock 隔离外部依赖

**测试用例**:

```go
// ConfigService 测试
func TestConfigService_LoadConfig(t *testing.T)
func TestConfigService_Validate(t *testing.T)
func TestConfigService_GetProvider(t *testing.T)

// CommitService 测试
func TestCommitService_GenerateCommit(t *testing.T)
func TestCommitService_GenerateCommit_APIError(t *testing.T)
func TestCommitService_GenerateCommit_EmptyDiff(t *testing.T)
func TestCommitService_GenerateCommit_StreamOutput(t *testing.T)
```

### 4.3 Git 操作层测试

**目标**: 验证 Git 命令封装

**测试策略**: 使用临时 Git 仓库作为测试固件

```go
func setupTestRepo(t *testing.T) string {
    // 创建临时目录
    // 初始化 Git 仓库
    // 创建测试文件和 commit
    // 返回仓库路径
}

func TestGetProjectStatus(t *testing.T)
func TestGetDiff(t *testing.T)
func TestCommitChanges(t *testing.T)
func TestGetCurrentBranch(t *testing.T)
```

### 4.4 AI Client 测试

**目标**: 验证各 Provider 实现

**测试方法**: 使用 httptest 模拟 API 响应

```go
func TestOpenAIClient_Generate(t *testing.T)
func TestAnthropicClient_Generate(t *testing.T)
func TestDeepSeekClient_Generate(t *testing.T)
func TestStreamingClient_GenerateStream(t *testing.T)
func TestClient_APIError(t *testing.T)
func TestClient_Timeout(t *testing.T)
```

### 4.5 前端单元测试

**目标**: 验证 Vue 组件和 Store

**测试工具**: Vitest + Vue Test Utils

**测试文件**: `frontend/tests/unit/components/ProjectList.spec.ts`

```typescript
describe('ProjectList.vue', () => {
  it('renders project list correctly')
  it('filters projects by search query')
  it('emits select event when project clicked')
  it('disables delete button during loading')
  it('handles drag and drop reordering')
  it('shows empty state when no projects')
})
```

**CommitPanel 组件测试**:

```typescript
describe('CommitPanel.vue', () => {
  it('displays project status correctly')
  it('shows AI settings controls')
  it('generates commit message on button click')
  it('displays streaming output')
  it('handles copy to clipboard')
  it('commits locally with generated message')
  it('loads and displays history')
})
```

**Store 测试**:

```typescript
describe('projectStore', () => {
  it('loads projects on initialization')
  it('adds new project')
  it('deletes project')
  it('moves project up/down')
  it('reorders projects')
})

describe('commitStore', () => {
  it('generates commit message')
  it('handles streaming updates')
  it('clears error on new attempt')
  it('loads project status')
  it('saves commit history')
})
```

## 5. 集成测试层

### 5.1 Wails API 绑定测试

**目标**: 验证前后端通信正确性

**测试文件**: `tests/integration/api_test.go`

```go
func TestAppAPI_AddProject(t *testing.T)
func TestAppAPI_DeleteProject(t *testing.T)
func TestAppAPI_MoveProject(t *testing.T)
func TestAppAPI_ReorderProjects(t *testing.T)
func TestAppAPI_GetProjectStatus(t *testing.T)
func TestAppAPI_CommitLocally(t *testing.T)
func TestAppAPI_SaveCommitHistory(t *testing.T)
func TestAppAPI_GetProjectHistory(t *testing.T)
```

### 5.2 Wails Events 测试

**目标**: 验证流式输出事件机制

**测试文件**: `tests/integration/events_test.go`

```go
func TestCommitGeneration_Events(t *testing.T)
func TestEvent_DeltaStream(t *testing.T)
func TestEvent_Complete(t *testing.T)
func TestEvent_ErrorPropagation(t *testing.T)
```

### 5.3 数据库集成测试

**目标**: 验证跨 Repository 的事务和数据一致性

```go
func TestDatabase_Migration(t *testing.T)
func TestDatabase_CascadeDelete(t *testing.T)
func TestDatabase_ConcurrentAccess(t *testing.T)
func TestDatabase_TransactionRollback(t *testing.T)
```

### 5.4 配置文件集成测试

```go
func TestConfigService_ProviderConfiguration(t *testing.T)
func TestConfigService_CustomPromptTemplate(t *testing.T)
func TestConfigService_LanguageSettings(t *testing.T)
func TestConfigService_InvalidConfiguration(t *testing.T)
```

### 5.5 Git 操作集成测试

```go
func TestGitIntegration_CompleteWorkflow(t *testing.T)
func TestGitIntegration_StatusChanges(t *testing.T)
func TestGitIntegration_DiffGeneration(t *testing.T)
func TestGitIntegration_CommitExecution(t *testing.T)
```

## 6. E2E 测试层

### 6.1 完整工作流测试

**测试文件**: `tests/e2e/app_test.go`

```go
func TestE2E_CompleteWorkflow(t *testing.T) {
    // 1. 启动 Wails 测试应用
    // 2. 准备测试 Git 仓库
    // 3. 添加项目
    // 4. 选择项目
    // 5. 生成 commit 消息
    // 6. 提交到本地
    // 7. 验证历史记录
}
```

### 6.2 场景化测试用例

```go
// 场景 1: 新用户首次使用
func TestE2E_Scenario_NewUser(t *testing.T)

// 场景 2: 多项目管理
func TestE2E_Scenario_MultipleProjects(t *testing.T)

// 场景 3: 错误处理
func TestE2E_Scenario_ErrorHandling(t *testing.T)

// 场景 4: AI Provider 切换
func TestE2E_Scenario_ProviderSwitching(t *testing.T)

// 场景 5: 搜索和过滤
func TestE2E_Scenario_SearchAndFilter(t *testing.T)

// 场景 6: 项目排序
func TestE2E_Scenario_ProjectReordering(t *testing.T)

// 场景 7: 历史记录管理
func TestE2E_Scenario_HistoryManagement(t *testing.T)

// 场景 8: 配置文件夹访问
func TestE2E_Scenario_ConfigFolderAccess(t *testing.T)
```

### 6.3 测试固件和辅助函数

**测试文件**: `tests/e2e/helpers/fixtures.go`

```go
// setupE2EApp 创建测试用 App 实例
func setupE2EApp(t *testing.T) *App

// teardownE2EApp 清理测试环境
func teardownE2EApp(t *testing.T, app *App)

// prepareTestRepo 创建测试 Git 仓库
func prepareTestRepo(t *testing.T) *TestRepo

// prepareTestRepoNamed 创建指定名称的测试仓库
func prepareTestRepoNamed(t *testing.T, name string) *TestRepo

// collectEvents 收集 Wails Events
func collectEvents(t *testing.T, app *App, eventName string) <-chan []string

// waitForEvent 等待特定事件
func waitForEvent(t *testing.T, app *App, eventName string, timeout time.Duration) <-chan bool
```

### 6.4 前端 E2E 测试

**测试文件**: `frontend/tests/e2e/commit.spec.ts`

```typescript
test.describe('Commit Panel', () => {
  test('should generate commit message')
  test('should handle errors gracefully')
  test('should display streaming output')
  test('should commit locally')
  test('should save and load history')
  test('should switch between providers')
})
```

## 7. 日志分析

### 7.1 日志捕获框架

**测试文件**: `tests/helpers/log_capture.go`

```go
// LogCapture 捕获日志输出
type LogCapture struct {
    buffer *bytes.Buffer
    logger *logger.Logger
}

// NewLogCapture 创建日志捕获器
func NewLogCapture(t *testing.T) *LogCapture

// GetLogs 获取所有日志
func (lc *LogCapture) GetLogs() string

// GetLogsByLevel 按级别获取日志
func (lc *LogCapture) GetLogsByLevel(level string) []string

// Contains 验证日志包含特定内容
func (lc *LogCapture) Contains(substring string) bool

// ContainsError 验证有错误日志
func (lc *LogCapture) ContainsError() bool
```

### 7.2 日志断言

**测试文件**: `tests/helpers/log_assertions.go`

```go
// LogAssertions 日志断言
type LogAssertions struct {
    t      *testing.T
    capture *LogCapture
}

// AssertLogContains 断言日志包含文本
func (a *LogAssertions) AssertLogContains(substring string) bool

// AssertLogNotContains 断言日志不包含文本
func (a *LogAssertions) AssertLogNotContains(substring string) bool

// AssertLogCount 断言日志出现次数
func (a *LogAssertions) AssertLogCount(substring string, expectedCount int) bool

// AssertLogPattern 断言日志匹配正则表达式
func (a *LogAssertions) AssertLogPattern(pattern string) bool

// AssertNoErrors 断言没有错误或警告日志
func (a *LogAssertions) AssertNoErrors() bool

// AssertLogSequence 断言日志按顺序出现
func (a *LogAssertions) AssertLogSequence(substrings ...string) bool
```

### 7.3 日志分析测试用例

```go
func TestLogAnalysis_CommitGeneration(t *testing.T)
func TestLogAnalysis_ErrorHandling(t *testing.T)
func TestLogAnalysis_PerformanceMetrics(t *testing.T)
func TestLogAnalysis_EventFlow(t *testing.T)
```

### 7.4 结构化日志分析

**测试文件**: `tests/helpers/log_analyzer.go`

```go
// LogEntry 结构化日志条目
type LogEntry struct {
    Level     string
    Message   string
    Timestamp string
    Fields    map[string]interface{}
}

// LogAnalyzer 日志分析器
type LogAnalyzer struct {
    entries []LogEntry
}

// NewLogAnalyzer 从 JSON 格式日志创建分析器
func NewLogAnalyzer(jsonLogs string) (*LogAnalyzer, error)

// GetErrors 获取所有错误日志
func (la *LogAnalyzer) GetErrors() []LogEntry

// GetByField 按字段过滤
func (la *LogAnalyzer) GetByField(key, value string) []LogEntry

// AnalyzePerformance 分析性能指标
func (la *LogAnalyzer) AnalyzePerformance() PerformanceReport
```

### 7.5 测试报告生成

**测试文件**: `tests/helpers/report_generator.go`

```go
// TestReport 测试报告
type TestReport struct {
    TestName    string
    Duration    time.Duration
    Passed      bool
    Logs        string
    ErrorCount  int
    Performance PerformanceReport
}

// ReportGenerator 报告生成器
type ReportGenerator struct {
    reports []TestReport
    output  string
}

// Generate 生成 JSON 报告
func (rg *ReportGenerator) Generate() error

// GenerateHTML 生成 HTML 报告
func (rg *ReportGenerator) GenerateHTML() error
```

## 8. CI/CD 集成

### 8.1 GitHub Actions 工作流

**配置文件**: `.github/workflows/test.yml`

```yaml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  backend-test:
    name: Backend Tests
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [windows-latest, ubuntu-latest, macos-latest]

  frontend-test:
    name: Frontend Tests
    runs-on: ubuntu-latest

  e2e-test:
    name: E2E Tests
    runs-on: ${{ matrix.os }}
    needs: [backend-test, frontend-test]

  build-check:
    name: Build Check
    runs-on: ${{ matrix.os }}
```

### 8.2 本地测试脚本

**Windows 批处理脚本**: `scripts/run-tests.bat`

```batch
@echo off
setlocal enabledelayedexpansion

:: Parse arguments
set RUN_UNIT=1
set RUN_INTEGRATION=1
set RUN_E2E=0
set RUN_FRONTEND=1
set COVERAGE=0

:: Run tests based on flags
:: Generate coverage reports
:: Display summary
```

### 8.3 Makefile 测试目标

```makefile
.PHONY: test test-unit test-integration test-e2e test-frontend test-all coverage clean

test-all:
	@scripts/run-tests.bat

test-unit:
	@go test ./pkg/... -v -cover

test-integration:
	@go test ./tests/integration/... -v

test-e2e:
	@go test ./tests/e2e/... -v -timeout 30m

test-frontend:
	@cd frontend && npm test -- --run --coverage

coverage:
	@go tool covmerge tmp/coverage-*.out > tmp/coverage.out
	@go tool cover -html=tmp/coverage.out -o tmp/coverage.html
```

### 8.4 手动测试触发命令

```bash
# 快速验证 (开发中频繁使用)
make test-unit

# 完整测试套件 (提交前)
make test-all

# 仅测试变更的模块
go test ./$(git diff --name-only | grep pkg | sed 's|/[^/]*$||' | sort -u | tr '\n' ' ')

# 特定场景测试
go test ./tests/e2e/... -run TestE2E_Scenario_NewUser -v

# 带日志分析的测试
go test ./tests/integration/... -v -with-logs

# 性能测试
go test ./... -bench=. -benchmem
```

## 9. 实施计划

### 9.1 分阶段实施

```
阶段 1: 基础单元测试 (1-2 周)
├── Repository 层测试
├── Git 操作层测试
├── AI Client Mock 测试
└── 前端 Pinia Store 测试

阶段 2: 集成测试 (1-2 周)
├── Wails API 绑定测试
├── Wails Events 流式测试
├── 数据库集成测试
└── 配置管理测试

阶段 3: E2E 测试 (2-3 周)
├── 测试固件开发
├── 场景化测试用例
├── 日志分析集成
└── 前端 E2E 测试

阶段 4: CI/CD 和报告 (1 周)
├── GitHub Actions 配置
├── 测试报告生成
├── 覆盖率监控
└── 本地测试脚本优化
```

### 9.2 测试工具清单

| 层级 | 工具 | 用途 |
|------|------|------|
| Go 后端 | testing, testify/assert | 单元测试断言 |
| Go Mock | gomock, mockery | Mock 外部依赖 |
| 数据库 | SQLite 内存模式 | 隔离数据库测试 |
| 前端单元 | Vitest, Vue Test Utils | Vue 组件测试 |
| E2E | Wails 测试框架 | 端到端测试 |
| 日志 | 自定义 LogCapture | 日志捕获和分析 |
| 覆盖率 | go tool cover, codecov | 覆盖率报告 |
| CI/CD | GitHub Actions | 自动化测试流水线 |

### 9.3 测试用例统计

```
┌─────────────────┬──────────┬──────────┐
│   测试层级      │ 用例数   │ 预估时间 │
├─────────────────┼──────────┼──────────┤
│ Repository      │    15    │   1周    │
│ Service         │    20    │   1周    │
│ Git 操作        │    10    │   3天    │
│ AI Client       │    12    │   4天    │
│ 前端组件        │    25    │   1周    │
│ 前端 Store      │    15    │   3天    │
│ 集成测试        │    18    │   1周    │
│ E2E 测试        │    12    │   2周    │
├─────────────────┼──────────┼──────────┤
│   总计          │   127    │  8-10周  │
└─────────────────┴──────────┴──────────┘
```

## 10. 关键成功指标

### 10.1 代码覆盖率目标

- Repository 层: > 85%
- Service 层: > 80%
- Git 操作层: > 75%
- 前端组件: > 70%
- **整体目标: > 75%**

### 10.2 测试执行时间目标

- 单元测试: < 2 分钟
- 集成测试: < 5 分钟
- E2E 测试: < 15 分钟
- **完整套件: < 25 分钟**

### 10.3 Bug 发现率目标

- 单元测试发现: > 60%
- 集成测试发现: > 25%
- E2E 测试发现: > 10%
- **生产环境: < 5%**

## 11. 风险和注意事项

### 11.1 技术风险

| 风险 | 缓解措施 |
|------|----------|
| Wails E2E 测试支持不完善 | 优先实现集成测试，E2E 采用混合方案 |
| Git 操作依赖外部 git 命令 | 使用临时仓库和 mock Git 环境 |
| AI Provider API 调用成本 | 全面使用 Mock，真实 API 仅用于烟雾测试 |
| Windows/Linux/macOS 差异 | CI 覆盖三个平台，测试固件跨平台兼容 |
| 数据库并发问题 | 测试中使用事务回滚，确保隔离 |

### 11.2 实施建议

1. **YAGNI 原则**: 不测试显而易见的功能（如 getter/setter）
2. **测试隔离**: 每个测试独立运行，不依赖执行顺序
3. **快速反馈**: 单元测试应该快速运行，慢的测试标记为 `integration` 或 `e2e`
4. **可维护性**: 测试代码质量等同生产代码
5. **真实场景**: E2E 测试覆盖真实用户使用路径

## 12. 下一步行动

测试计划设计已完成，可以开始实施：

1. **搭建测试框架** - 创建目录结构和辅助函数
2. **编写单元测试** - 从 Repository 层开始
3. **实施集成测试** - Wails API 绑定测试
4. **开发 E2E 测试** - 场景化测试用例
5. **配置 CI/CD** - GitHub Actions 工作流
6. **监控覆盖率** - 生成测试报告

---

**文档结束**
