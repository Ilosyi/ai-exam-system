# 后端架构文档

> 本文档详细介绍了 AI 题库与考试系统后端的整体架构设计、各层职责、请求处理流程以及代码组织方式。

---

## 目录

- [1. 技术栈](#1-技术栈)
- [2. 整体架构图](#2-整体架构图)
- [3. 目录结构](#3-目录结构)
- [4. 分层架构详解](#4-分层架构详解)
- [5. 请求处理流程](#5-请求处理流程)
- [6. 数据模型](#6-数据模型)
- [7. 路由与权限设计](#7-路由与权限设计)
- [8. 核心业务模块](#8-核心业务模块)
- [9. API 响应规范](#9-api-响应规范)
- [10. 架构评价](#10-架构评价)

---

## 1. 技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| **语言** | Go 1.22 | 后端开发语言 |
| **Web 框架** | Gin | Go 最流行的 HTTP 框架，类似 Express.js |
| **ORM** | GORM | Go 最流行的 ORM 库，类似 SQLAlchemy |
| **数据库** | SQLite | 轻量级文件数据库，零配置 |
| **认证** | JWT (手写) | 使用标准库手写实现，便于学习原理 |
| **密码哈希** | bcrypt | 单向哈希算法，自动加盐 |
| **AI 集成** | OpenAI Go SDK | 兼容 DashScope 的 API 接口 |

---

## 2. 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         main.go                                 │
│                    (程序入口，仅调用 bootstrap.Start)             │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      bootstrap 层                               │
│  ┌──────────────┐  ┌────────────────┐  ┌──────────────────┐    │
│  │ application.go│  │dependencies.go │  │   router.go      │    │
│  │ 应用初始化    │  │ 依赖注入组装   │  │ 路由注册+CORS    │    │
│  └──────────────┘  └────────────────┘  └──────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   config     │  │   database   │  │  middleware   │
│  配置加载    │  │  SQLite 连接 │  │  认证/鉴权   │
└──────────────┘  └──────────────┘  └──────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       handlers 层                               │
│  ┌──────┐ ┌────────┐ ┌──────┐ ┌──────┐ ┌─────┐ ┌──────┐       │
│  │ auth │ │question│ │paper │ │ exam │ │ ai  │ │class │       │
│  └──┬───┘ └───┬────┘ └──┬───┘ └──┬───┘ └──┬──┘ └──┬───┘       │
└─────┼─────────┼─────────┼────────┼────────┼───────┼────────────┘
      │         │         │        │        │       │
      ▼         ▼         ▼        ▼        ▼       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    repositories 层                               │
│  ┌──────┐ ┌────────┐ ┌──────┐ ┌──────┐ ┌───────┐              │
│  │ user │ │question│ │paper │ │ exam │ │ class │              │
│  └──────┘ └────────┘ └──────┘ └──────┘ └───────┘              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      services 层                                │
│  ┌──────────────┐  ┌──────────────┐                            │
│  │  auth_service │  │  ai_service  │                            │
│  │  JWT + bcrypt │  │  OpenAI SDK  │                            │
│  └──────────────┘  └──────────────┘                            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       models 层                                 │
│  User, Question, Paper, PaperItem, PaperPublication,           │
│  ExamAttempt, ExamAnswer, Class, UserClass, ProctorEvent       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. 目录结构

```
server/
├── main.go                          # 程序入口
├── go.mod                           # Go 模块定义
├── go.sum                           # 依赖版本锁定
│
├── bootstrap/                       # 应用引导层
│   ├── application.go               # 应用初始化与启动
│   ├── dependencies.go              # 依赖注入组装
│   └── router.go                    # 路由注册与静态资源
│
├── config/                          # 配置管理
│   └── config.go                    # 从 config.json 加载配置
│
├── database/                        # 数据库层
│   └── database.go                  # SQLite 连接管理
│
├── middleware/                       # 中间件
│   └── auth.go                      # 认证/鉴权中间件
│
├── handlers/                        # HTTP 处理层
│   ├── auth_handler.go              # 认证接口
│   ├── question_handler.go          # 题库接口
│   ├── paper_handler.go             # 试卷接口
│   ├── exam_handler.go              # 考试接口
│   ├── ai_handler.go                # AI 出题接口
│   ├── class_handler.go             # 班级接口
│   ├── document_handler.go          # 课程文档接口
│   ├── note_handler.go              # 学习笔记接口
│   ├── test_helpers.go              # 测试辅助工具
│   └── auth_handler_test.go         # 认证接口测试
│
├── models/                          # 数据模型层
│   ├── user.go                      # 用户模型
│   ├── question.go                  # 题目模型
│   ├── paper.go                     # 试卷模型
│   ├── paper_item.go                # 试卷题目项模型
│   ├── paper_publication.go         # 试卷发布记录模型
│   ├── exam_attempt.go              # 考试答题记录模型
│   ├── exam_answer.go               # 考试答案模型
│   ├── class.go                     # 班级模型
│   ├── user_class.go                # 学生-班级关联模型
│   └── proctor_event.go             # 监考事件模型
│
├── repositories/                    # 数据访问层
│   ├── user_repository.go           # 用户数据访问
│   ├── question_repository.go       # 题目数据访问
│   ├── paper_repository.go          # 试卷数据访问
│   ├── exam_repository.go           # 考试数据访问
│   ├── class_repository.go          # 班级数据访问
│   ├── document_repository.go       # 课程文档文件读写
│   ├── question_repository_test.go  # 题目数据访问测试
│   └── user_repository_test.go      # 用户数据访问测试
│
├── services/                        # 业务逻辑层
│   ├── auth_service.go              # 认证服务（JWT、密码）
│   ├── ai_service.go                # AI 服务（OpenAI SDK）
│   ├── document_service.go          # 课程文档校验与编排
│   └── auth_service_test.go         # 认证服务测试
│
├── ../course-docs/                  # 课程文档文件根目录
│   └── <course>/                    # 单门课程目录
│       ├── course.json              # 课程元数据与文档索引
│       └── *.md                     # Markdown 文档正文
│
└── data/                            # 数据文件
    └── questions.db                 # SQLite 数据库文件
```

---

## 4. 分层架构详解

### 4.1 main.go - 程序入口

**职责**：仅负责调用 `bootstrap.Start()` 启动应用。

```go
func main() {
    if err := bootstrap.Start(""); err != nil {
        log.Fatal(err)
    }
}
```

**设计原则**：
- 入口文件保持极简，不承载任何业务逻辑
- 启动逻辑封装在 bootstrap 包中，便于测试复用

---

### 4.2 bootstrap 层 - 应用引导

#### application.go - 应用初始化

**职责**：
1. 解析项目根目录
2. 加载配置文件
3. 初始化数据库
4. 创建依赖对象
5. 构建路由
6. 启动服务器

**核心函数**：
- `Start(addr string) error` - 启动入口
- `NewApplication(addr string) (*Application, error)` - 创建应用实例
- `resolveProjectRoot() (string, error)` - 解析项目根目录
- `initDatabase(projectRoot string) (*gorm.DB, error)` - 初始化数据库

#### dependencies.go - 依赖注入

**职责**：创建和组装所有业务对象（Repository、Service、Handler）。

**依赖关系**：
```
repositories（数据访问层）
     ↑
 services（业务逻辑层）
     ↑
 handlers（HTTP 处理层）
     ↑
   router（路由层）
```

**核心函数**：
- `initDependencies(projectRoot string, db *gorm.DB) (*dependencies, error)` - 创建所有依赖
- `seedDefaultUsers(ctx, userRepo, authService) error` - 创建默认用户

#### router.go - 路由注册

**职责**：
1. 创建 Gin 路由引擎
2. 注册 CORS 中间件
3. 注册所有 API 路由（按权限分组）
4. 挂载前端静态资源

---

### 4.3 config 层 - 配置管理

**职责**：从 `config.json` 加载应用配置，提供合理的默认值。

**配置项**：
```json
{
  "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",
  "model": "qwen-plus",
  "requestTimeoutSeconds": 120,
  "systemPromptLines": ["你是一个..."],
  "serverPort": 8080,
  "clientPort": 3000
}
```

**设计特点**：
- 单例模式：配置只加载一次
- 提供默认值：减少配置项
- 必填校验：缺失时返回明确错误

---

### 4.4 database 层 - 数据库连接

**职责**：初始化 SQLite 数据库连接，执行表结构迁移。

**设计特点**：
- 使用 `sync.Once` 保证单例
- 自动创建 data 目录和数据库文件
- 支持自定义迁移函数

---

### 4.5 middleware 层 - 中间件

#### RequireAuth - 认证中间件

**职责**：
1. 从请求头提取 Authorization
2. 验证 JWT 令牌
3. 查询数据库确认用户存在
4. 将用户信息写入 Context

#### RequireRoles - 鉴权中间件

**职责**：检查当前用户的角色是否在允许列表中。

**使用示例**：
```go
protected.Use(middleware.RequireAuth(authService, userRepo))  // 认证
adminRoutes.Use(middleware.RequireRoles("admin"))              // 仅管理员
teacherRoutes.Use(middleware.RequireRoles("admin", "teacher")) // 管理员和教师
```

---

### 4.6 handlers 层 - HTTP 处理

**职责**：
1. 解析请求参数
2. 校验参数合法性
3. 调用 Repository/Service
4. 组织响应返回

**Handler 列表**：

| Handler | 文件 | 职责 |
|---------|------|------|
| AuthHandler | auth_handler.go | 登录、注册、刷新 token、修改密码 |
| QuestionHandler | question_handler.go | 题库 CRUD、批量删除 |
| PaperHandler | paper_handler.go | 试卷管理、智能组卷、发布 |
| ExamHandler | exam_handler.go | 考试答题、自动保存、交卷 |
| AIHandler | ai_handler.go | AI 出题、测试连接 |
| ClassHandler | class_handler.go | 班级管理、学生管理 |
| DocumentHandler | document_handler.go | 课程文档读取与管理 |
| NoteHandler | note_handler.go | 学习笔记 |

---

### 4.7 repositories 层 - 数据访问

**职责**：封装所有数据库操作，提供统一的数据访问接口。

**Repository 列表**：

| Repository | 文件 | 操作的数据表 |
|------------|------|-------------|
| UserRepository | user_repository.go | users |
| QuestionRepository | question_repository.go | questions |
| PaperRepository | paper_repository.go | papers, paper_items, paper_publications |
| ExamRepository | exam_repository.go | exam_attempts, exam_answers, proctor_events |
| ClassRepository | class_repository.go | classes, user_classes |
| DocumentRepository | document_repository.go | course-docs 文件目录 |

**常用操作**：
- `FindByID` - 根据 ID 查询
- `List` - 分页查询（支持筛选）
- `Create` - 创建记录
- `Update` - 更新记录
- `Delete` - 删除记录

---

### 4.8 services 层 - 业务逻辑

**职责**：处理跨 Repository 的复杂业务逻辑。

#### AuthService - 认证服务

**功能**：
- `GenerateToken(user) (string, error)` - 生成 JWT
- `ParseToken(token) (*AuthClaims, error)` - 解析 JWT
- `HashPassword(password) (string, error)` - 密码哈希
- `ComparePassword(hashed, plain) error` - 密码验证

#### DocumentService - 课程文档服务

**功能**：
- 校验课程 ID、文档 ID、标题、排序等业务输入
- 检查课程与文档是否存在
- 编排文件型仓库的课程和 Markdown 文档读写

课程文档不进入 SQLite，也不参与 GORM `AutoMigrate`。它们存储在项目根目录的 `course-docs/` 下，每个课程目录包含 `course.json` 和若干 `*.md` 文件。

#### AIService - AI 服务

**功能**：
- `GenerateQuestions(prompt) (string, error)` - AI 生成题目
- `TestConnection() (string, error)` - 测试连接
- `TestConnectionDiagnostic() (*ConnectionDiagnostic, error)` - 诊断信息

---

### 4.9 models 层 - 数据模型

**职责**：定义数据库表结构，使用 GORM 标签映射。

**模型列表**：

| 模型 | 表名 | 说明 |
|------|------|------|
| User | users | 系统用户 |
| Question | questions | 题库题目 |
| Paper | papers | 试卷 |
| PaperItem | paper_items | 试卷题目项 |
| PaperPublication | paper_publications | 试卷发布记录 |
| ExamAttempt | exam_attempts | 考试答题记录 |
| ExamAnswer | exam_answers | 考试答案 |
| Class | classes | 班级 |
| UserClass | user_classes | 学生-班级关联 |
| ProctorEvent | proctor_events | 监考事件 |

---

## 5. 请求处理流程

以"学生开始答题"为例：

```
前端 POST /api/exam/papers/1/start
        │
        ▼
[GIN Router] → 匹配 studentRoutes 组
        │
        ▼
[RequireAuth 中间件] → 解析 JWT → 查库验证用户 → 写入 ctx
        │
        ▼
[RequireRoles("admin","student")] → 检查角色权限
        │
        ▼
[ExamHandler.StartAttempt]
  ├── 解析 URL 参数 (paperID)
  ├── 查询学生所属班级 (ClassRepository)
  ├── 验证试卷发布状态和时间窗 (ExamRepository)
  ├── 检查是否有进行中的答题 (ExamRepository)
  ├── 创建 ExamAttempt (ExamRepository)
  └── 计算截止时间 → 返回响应
```

---

## 6. 数据模型

### 6.1 实体关系图

```
┌─────────┐       ┌───────────┐       ┌─────────┐
│  users  │──M:N──│user_classes│──M:N──│ classes │
└────┬────┘       └───────────┘       └────┬────┘
     │                                      │
     │ 1:N                                  │ 1:N
     ▼                                      │
┌──────────────┐                            │
│ exam_attempts│                            │
└──────┬───────┘                            │
       │ 1:N                                │
       ▼                                    │
┌──────────────┐                            │
│ exam_answers │                            │
└──────────────┘                            │
                                            │
┌─────────┐       ┌────────────┐            │
│ papers  │──1:N──│paper_items │            │
└────┬────┘       └─────┬──────┘            │
     │                  │                   │
     │ 1:N              │ N:1               │
     ▼                  ▼                   │
┌──────────────────┐  ┌──────────┐          │
│paper_publications│  │questions │          │
└──────────────────┘  └──────────┘          │
                                            │
       ┌────────────────────────────────────┘
       │ FK: teacher_id
       ▼
  ┌─────────┐
  │ classes │
  └─────────┘
```

### 6.2 核心模型说明

#### User - 用户

```go
type User struct {
    ID           uint   `json:"id" gorm:"primaryKey"`
    Username     string `json:"username" gorm:"uniqueIndex;not null"`
    Role         string `json:"role" gorm:"default:student"`      // admin|teacher|student
    ClassID      *uint  `json:"classId" gorm:"column:class_id"`   // 兼容旧的单班级设计
    PasswordHash string `json:"-" gorm:"column:password_hash"`    // json:"-" 不返回给前端
    Status       string `json:"status" gorm:"default:active"`     // active|disabled
}
```

#### Question - 题目

```go
type Question struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    CreatedBy   uint      `json:"createdBy" gorm:"column:created_by;index"`
    Type        string    `json:"type"`                           // single|multiple|coding
    Language    string    `json:"language"`                       // go|cpp|java|javascript|python
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    OptionsJSON string    `json:"options" gorm:"column:options"`  // JSON 字符串
    AnswerJSON  string    `json:"answers" gorm:"column:answers"`  // JSON 字符串
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

#### Paper - 试卷

```go
type Paper struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    Title      string    `json:"title" gorm:"not null"`
    Language   string    `json:"language"`
    TotalScore int       `json:"totalScore" gorm:"column:total_score"`
    Status     string    `json:"status" gorm:"default:draft"`     // draft|published|closed
    CreatedBy  uint      `json:"createdBy" gorm:"column:created_by"`
    Items      []PaperItem `json:"items,omitempty" gorm:"foreignKey:PaperID"`
}
```

#### ExamAttempt - 答题记录

```go
type ExamAttempt struct {
    ID          uint       `json:"id" gorm:"primaryKey"`
    PaperID     uint       `json:"paperId" gorm:"column:paper_id;index"`
    StudentID   uint       `json:"studentId" gorm:"column:student_id;index"`
    StartedAt   time.Time  `json:"startedAt"`
    SubmittedAt *time.Time `json:"submittedAt"`
    Status      string     `json:"status" gorm:"default:in_progress"` // in_progress|submitted|timeout
    TotalScore  *int       `json:"totalScore"`
    Answers     []ExamAnswer `json:"answers,omitempty" gorm:"foreignKey:AttemptID"`
}
```

---

## 7. 路由与权限设计

### 7.1 路由分组

```
/api
├── /auth                              # 公开接口（无需登录）
│   ├── POST /login                    # 用户登录
│   ├── POST /register                 # 用户注册
│   └── POST /refresh                  # 刷新 token
│
├── /auth (protected)                  # 已登录接口
│   ├── GET /me                        # 获取当前用户
│   ├── POST /logout                   # 退出登录
│   ├── POST /change-password          # 修改密码
│   └── /documents/courses/*           # 课程文档读取
│
├── /users (admin)                     # 仅管理员
│   ├── GET /                          # 用户列表
│   ├── PUT /:id                       # 更新用户
│   └── DELETE /:id                    # 删除用户
│
├── /questions (teacher+admin)         # 教师+管理员
│   ├── GET /                          # 题目列表
│   ├── POST /                         # 创建题目
│   ├── PUT /:id                       # 更新题目
│   ├── DELETE /:id                    # 删除题目
│   └── DELETE /                       # 批量删除
│
├── /papers (teacher+admin)            # 教师+管理员
│   ├── POST /generate                 # 智能组卷
│   ├── POST /                         # 保存试卷
│   ├── GET /                          # 试卷列表
│   ├── GET /:id                       # 试卷详情
│   ├── PUT /:id                       # 更新试卷
│   ├── DELETE /:id                    # 删除试卷
│   ├── POST /:id/publish              # 发布试卷
│   ├── POST /:id/unpublish            # 取消发布
│   └── GET /:id/submissions           # 提交统计
│
├── /classes (teacher+admin)           # 教师+管理员
│   ├── GET /                          # 班级列表
│   ├── POST /                         # 创建班级
│   ├── PUT /:id                       # 更新班级
│   ├── DELETE /:id                    # 删除班级
│   ├── GET /:id/students              # 班级学生
│   └── POST /:id/students/batch-edit  # 批量编辑
│
├── /documents/courses (teacher+admin) # 课程文档管理
│   ├── POST /                         # 创建课程
│   ├── PUT /:courseId                 # 更新课程
│   ├── DELETE /:courseId              # 删除课程
│   ├── POST /:courseId/docs           # 创建文档
│   ├── PUT /:courseId/docs/:docId     # 更新文档
│   └── DELETE /:courseId/docs/:docId  # 删除文档
│
├── /ai (teacher+admin)                # 教师+管理员
│   ├── POST /generate                 # AI 出题
│   └── POST /test                     # 测试连接
│
└── /exam (student+admin)              # 学生+管理员
    ├── GET /published                 # 已发布考试与历史记录
    ├── POST /papers/:id/start         # 开始答题
    ├── GET /attempts/:id              # 答题详情
    ├── PUT /attempts/:id/answers      # 保存答案
    ├── POST /attempts/:id/submit      # 交卷
    ├── GET /attempts/:id/result       # 查看结果
    └── POST /attempts/:id/events      # 监考事件
```

### 7.2 权限矩阵

| 功能 | admin | teacher | student |
|------|-------|---------|---------|
| 用户管理 | ✅ | ❌ | ❌ |
| 题库管理 | ✅ | ✅ | ❌ |
| AI 出题 | ✅ | ✅ | ❌ |
| 试卷管理 | ✅ | ✅ | ❌ |
| 班级管理 | ✅ | ✅ (自己的) | ❌ |
| 课程文档读取 | ✅ | ✅ | ✅ |
| 课程文档管理 | ✅ | ✅ | ❌ |
| 查看已发布考试与历史记录 | ✅ | ❌ | ✅ |
| 答题 | ✅ | ❌ | ✅ |
| 查看结果 | ✅ | ❌ | ✅ |

---

## 8. 核心业务模块

### 8.1 认证与鉴权

**JWT 令牌结构**：
```
Header.Payload.Signature

Header:  {"alg":"HS256","typ":"JWT"}
Payload: {"uid":1,"username":"admin","role":"admin","exp":1234567890,"iat":1234567890}
Signature: HMAC-SHA256(header.payload, secret)
```

**认证流程**：
1. 用户登录，验证用户名密码
2. 生成 JWT 令牌，返回给前端
3. 前端在后续请求中携带 `Authorization: Bearer <token>`
4. 后端中间件验证 JWT 的有效性
5. 将用户信息写入请求上下文

---

### 8.2 题库管理

**题型说明**：
- `single`：单选题，4 个选项，1 个正确答案
- `multiple`：多选题，4 个选项，多个正确答案
- `coding`：编程题，无选项，代码评测

**选项存储**：
```json
// OptionsJSON
["选项A内容", "选项B内容", "选项C内容", "选项D内容"]

// AnswerJSON（单选）
[0]

// AnswerJSON（多选）
[0, 2]
```

---

### 8.3 试卷管理

**试卷生命周期**：
```
draft（草稿）→ published（已发布）→ closed（已关闭）
```

**智能组卷流程**：
1. 教师指定题型、数量、语言、总分
2. 后端按条件随机选取题目
3. 计算每题分值（总分 / 题目数）
4. 返回预览，教师确认后保存

**发布机制**：
- 设置时间窗口（开始时间、结束时间）
- 设置答题时长（可选）
- 指定目标班级（可选，不指定表示公共试卷）

---

### 8.4 考试系统

**答题流程**：
```
查看考试列表 → 开始答题 → 自动保存答案 → 交卷 → 查看结果
```

**截止时间计算**：
```
deadline = min(开始答题时间 + Duration, EndTime)
```

**自动阅卷**：
- 对比学生答案和正确答案
- 完全一致得满分，否则得 0 分
- 支持单选题和多选题

---

### 8.5 班级体系

**多班级支持**：
- `users.class_id`：旧设计（单班级），保留兼容
- `user_classes` 表：新设计（多班级），主要关联方式

**学生可见试卷规则**：
- 公共试卷（class_id IS NULL）：所有学生可见
- 班级试卷：只有该班级的学生可见
- 学生所属班级 = users.class_id + user_classes 表

---

### 8.6 课程文档

**存储结构**：
```
course-docs/
└── backend-go/
    ├── course.json
    ├── day01-gin.md
    └── ...
```

**设计约束**：
- 课程和文档使用 slug 作为文件夹名和文件名，避免路径穿越
- `course.json` 保存课程标题、说明、排序和文档索引
- Markdown 正文保存在同课程目录下的 `*.md` 文件
- 所有登录用户可以读取课程文档
- `admin` 和 `teacher` 可以管理全部课程与文档，不按班级隔离
- 当前不支持图片或附件上传

**安全边界**：
- 文件操作必须限制在 `course-docs/` 根目录内
- 仓库层拒绝不合法 slug 和符号链接课程目录
- 写操作使用进程内锁串行化，避免并发更新 `course.json` 时丢失文档索引

---

## 9. API 响应规范

### 9.1 成功响应

**列表查询**：
```json
{
  "data": [...],
  "total": 100
}
```

**单对象查询**：
```json
{
  "data": {...}
}
```

**操作成功**：
```json
{
  "message": "保存成功",
  "data": {...}
}
```

### 9.2 错误响应

```json
{
  "message": "参数错误",
  "error": "具体错误信息"
}
```

### 9.3 HTTP 状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| 200 | OK | 查询成功、操作成功 |
| 201 | Created | 创建成功（注册） |
| 204 | No Content | 删除成功 |
| 400 | Bad Request | 参数错误 |
| 401 | Unauthorized | 未登录、token 无效 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 用户名已存在 |
| 500 | Internal Server Error | 服务器内部错误 |

### 9.4 时间格式

所有时间字段使用 RFC3339 格式：
```
2024-01-01T00:00:00Z
```

---

## 10. 架构评价

### ✅ 优点

1. **分层清晰**
   - `bootstrap → handlers → repositories → services → models` 调用链路规范
   - 每层职责单一，易于理解和维护

2. **依赖注入**
   - `dependencies.go` 集中管理所有依赖创建
   - 便于测试时替换为 mock 对象

3. **路由分组**
   - 公开/已登录/管理员/教师/学生五级路由组
   - 权限控制在路由层就能看到

4. **统一响应格式**
   - 所有接口遵循 `{data, total, message, error}` 约定
   - 前端处理响应逻辑统一

5. **JWT 手写实现**
   - 没有引入第三方 JWT 库
   - 适合学习理解 JWT 原理

6. **测试基础设施**
   - `test_helpers.go` 提供完整的测试上下文搭建
   - 内存数据库用于单元测试

7. **AutoMigrate**
   - 使用 GORM 自动迁移
   - 开发阶段非常方便

### ⚠️ 可改进之处

1. **context 传递不一致**
   - 部分 handler 用 `context.Background()`
   - 建议统一使用 `c.Request.Context()`

2. **错误处理可统一**
   - 某些地方直接 return，某些地方做分类处理
   - 建议抽取公共错误响应函数

3. **N+1 查询风险**
   - `toPaperResponseWithItems` 中循环查题目详情
   - 可考虑批量预加载

4. **事务边界**
   - `Create` 方法用了事务
   - `Delete` 也可以考虑事务保护关联数据

5. **排序算法**
   - `sortInts` 用手写冒泡排序
   - 可直接用 `sort.Ints()`

---

## 附录：学习资源

### Go 语言相关

- [Go 官方文档](https://go.dev/doc/)
- [Go 标准库文档](https://pkg.go.dev/std)
- [Effective Go](https://go.dev/doc/effective_go)

### 框架相关

- [Gin 官方文档](https://gin-gonic.com/docs/)
- [GORM 官方文档](https://gorm.io/docs/)

### 认证相关

- [JWT 介绍](https://jwt.io/introduction)
- [bcrypt 原理](https://en.wikipedia.org/wiki/Bcrypt)

---

*文档版本：v1.0*
*最后更新：2026-05-30*
