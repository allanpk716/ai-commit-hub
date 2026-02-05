# Coding Conventions

**Analysis Date:** 2026-02-05

## Naming Patterns

### Files
- **Go files**: Lowercase with underscores (e.g., `commit_service.go`, `config_manager.go`)
- **Test files**: `_test.go` suffix (e.g., `commit_service_test.go`, `status_cache.spec.ts`)
- **TypeScript files**: PascalCase for component files, lowercase for utilities (e.g., `statusCache.ts`, `versionFormat.ts`)

### Functions
- **Go**: PascalCase for exported functions, lowercase for private ones
  ```go
  func GenerateCommit(...) // Exported
  func (s *CommitService) validateProvider(...) // Private
  ```
- **TypeScript**: camelCase for all functions
  ```typescript
  function getStatus(path: string)
  function refreshWithRetry(path: string, maxRetries = 2)
  ```

### Variables
- **Go**: PascalCase for exported, lowercase for private
  ```go
  var DefaultProvider = "phind"
  func (s *CommitService) configService *ConfigService
  ```
- **TypeScript**: camelCase with descriptive names
  ```typescript
  const cache = ref<ProjectStatusCacheMap>({})
  const pendingRequests = ref<Set<string>>(new Set())
  ```

### Types
- **Go**: PascalCase for all types
  ```go
  type CommitService struct {}
  type ProviderSettings struct {}
  type Config struct {}
  ```
- **TypeScript**: PascalCase for interfaces and types
  ```typescript
  interface ProjectStatusCache {}
  type CacheOptions = {}
  ```

## Code Style

### Go
- **Formatting**: Use `gofmt` for standard formatting
- **Line Length**: Aim for under 80 characters when possible
- **Error Handling**: Always use `fmt.Errorf` with `%w` wrapper
- **Logging**: Use `github.com/WQGroup/logger` with consistent patterns
  ```go
  logger.Info("开始生成 Commit 消息")
  logger.Infof("项目路径: %s", projectPath)
  logger.Errorf("加载配置失败: %v", err)
  ```

### TypeScript/Vue
- **Formatting**: Use Prettier with TypeScript configuration
- **Line Length**: Max 100 characters
- **Vue Composition API**: Use `<script setup>` with proper TypeScript
- **Store Pattern**: Follow Pinia conventions with `defineStore`

## Import Organization

### Go
```go
import (
    // Standard library (sorted)
    "context"
    "fmt"
    "os"

    // Third-party (sorted by import path)
    "github.com/WQGroup/logger"
    "github.com/allanpk716/ai-commit-hub/pkg/ai"
    "github.com/allanpk716/ai-commit-hub/pkg/git"

    // Internal packages (sorted by path)
    "github.com/allanpk716/ai-commit-hub/pkg/config"
    "github.com/allanpk716/ai-commit-hub/pkg/service"
)
```

### TypeScript
```typescript
// Standard library imports
import { ref, computed } from 'vue'

// Third-party imports
import { defineStore } from 'pinia'
import { vi } from 'vitest'

// Internal imports
import type { ProjectStatusCache } from '../types/status'
import { useStatusCache } from '../stores/statusCache'
```

## Error Handling

### Go Patterns
```go
// Standard error wrapping
return fmt.Errorf("创建 AI client 失败: %w", err)

// Error context with logging
logger.Errorf(errMsg)
runtime.EventsEmit(s.ctx, "commit-error", errMsg)
return fmt.Errorf("provider not configured: %s", cfg.Provider)
```

### TypeScript Patterns
```typescript
// Async error handling with try/catch
try {
    const response = await apiCall()
} catch (error) {
    console.error('Operation failed:', error)
    throw error
}

// Type-safe error handling
function isRetryable(error: unknown): boolean {
    if (error instanceof Error) {
        return error.message.includes('network') ||
               error.message.includes('timeout')
    }
    return false
}
```

## Logging

### Framework
- **Go**: `github.com/WQGroup/logger`
- **TypeScript**: Console methods with proper logging levels

### Patterns
```go
// Go logging with Chinese messages
logger.Info("开始生成 Commit 消息")
logger.Infof("项目路径: %s", projectPath)
logger.Warn("暂存区没有变更")
logger.Errorf("加载配置失败: %v", err)
```

```typescript
// TypeScript logging
console.debug('[StatusCache] 检测到不一致，已自动修正', data)
console.warn('[StatusCache] 刷新项目状态失败:', path, error)
console.error('Preload failed, falling back to individual loads:', error)
```

## Comments

### When to Comment
- Public APIs and their purpose
- Complex business logic decisions
- Important performance considerations
- Configuration options and their defaults

### Documentation Patterns
```go
// CommitService handles AI-powered commit message generation
type CommitService struct {
    ctx           context.Context
    configService *ConfigService
}

// GenerateCommit generates commit messages using AI providers
// Supports both streaming and non-streaming modes
func (s *CommitService) GenerateCommit(projectPath, providerName, language string) error
```

```typescript
/**
 * StatusCache Store - 项目状态缓存管理
 *
 * 用于缓存项目状态信息，避免频繁调用后端 API 导致 UI 闪烁
 * 支持 TTL 过期、乐观更新、批量预加载等特性
 */
export const useStatusCache = defineStore('statusCache', () => {
```

## Function Design

### Size Guidelines
- **Go**: Functions should be under 50 lines when possible
- **TypeScript**: Keep functions focused, single responsibility

### Parameters
- **Go**: Limit to 3-4 parameters, use structs for complex cases
- **TypeScript**: Use destructuring for multiple parameters

### Return Values
- **Go**: Return single value + error pattern
- **TypeScript**: Use Promises for async, direct returns for sync

## Module Design

### Exports
- **Go**: Export only what's necessary, minimize public API
- **TypeScript**: Use named exports for utilities, default for components

### Barrel Files
- **Go**: No barrel files pattern
- **TypeScript**: Use barrel files for type exports (`types/index.ts`)

---

*Convention analysis: 2026-02-05*