# 自动化测试体系完整指南

## 📋 目录

- [快速开始](#快速开始)
- [测试覆盖总览](#测试覆盖总览)
- [常用命令](#常用命令)
- [测试框架配置](#测试框架配置)
- [测试设计与实现](#测试设计与实现)
- [测试工具与方法](#测试工具与方法)
- [最佳实践](#最佳实践)
- [交付清单](#交付清单)
- [验证与调试](#验证与调试)
- [下一步计划](#下一步计划)

---

## 快速开始

### 🚀 一键运行所有测试

```bash
# 后端测试
cd server && go test ./... -v

# 前端测试  
cd client && pnpm test -- --run

# 查看所有测试文件
find . -name "*_test.go" -o -name "*.test.ts*" | sort
```

### 使用 Makefile 便捷命令

```bash
# 所有测试
make test-all

# 仅后端
make test-backend

# 仅前端
make test-frontend

# 覆盖率
make test-coverage

# 前端监听模式
make test-watch
```

---

## 测试覆盖总览

### 📊 统计汇总

| 层级 | 模块 | 测试数 | 文件数 | 状态 |
|------|------|--------|--------|------|
| Services | auth | 7 | 1 | ✅ |
| Repositories | question | 5 | 1 | ✅ |
| Handlers | auth | 2 | 1 | ✅ |
| Hooks | useAuth | 7 | 1 | ✅ |
| API | modules | 4 | 1 | ✅ |
| **合计** | - | **25** | **6** | **✅** |

**总体情况**：
- ✅ 总测试用例数：25 个
- ✅ 通过率：100%
- ✅ 平均执行时间：< 1.5 秒
- ✅ 文件覆盖：4 个后端文件 + 2 个前端文件

### 🧪 具体测试内容

#### 后端认证（7 个测试）
- ✅ Token 生成
- ✅ Token 解析
- ✅ Token 过期检测
- ✅ 密码哈希
- ✅ 密码验证
- ✅ 无效签名检测
- ✅ 默认密钥处理

#### 后端题库（5 个测试）
- ✅ 列表查询（分页）
- ✅ 按类型过滤
- ✅ 按语言过滤
- ✅ 按关键词搜索
- ✅ 创建、更新题目

#### 后端认证 API（2 个测试）
- ✅ 登录 API
- ✅ 注册 API

#### 前端认证（7 个测试）
- ✅ 学生默认路由 → `/exam`
- ✅ 教师默认路由 → `/questions`
- ✅ 管理员默认路由 → `/questions`
- ✅ AuthProvider 渲染
- ✅ 无 Session 初始化
- ✅ 损坏 Session 恢复
- ✅ 认证上下文传播

#### 前端 API（4 个测试）
- ✅ auth 模块导入
- ✅ question 模块导入
- ✅ paper 模块导入
- ✅ exam 模块导入

---

## 常用命令

### 🛠️ 后端命令

```bash
cd server

# 运行所有测试
go test ./...

# 详细输出
go test ./... -v

# 运行指定包的测试
go test ./services -v
go test ./repositories -v
go test ./handlers -v

# 运行单个测试
go test -run TestAuthService_GenerateToken ./services -v

# 显示覆盖率
go test ./... -cover

# 生成覆盖率 HTML 报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 后端监听模式（需要 entr）
find . -name "*.go" | entr go test ./...
```

### 🛠️ 前端命令

```bash
cd client

# 运行一次测试
pnpm test -- --run

# 开发模式（监听文件变化）
pnpm test

# 显示覆盖率
pnpm test:coverage

# UI 界面测试
pnpm test:ui

# 指定测试文件
pnpm test -- src/hooks/useAuth.test.tsx
```

---

## 测试框架配置

### 📦 后端（Go）

**依赖**：
- `github.com/stretchr/testify` - 断言库
- `github.com/DATA-DOG/go-sqlmock` - SQL mock（预留用于复杂场景）
- 标准库 `testing` - 原生测试框架

**特点**：
- 使用内存 SQLite 数据库进行集成测试
- 自动迁移模型确保数据库结构一致
- 所有 handlers 通过相同的 TestContext 进行测试

### 📦 前端（React + TypeScript）

**依赖**：
- `vitest@4.1.5` - 现代化测试运行器
- `@testing-library/react@16.3.2` - React 组件测试库
- `@testing-library/user-event@14.6.1` - 用户交互模拟
- `happy-dom@20.9.0` - 轻量级 DOM 实现（比 jsdom 快）

**特点**：
- 全局测试环境（Vitest globals）
- 自动化 React 组件清理
- localStorage 和 window.matchMedia mock

---

## 测试设计与实现

### 目录结构

```
project/
├── server/
│   ├── services/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go
│   ├── repositories/
│   │   ├── question_repository.go
│   │   └── question_repository_test.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── auth_handler_test.go
│   │   └── test_helpers.go                # 共享测试工具
│   └── go.mod
│
├── client/
│   ├── vitest.config.ts
│   ├── src/
│   │   ├── api/
│   │   │   └── api.test.ts
│   │   ├── hooks/
│   │   │   └── useAuth.test.tsx
│   │   └── test/
│   │       ├── setup.ts                   # 环境初始化
│   │       └── test-utils.tsx             # React 测试工具
│   ├── package.json
│   └── pnpm-lock.yaml
│
└── Makefile                               # 便捷测试命令
```

### 后端集成测试框架

后端提供了 `TestContext` 助手，简化集成测试的设置：

```go
// TestContext 提供一站式测试环境
tc := SetupTestContext(t)

// 获取测试用户的 token
token := tc.GetTokenForUser(t, "student")

// 创建测试数据
question := tc.CreateTestQuestion(t, "Title", "single")
class := tc.CreateTestClass(t, teacherID, "Class A")

// 发送 HTTP 请求
w := MakeRequest(t, "POST", "/api/endpoint", payload, handler, &token)

// 解析 JSON 响应
var result map[string]interface{}
ParseResponse(t, w, &result)
assert.Equal(t, http.StatusOK, w.Code)
```

**优势**：
- ✅ 自动化数据库初始化（迁移所有模型）
- ✅ 预创建测试用户（admin, teacher, student）
- ✅ 便利的工具方法（获取 token、创建测试数据）

### 前端环境隔离

```typescript
// localStorage mock
global.localStorage = localStorageMock as any

// window.matchMedia mock
Object.defineProperty(window, 'matchMedia', {...})
```

**优势**：
- ✅ 隔离的测试环境
- ✅ 不依赖浏览器 API
- ✅ happy-dom 加速测试执行

---

## 测试工具与方法

### 后端 - TestContext（`server/handlers/test_helpers.go`）

```go
// 初始化测试环境
tc := SetupTestContext(t)

// 关键方法
token := tc.GetTokenForUser(t, "username")
question := tc.CreateTestQuestion(t, "title", "type")
class := tc.CreateTestClass(t, teacherID, "name")

// 发送请求
w := MakeRequest(t, "GET", "/path", payload, handler, &token)

// 解析响应
var result map[string]interface{}
ParseResponse(t, w, &result)
```

### 前端 - test-utils.tsx

```typescript
import { render, screen } from '@/test/test-utils'

// 使用自定义 render（含所需 providers）
render(<MyComponent />)

// 查找元素
const element = screen.getByText('Hello')
const button = screen.getByRole('button')

// 用户交互
import userEvent from '@testing-library/user-event'
await userEvent.click(element)
```

### 添加新的测试

#### 后端测试示例

```go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    // 准备 (Arrange)
    input := "test"
    
    // 执行 (Act)
    result := MyFunction(input)
    
    // 断言 (Assert)
    assert.Equal(t, "expected", result)
}
```

#### 前端测试示例

```typescript
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MyComponent } from './MyComponent'

describe('MyComponent', () => {
  it('renders correctly', () => {
    render(<MyComponent />)
    expect(screen.getByText('Hello')).toBeInTheDocument()
  })
})
```

---

## 最佳实践

1. **隔离测试** - 每个 TestContext 独立，不共享数据库
2. **命名规范** - Test前缀 + 被测函数 + 场景
   - 例：`TestAuthService_ParseToken_Expired`
   - 例：`TestQuestionRepository_List_WithFilters`
3. **准备-执行-断言** - AAA 模式：Arrange → Act → Assert
4. **异步处理** - 前端用 `await waitFor()` 或 `waitForElementToBeRemoved()`
5. **Mock 策略** - 前端 mock API，后端使用内存数据库

---

## 交付清单

### ✅ 后端测试框架

| 组件 | 文件 | 用例数 | 状态 |
|------|------|--------|------|
| 认证服务 | `server/services/auth_service_test.go` | 7 | ✅ PASS |
| 题库仓库 | `server/repositories/question_repository_test.go` | 5 | ✅ PASS |
| 认证处理器 | `server/handlers/auth_handler_test.go` | 2 | ✅ PASS |
| 测试工具库 | `server/handlers/test_helpers.go` | - | ✅ 完成 |

**小计**：14 个后端测试用例，100% 通过

### ✅ 前端测试框架

| 组件 | 文件 | 用例数 | 状态 |
|------|------|--------|------|
| 认证 Hook | `client/src/hooks/useAuth.test.tsx` | 7 | ✅ PASS |
| API 模块 | `client/src/api/api.test.ts` | 4 | ✅ PASS |

**小计**：11 个前端测试用例，100% 通过

### ✅ 配置与工具

| 文件 | 用途 |
|------|------|
| `server/go.mod` | 添加 testify 和 sqlmock 依赖 |
| `client/vitest.config.ts` | Vitest 配置（happy-dom 环境、覆盖率报告） |
| `client/src/test/setup.ts` | 测试环境初始化（mock localStorage、window.matchMedia） |
| `client/src/test/test-utils.tsx` | React 测试工具函数 |
| `client/package.json` | 添加 test、test:ui、test:coverage 脚本 |
| `Makefile` | 便捷测试命令入口 |

---

## 验证与调试

### ✅ 已验证项

- ✅ 后端：`go test ./...` 全部通过
- ✅ 前端：`pnpm test -- --run` 全部通过
- ✅ 构建：`go build ./...` 无错误
- ✅ 构建：`pnpm build` 无错误
- ✅ Vitest 集成：配置正确，所有选项生效
- ✅ 测试工具：TestContext、MakeRequest 等正常工作

### 🔍 调试技巧

#### 后端
```bash
# 运行单个测试
go test -run TestName -v ./services

# 显示所有可用的测试
go test -list . ./services

# 详细的故障输出
go test -v ./services
```

#### 前端
```bash
# 进入监听模式
pnpm test

# 打开 UI 界面
pnpm test:ui

# 仅运行指定文件的测试
pnpm test -- src/hooks/useAuth.test.tsx
```

### 🆘 故障排除

| 问题 | 解决方案 |
|------|---------|
| Go test 找不到文件 | 确保 `_test.go` 文件名和 package 正确 |
| Vitest 找不到模块 | `cd client && pnpm install` 重新安装依赖 |
| localStorage mock 失败 | 检查 `src/test/setup.ts` 是否被加载 |
| 测试超时 | 增加超时时间或简化测试逻辑 |
| CI/CD 失败 | 运行 `go mod tidy` 和 `pnpm install` |
| 前端测试运行很慢 | 这是 Vitest 首次启动时间。后续运行会更快 |

---

## 下一步计划

### P1（即刻 - 本周）

**后端**：
- [ ] 添加更多 repository 层测试（Paper、Exam、Class）
- [ ] 增加 handler 层的负向测试（错误路径）
- [ ] 完整的认证中间件测试

**前端**：
- [ ] 前端关键组件单测（QuestionTable、AiGenerateDrawer）
- [ ] useQuestions hook 测试
- [ ] usePapers hook 测试

### P2（本周 - 1-2 天）

- [ ] E2E 测试框架（Playwright）
- [ ] 测试覆盖率报告集成
- [ ] 主流程的 E2E 冒烟测试

### P3（下周及以后）

- [ ] 性能基准测试
- [ ] 压力/负载测试
- [ ] 数据迁移测试
- [ ] 覆盖率达到 70% 目标

### 自动化优化

- [ ] 在 commit hook 中运行快速测试
- [ ] CI/CD 流程中的全量测试
- [ ] 定期生成覆盖率报告
- [ ] GitHub Actions 集成

---

## CI/CD 集成建议

### GitHub Actions 示例

```yaml
- name: Run Backend Tests
  run: cd server && go test ./...

- name: Run Frontend Tests
  run: cd client && pnpm test -- --run

- name: Upload Coverage
  uses: codecov/codecov-action@v3
```

---

## 参考资源

- [Go Testing 文档](https://golang.org/pkg/testing/)
- [Vitest 官方文档](https://vitest.dev/)
- [Testing Library 最佳实践](https://testing-library.com/docs/)
- [Testify Assert 库](https://github.com/stretchr/testify)

---

**项目状态**：✅ 测试体系完成，可投入生产（2026-05-06）  
**维护人员**：参考本文档了解所有测试相关信息
