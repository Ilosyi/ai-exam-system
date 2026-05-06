# CLAUDE.md

## 1. 文档目的

本文件是当前项目的“开发基线说明书”，用于回答以下问题：

1. 这个项目现在是什么状态
2. 代码应该按什么方式继续扩展
3. 哪些约束必须保持一致
4. 新功能落地时应优先遵守什么规范

如果 `CLAUDE.md` 与旧实现、旧计划冲突，应以“当前代码真实状态 + 本文档约束”为准。

---

## 2. 项目定位

项目是一个面向教学场景的 AI 题库与考试系统，包含教师后台与学生考试端两套使用路径。

当前已形成的主链路：

- 教师登录后台
- 题库管理与 AI 出题
- 智能组卷
- 试卷发布到班级或公共范围
- 学生查看可参加考试
- 学生答题、自动保存、交卷、查看结果
- 教师按班级管理学生，并查看试卷提交情况

技术栈：

- 前端：React 18 + Vite + TypeScript + Ant Design
- 后端：Go 1.22 + Gin + GORM + SQLite
- AI：OpenAI 兼容协议，当前通过 `openai-go` 官方 SDK 对接 DashScope 兼容接口

---

## 3. 当前系统结构

### 3.1 前端结构

目录职责：

- `client/src/pages`
  页面级编排，负责布局、页面状态组合、接口调用调度
- `client/src/components`
  可复用展示组件与交互组件
- `client/src/api`
  所有 HTTP 请求封装与返回类型定义
- `client/src/hooks`
  负责页面数据加载、CRUD 行为、鉴权态封装
- `client/src/types`
  业务类型定义

当前主要页面：

- `/login`
  登录/注册页
- `/questions`
  题库管理
- `/notes`
  学习心得页
- `/papers/generate`
  智能组卷
- `/papers`
  试卷管理
- `/papers/:id/edit`
  试卷编辑
- `/classes`
  班级管理
- `/users`
  管理员用户管理
- `/exam`
  学生考试列表
- `/exam/:id/take`
  学生答题页
- `/exam/:id/result`
  学生成绩页

权限分层：

- `admin`
  可访问后台全部内容，包括用户管理
- `teacher`
  可访问题库、试卷、班级、AI 出题等教师端功能
- `student`
  仅访问考试端

### 3.2 后端结构

目录职责：

- `server/main.go`
  仅负责调用 `bootstrap.Start`
- `server/bootstrap`
  应用装配、依赖初始化、路由注册、静态资源挂载
- `server/handlers`
  参数校验、权限校验、响应组织
- `server/repositories`
  数据访问与查询逻辑
- `server/models`
  GORM 模型定义
- `server/services`
  AI、认证等领域服务
- `server/config`
  配置文件加载
- `server/database`
  SQLite 初始化

当前后端已完成路由分组：

- 公开接口
  - `/api/auth/login`
  - `/api/auth/register`
  - `/api/auth/refresh`
- 已登录接口
  - `/api/auth/me`
  - `/api/auth/logout`
  - `/api/auth/change-password`
- 管理员接口
  - `/api/users`
- 教师/管理员接口
  - `/api/questions`
  - `/api/ai/generate`
  - `/api/ai/test`
  - `/api/papers/*`
  - `/api/classes*`
- 学生/管理员接口
  - `/api/exam/*`

---

## 4. 当前核心业务模块

### 4.1 认证与鉴权

已完成：

- 用户登录、注册、刷新 token、登出、修改密码
- 基于 JWT 的认证中间件
- 路由级角色隔离
- 管理员用户管理

当前约束：

- 后端统一使用 Bearer Token
- 前端统一由 `useAuth` 管理登录态
- 401 视为登录失效，前端应回到登录流程

### 4.2 题库管理

已完成：

- 题目 CRUD
- 批量删除
- AI 出题
- 创建人字段展示

当前字段约束：

- `type`: `single | multiple | coding`
- `language`: `go | cpp | java | javascript | python`

### 4.3 AI 出题

当前实现：

- 后端使用 `openai-go` 官方 SDK
- `baseURL` 只配置到兼容 API 根路径，例如：
  `https://dashscope.aliyuncs.com/compatible-mode/v1`
- 出题请求末尾自动追加 `/no_think`
- 前端支持计时、等待提示、测试连接按钮
- 测试连接接口不带系统提示词，只发送 `who are you`
- 前端显示结构化诊断卡片：
  - 当前模型
  - Base URL
  - 请求耗时
  - 回复摘要
  - 原始回复

配置文件要求：

- `config.json`
  - `baseURL`
  - `model`
  - `requestTimeoutSeconds`
  - `systemPromptLines`
- `.env`
  - `DASHSCOPE_API_KEY`

### 4.4 试卷与考试

已完成：

- 智能组卷
- 草稿试卷保存
- 编辑试卷、替换题目、删除题目
- 发布/取消发布
- 学生考试列表
- 学生开始答题
- 自动保存
- 自动阅卷
- 查看结果
- 监考事件上报

当前业务规则：

