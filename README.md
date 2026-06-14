# Week05 · AI 题库与考试系统

一个面向教学场景的 AI 题库、组卷、发布、考试与成绩回看系统。

项目当前已经形成完整主链路：

- 教师登录后台
- 题库管理与 AI 出题
- 智能组卷
- 试卷发布到班级或公共范围
- 学生参加考试
- 自动保存、交卷、自动阅卷
- 教师查看班级成员、学生考试记录与试卷提交情况

---

## 项目概览

系统由前后端两部分组成：

- `client`
  React + Vite + TypeScript + Ant Design 教师端 / 学生端前端
- `server`
  Go + Gin + GORM + SQLite 后端服务

AI 能力通过 OpenAI 兼容协议接入，目前已使用 `openai-go` 官方 SDK 对接 DashScope 兼容接口。

---

## 当前功能

### 教师端

- 登录 / 注册 / 修改密码
- 题库 CRUD
- AI 出题
- AI 测试连接与诊断卡片
- 智能组卷
- 试卷管理
- 试卷编辑
- 按班级发布试卷
- 班级管理
- 班级成员管理（批量加入 / 移出）
- 查看单学生考试记录
- 查看试卷提交情况

### 学生端

- 查看已发布考试与历史答题记录
- 开始答题
- 自动保存答案
- 倒计时
- 提交试卷
- 查看成绩与答案

### 管理员端

- 拥有全部教师端能力
- 用户管理

---

## 技术栈

### 前端

- React 18
- TypeScript
- Vite
- Ant Design
- React Router
- Day.js

### 后端

- Go 1.22
- Gin
- GORM
- SQLite
- `openai-go` 官方 SDK

---

## 目录结构

```text
week05/homework
├── client/                 # React 前端
│   ├── src/
│   │   ├── api/            # 请求封装
│   │   ├── components/     # 复用组件
│   │   ├── hooks/          # 页面状态与副作用
│   │   ├── pages/          # 页面
│   │   └── types/          # 类型定义
│   └── dist/               # 构建产物
├── server/                 # Go 后端
│   ├── bootstrap/          # 启动装配、路由注册
│   ├── config/             # 配置加载
│   ├── database/           # DB 初始化
│   ├── handlers/           # HTTP handler
│   ├── models/             # GORM 模型
│   ├── repositories/       # 数据访问层
│   ├── services/           # 认证 / AI 服务
│   └── data/questions.db   # SQLite 数据库
├── docs/                   # 项目文档
├── config.json             # AI 配置
└── CLAUDE.md               # 架构与开发规范说明
```

---

## 环境要求

- Go >= 1.22
- Node.js >= 18
- `pnpm` >= 10

---

## 配置说明

### 1. 环境变量

在项目根目录创建 `.env`：

```env
DASHSCOPE_API_KEY=your_api_key_here
```

### 2. AI 配置

项目根目录 `config.json` 目前采用 OpenAI 兼容配置方式，例如：

```json
{
  "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",
  "model": "qwen3.5-plus",
  "requestTimeoutSeconds": 120,
  "systemPromptLines": [
    "..."
  ]
}
```

说明：

- `baseURL` 只需要写到兼容根路径 `/v1`
- 不需要手动写 `/chat/completions`
- 出题请求会自动追加 `/no_think`
- 前端支持“测试连接”按钮，会发送一个简短的 `who are you`

---

## 默认账号

系统启动时会自动确保以下默认账号存在：

- `admin / admin123`
- `teacher01 / teacher123`
- `student01 / student123`

---

## 本地开发

### 1. 安装依赖

```bash
cd server
go mod tidy

cd ../client
pnpm install
```

### 2. 启动后端

```bash
cd server
go run .
```

默认端口：

- `http://localhost:8080`

### 3. 启动前端

```bash
cd client
pnpm dev
```

默认端口：

- `http://localhost:3000`

开发环境下，前端会将 `/api` 请求代理到后端。

---

## 生产构建

### 前端构建

```bash
cd client
pnpm build
```

### 后端运行

```bash
cd server
go run .
```

后端会自动托管 `client/dist` 静态资源。

---

## 常用验证命令

### 前端

```bash
cd client
pnpm build
```

### 后端

```bash
cd server
go build ./...
```

---

## 主要页面

### 教师端

- `/questions`
  题库管理
- `/papers/generate`
  智能组卷
- `/papers`
  试卷管理
- `/papers/:id/edit`
  试卷编辑
- `/classes`
  班级管理
- `/users`
  用户管理（仅管理员）

### 学生端

- `/home`
  学生个人中心、课程文档与考试历史
- `/exam`
  兼容旧入口，自动跳转到 `/home`
- `/exam/:id/take`
  答题页
- `/exam/:id/result`
  成绩页

---

## 当前已知情况

### 已完成

- 认证体系
- 题库 CRUD
- AI 出题与诊断
- 智能组卷
- 试卷发布与管理
- 学生考试主链路
- 班级管理
- 教师统计视角

### 尚未完成

- 编程题完整闭环
- 监考增强面板
- 成绩分析与导出
- 自动化测试体系

---

## 文档入口

- [CLAUDE.md](./CLAUDE.md)
  当前架构、开发规范、约束说明
- [docs/progress.md](./docs/progress.md)
  当前开发状态、已完成内容、下一步优先级

---

## 后续方向

当前推荐优先级：

1. 编程题闭环
2. 监考增强
3. 成绩分析与导出
4. 工程质量治理（自动化测试、分包、CI）

如果你要继续开发，建议优先阅读：

1. `CLAUDE.md`
2. `docs/progress.md`
3. `server/bootstrap/router.go`
4. `client/src/App.tsx`
