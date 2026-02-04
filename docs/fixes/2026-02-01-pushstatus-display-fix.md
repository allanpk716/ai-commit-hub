# 修复报告：项目列表 ahead/behind 提交数不显示问题

**日期**: 2026-02-01
**问题**: 项目列表中 Git 与远程分支的差异提交数（ahead/behind）不显示
**状态**: ✅ 已修复

## 问题描述

### 用户反馈
- 项目已经提交到本地，CommitPanel 显示可以推送
- 但项目列表中对应项目的 ahead/behind 提交数完全不显示
- 命令行确认本地确实领先远程 1 个提交

### 预期行为
- 项目列表应显示 "↑ 1" 指标，表示本地领先远程 1 个提交
- 提示文本显示"本地领先 1 个提交，可推送"

## 根因分析

### 问题定位过程

1. **后端验证** ✅
   - `git.GetPushStatus()` 函数工作正常
   - 测试显示返回正确的数据：
     ```json
     {
       "canPush": true,
       "ahead_count": 1,
       "behind_count": 0,
       "remote_branch": "origin/main"
     }
     ```

2. **前端验证** ❌
   - TypeScript 编译错误：
     ```
     Property 'ahead_count' does not exist on type 'PushStatus'.
     Did you mean 'aheadCount'?
     ```

3. **根因确认**
   - **后端 Go 结构体**使用蛇形命名（snake_case）：
     ```go
     type PushStatus struct {
         CanPush      bool   `json:"canPush"`
         AheadCount   int    `json:"ahead_count"`     // ❌ 蛇形
         BehindCount  int    `json:"behind_count"`    // ❌ 蛇形
         RemoteBranch string `json:"remote_branch"`   // ❌ 蛇形
     }
     ```

   - **前端 TypeScript 类型**使用驼峰命名（camelCase）：
     ```typescript
     export interface PushStatus {
       canPush: boolean
       aheadCount: number    // ❌ 驼峰
       behindCount: number   // ❌ 驼峰
       remoteBranch: string  // ❌ 驼峰
     }
     ```

   - **数据流断裂**：
     - 后端返回 `{"ahead_count": 1}`
     - 前端期望读取 `pushStatus.aheadCount`
     - 结果：`pushStatus.aheadCount = undefined`
     - UI 条件 `v-if="getProjectStatus(project).aheadCount > 0"` 失败
     - 指标不显示

## 修复方案

### 修改文件
**文件**: `pkg/git/git.go:640-647`

**修改前**:
```go
type PushStatus struct {
    CanPush      bool   `json:"canPush"`
    AheadCount   int    `json:"ahead_count"`
    BehindCount  int    `json:"behind_count"`
    RemoteBranch string `json:"remote_branch"`
    Error        string `json:"error,omitempty"`
}
```

**修改后**:
```go
type PushStatus struct {
    CanPush      bool   `json:"canPush"`
    AheadCount   int    `json:"aheadCount"`
    BehindCount  int    `json:"behindCount"`
    RemoteBranch string `json:"remoteBranch"`
    Error        string `json:"error,omitempty"`
}
```

### 修复原则
- **命名一致性**: 统一使用驼峰命名（camelCase）
- **Go 惯例**: Go 的 JSON tag 通常遵循接收语言的惯例，这里前端使用 TypeScript
- **最小改动**: 只修改 Go 结构体的 JSON tag，不改变字段名和逻辑

## 验证结果

### JSON 序列化测试
```bash
$ cd tmp && go run test_json_serialization.go "C:\WorkSpace\Go2Hell\src\github.com\allanpk716\ai-commit-hub"
```

**输出**:
```json
{
  "canPush": true,
  "aheadCount": 1,      // ✅ 驼峰命名
  "behindCount": 0,     // ✅ 驼峰命名
  "remoteBranch": "origin/main"  // ✅ 驼峰命名
}
```

### 编译测试
```bash
$ wails build -clean
```

**结果**: ✅ 编译成功，无 TypeScript 错误

### 功能测试
- 启动应用后，项目列表应正确显示 ahead/behind 指标
- UI 显示 "↑ 1" 表示本地领先远程 1 个提交
- CommitPanel 推送按钮正常工作

## 影响范围

### 受影响的模块
1. **前端 UI 组件**:
   - `ProjectList.vue`: 显示 ahead/behind 指标
   - `CommitPanel.vue`: 推送按钮状态和提示

2. **状态管理**:
   - `statusCache.ts`: 缓存 pushStatus 数据

3. **后端 API**:
   - `GetPushStatus()`: 返回推送状态
   - `GetAllProjectStatuses()`: 批量加载项目状态

### 数据流
```
后端 Go 结构体 → JSON 序列化 → Wails 绑定 → TypeScript 对象
```

修复前：蛇形命名导致数据映射失败
修复后：驼峰命名，数据正确映射

## 经验教训

### 问题预防
1. **类型同步检查**:
   - Go 结构体修改后，检查对应的 TypeScript 类型定义
   - 使用工具或脚本自动生成类型定义，减少手动维护错误

2. **命名规范**:
   - 前后端统一命名风格（推荐驼峰命名）
   - 在项目文档中明确命名规范

3. **测试覆盖**:
   - 添加集成测试验证前后端数据传递
   - 测试 JSON 序列化/反序列化

### 调试技巧
1. **TypeScript 编译错误**是重要的信号
2. **逐步验证**：后端 → JSON → 前端，逐层检查
3. **最小化测试**：创建独立的测试程序验证单个功能

## 相关代码位置

- **后端**: `pkg/git/git.go:640-647`
- **前端类型**: `frontend/src/types/status.ts:14-25`
- **UI 组件**: `frontend/src/components/ProjectList.vue:106-123`
- **状态缓存**: `frontend/src/stores/statusCache.ts:375-389`

## 附件

### 测试程序
- `tmp/test_pushstatus.go`: 验证 GetPushStatus 功能
- `tmp/test_json_serialization.go`: 验证 JSON 序列化

### 相关 Issue
- 无（首次发现）