- 发布支持 `classId`，不传表示公共试卷
- 学生可见试卷基于班级关系与发布时间窗过滤
- 截止时间按 `min(startedAt + duration, endTime)` 计算

### 4.5 班级体系

已完成：

- 班级 CRUD
- 教师只能管理自己班级
- 学生支持多班级关系（兼容 `users.class_id`）
- 班级成员管理抽屉
- 批量加入/移出
- 单学生考试记录
- 试卷提交情况统计

当前约束：

- 班级成员关系以 `user_classes` 为主，多对多
- `users.class_id` 仍保留兼容语义
- 教师不可越权查看其他教师班级

---

## 5. 数据模型总览

当前关键模型：

- `questions`
  题库题目，包含题型、语言、选项、答案、创建人
- `papers`
  试卷基本信息
- `paper_items`
  试卷题目项
- `paper_publications`
  试卷发布时间窗、发布班级、时长
- `exam_attempts`
  学生答题记录
- `exam_answers`
  学生题目答案
- `proctor_events`
  监考事件
- `users`
  系统用户
- `classes`
  班级
- `user_classes`
  学生-班级多对多关系

数据库：

- SQLite 文件路径：`server/data/questions.db`
- 迁移方式：GORM `AutoMigrate`

---

## 6. API 设计约定

### 6.1 响应格式

统一约定：

- 列表：
  `{"data": [...], "total": n}`
- 单对象：
  `{"data": {...}}`
- 动作型成功：
  `{"message": "..."}`
- 错误：
  `{"message": "...", "error": "..."}`

### 6.2 时间格式

- 后端统一输出 RFC3339
- 前端负责格式化展示

### 6.3 权限原则

- 优先在后端做权限控制
- 前端菜单可见性只作为体验优化，不作为安全边界
- 涉及班级、试卷、答题记录的接口必须做资源归属校验

---

## 7. 开发规范

### 7.1 前端规范

- 优先维护 `*.ts` / `*.tsx`
- 同名 `*.js` 文件视为镜像产物，不作为主维护面
- `api` 层不写页面逻辑
- `pages` 层不直接写大段底层请求细节
- 复杂副作用优先抽到 `hooks`
- 所有新类型优先补到 `types` 或 API 类型声明中

### 7.2 后端规范

- 参数校验写在 `handler`
- 查询与事务逻辑写在 `repository`
- 纯业务能力放在 `service`
- `bootstrap` 负责装配，不承载业务逻辑
- 新接口必须先考虑角色与资源归属校验

### 7.3 UI 规范

- 延续现有 Ant Design 体系
- 教师端页面允许做“教务面板化”的增强设计，但不能破坏表单与表格的一致性
- 关键动作必须有明确反馈：
  - loading
  - success
  - error
  - empty state

### 7.4 配置规范

- OpenAI 兼容 `baseURL` 只写到 `/v1`
- 不能把 `/chat/completions` 这种具体路径硬编码到配置中
- 超时必须走配置，不应在代码里写死魔法值

---

## 8. 当前技术债务

1. `client/src` 下仍保留大量 TS 同名 JS 镜像文件，存在维护噪音
2. 前端产物 `client/dist` 当前仍在仓库变更面中频繁更新
3. 前端包体偏大，构建存在 chunk size 警告
4. 自动化测试不足，当前仍以手工联调为主
5. 班级成员管理的“候选学生池”逻辑已可用，但后续仍可继续拆成独立 hook / 子组件
6. AI 兼容层刚切到 `openai-go`，后续仍应补充分供应商兼容性验证

---

## 9. 回归验证基线

每次涉及主链路改动，至少验证以下内容：

### 9.1 后端

```bash
cd server
go run .
```

或：

```bash
cd server
go build ./...
```

### 9.2 前端

```bash
cd client
pnpm build
```

开发模式：

```bash
cd client
pnpm dev
```

### 9.3 冒烟场景

至少覆盖：

1. 登录成功，角色跳转正确
2. 题库 CRUD 正常
3. AI 出题与测试连接正常
4. 智能组卷与试卷发布正常
5. 班级成员管理批量加入/移出正常
6. 试卷提交情况统计正常
7. 学生答题主链路正常

---

## 10. 下一阶段优先级

当前推荐优先级：

### P1：编程题闭环

- 代码编辑器
- 提交评测
- 评测结果展示

### P2：监考增强

- tab 切换
- heartbeat
- 教师监考看板

### P3：成绩分析与导出

- 班级维度成绩统计
- CSV 导出

### P4：工程质量治理

- E2E 冒烟脚本
- 关键 hooks 单测
- 构建与测试纳入 CI

---

## 11. 开发时的判断原则

后续开发优先遵守以下顺序：

1. 先保证权限与数据归属正确
2. 再保证接口与页面主链路闭环
3. 再做体验和视觉细节
4. 最后处理抽象、拆分与性能优化

如果需要新增能力，默认延续当前架构，不要绕开：

- 后端：`bootstrap -> handlers -> repositories/services`
- 前端：`api -> hooks -> pages/components`

这份文档的目标不是描述“曾经计划做什么”，而是明确“现在项目应该如何继续往前开发”。
