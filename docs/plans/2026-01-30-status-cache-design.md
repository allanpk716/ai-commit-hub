# Status Cache Design

**Date**: 2026-01-30
**Author**: Claude Code
**Status**: Design

## Problem Statement

Current project switching UX suffers from:
1. **Double refresh**: `CommitPanel` and `ProjectList` both trigger status loads
2. **Status flickering**: UI shows "untracked/not installed" then correct state
3. **Perceptible delay**: Users wait 300-500ms on each project switch
4. **Redundant API calls**: Same project status fetched multiple times

## Solution Overview

Implement a **StatusCache** layer that:
- Preloads all project statuses on startup
- Returns cached data immediately on switch
- Silently refreshes in background
- Optimistically updates after user actions
- Handles failures gracefully

## Architecture

### Component Structure

```
┌─────────────────┐
│   UI Components │
│  (ProjectList,  │
│   CommitPanel)  │
└────────┬────────┘
         │
┌────────▼────────┐
│   StatusCache   │ ← Status caching layer
│  (cache + API)  │
└────────┬────────┘
         │
┌────────▼────────┐
│  Backend APIs   │
│  (Go Services)  │
└─────────────────┘
```

### Cache Data Structure

```typescript
interface ProjectStatusCache {
  [projectPath: string]: {
    gitStatus: GitStatus | null;
    stagingStatus: StagingStatus | null;
    untrackedCount: number;
    pushoverStatus: PushoverHookStatus | null;
    lastUpdated: number;
    loading: boolean;
    error: string | null;
    stale: boolean;
  }
}
```

## Frontend Implementation

### New Store: StatusCache

Location: `frontend/src/stores/statusCache.ts`

**Key Methods**:

- `init()` - Initialize and preload all projects
- `getStatus(path)` - Get cached status for a project
- `refresh(path)` - Refresh a single project (with deduplication)
- `preload(paths)` - Batch preload multiple projects
- `updateOptimistic(path, data)` - Update cache optimistically
- `isExpired(path)` - Check if cache entry expired

**Cache TTL**: 30 seconds

### ProjectList.vue Changes

- Use `StatusCache.getStatus()` instead of direct store calls
- Show skeleton on first load instead of "untracked"
- Display stale indicator for expired cache

### CommitPanel.vue Changes

```typescript
watch(selectedProject, async (newProject) => {
  if (!newProject) return;

  // 1. Show cache immediately
  const cached = statusCache.getStatus(newProject.path);
  if (cached) {
    updateUIFromCache(cached);
  } else {
    showSkeleton();
  }

  // 2. Background refresh
  await statusCache.refresh(newProject.path);

  // 3. Update UI with fresh data
  const fresh = statusCache.getStatus(newProject.path);
  updateUIFromCache(fresh);
});
```

## Backend Implementation

### New API: GetAllProjectStatuses

Location: `app.go`

```go
func (a *App) GetAllProjectStatuses(projectPaths []string) (map[string]*ProjectFullStatus, error)
```

**Features**:
- Parallel querying with concurrency limit (10 goroutines)
- Returns map of path -> full status
- Individual failures don't block other projects
- Includes Git status, staging, untracked, Pushover

**New Type**:

```go
type ProjectFullStatus struct {
    GitStatus      *GitStatus
    StagingStatus  *StagingStatus
    UntrackedCount int
    PushoverStatus *PushoverHookStatus
    LastUpdated    time.Time
}
```

## Data Flows

### Startup Flow

```
App Start
  → statusCache.init()
  → Backend: GetAllProjects()
  → Backend: GetAllProjectStatuses(paths) [parallel]
  → Cache populated
  → UI displays immediately
```

### Project Switch Flow

```
User Click
  → CommitPanel watch triggers
  → 1. Display cached data (instant)
  → 2. Background refresh (non-blocking)
  → 3. UI updates when ready
```

### Post-Commit Flow

```
Commit Success
  → 1. Optimistic cache update
  → 2. UI reflects change immediately
  → 3. Background verification
  → 4. Rollback on failure
```

## Error Handling

### Cache Miss

- Show skeleton screen
- Display error on failure
- Provide retry button

### Network Error

- Use stale cache if < 5 minutes old
- Mark as `stale: true`
- Show "may be outdated" indicator

### Partial Failure

- Use successful API results
- Show warning for failed parts
- Example: Git status OK, Pushover failed

### Optimistic Update Failure

- Rollback to previous cache state
- Display error message
- Keep UI in consistent state

## Performance Targets

| Metric | Current | Target |
|--------|---------|--------|
| First load | 2-3s | ~1s |
| Project switch | 300-500ms | <100ms |
| Status flicker | 100% | 0% |
| API calls per switch | 4 | 0-1 |

## Testing Strategy

### Unit Tests

- Cache get/set operations
- TTL expiration logic
- Request deduplication
- Optimistic updates

### Integration Tests

- Startup preload flow
- Project switch timing
- No flicker on switch
- Error recovery

### Manual Checklist

- [ ] All projects load simultaneously on startup
- [ ] Switch shows cached state instantly
- [ ] No "untracked→tracked" flicker
- [ ] Errors handled gracefully
- [ ] Optimistic updates work correctly

## Implementation Phases

1. **Phase 1**: Create StatusCache store with basic caching
2. **Phase 2**: Implement startup preloading
3. **Phase 3**: Update ProjectList to use cache
4. **Phase 4**: Update CommitPanel to use cache
5. **Phase 5**: Add backend batch API
6. **Phase 6**: Implement optimistic updates
7. **Phase 7**: Add error handling and recovery

## Migration Notes

- Existing `projectStore` and `commitStore` remain
- StatusCache sits beneath them as optimization layer
- No breaking changes to existing APIs
- Can roll back by removing StatusCache usage
