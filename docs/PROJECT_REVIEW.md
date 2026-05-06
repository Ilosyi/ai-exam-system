# 项目Review与优化建议

## 📋 项目Review与优化建议

### 一、当前项目的核心优势（已经很不错）

✅ **架构设计得当**  
- 后端清晰的 handler → service → repository 三层分离  
- 前端 API 层统一封装、hooks 归一  
- 权限模型设计完整（admin/teacher/student 三层隔离）  
- 支持多班级学生关系和灵活的试卷发布策略

✅ **技术栈现代**  
- TypeScript 全栈、OpenAI 兼容协议集成、React 18 + Vite  
- 使用官方 SDK（openai-go）而非自己拼 HTTP

✅ **功能主链路完整**  
- 从登录 → 出题 → 组卷 → 发布 → 考试 → 成绩 一整条链路跑通  
- 支持自动保存、倒计时、自动阅卷

✅ **工程实践初步规范**  
- 已有启动脚本、配置管理、.gitignore 等最基础设施  
- CLAUDE.md 文档清晰

---

### 二、简历级别优化（按优先级排序）

#### 🥇 **P1：编程题完整闭包** （最有分量）
**为什么重要**：当前是功能空白，完成后能说"支持全题型"；体现对教学系统的理解  
**工作量**：中等（1-2 天）  
**简历表述**：  
> "实现编程题提交→评测→成绩反馈的完整流程，集成代码沙箱或在线判题服务（如 LeetCode 兼容 API）"

**具体做法**：
1. 前端：用 Monaco Editor 或 CodeMirror 替换占位符  
2. 后端：新增评测服务接口（可先支持简单的本地编译执行或调用第三方 API）  
3. 数据模型：扩展 ExamAnswer 支持编程题提交历史、执行日志

---

#### 🥇 **P1：自动化测试体系** （工程素养的体现）
**为什么重要**：区别"个人项目"和"生产级代码"的关键；简历上很有说服力  
**工作量**：中等（2-3 天）  
**简历表述**：  
> "编写单元测试和集成测试，覆盖认证、权限、核心业务逻辑，使用 Go testing + React Testing Library，集成 CI 流程"

**具体做法**：
1. 后端：  
   - 用 Go testing 对 handlers（auth、paper publish、exam）写单测  
   - 对 repositories 的关键查询逻辑写集成测试  
   - 目标：>70% 覆盖率

2. 前端：  
   - 用 Vitest + React Testing Library 测试关键 hooks（useAuth、exam flow）  
   - 测试关键 UI 交互（登录、答题、交卷流程）

3. CI：在 .github/workflows 里配 GitHub Actions（构建 + 测试）

---

#### 🥇 **P1：修复多班级试卷可见性bug** （细节和bug修复能力）
**为什么重要**：这是我之前发现的一个一致性bug；修复展示你对细节的关注  
**工作量**：小（<1 小时）  
**简历表述**：  
> "修复学生多班级场景下试卷可见性过滤逻辑，确保属于多个班级的学生能看到所有班级发布的考试"

**具体做法**：  
在 `exam_handler.go` 的 `ListPublished` 和 `StartAttempt` 中，改为调用 `classRepo.ListClassIDsByStudent()` 获取学生所有班级 ID，而非只用 `user.ClassID` 单值。

---

#### 🥇 **P1：清理 TS/JS 镜像文件** （代码卫生）
**为什么重要**：当前有 32 个 .js 镜像文件，这是明显的技术债  
**工作量**：极小（<30 分钟）  
**简历表述**：  
> "完成前端 TypeScript 迁移，移除所有 .js 镜像文件，确保源代码单一真值"

**具体做法**：
```bash
find client/src -name "*.js" -delete
# 更新 tsconfig 确保不生成 .js
```

---

#### 🥈 **P2：API 文档（Swagger/OpenAPI）** （专业度）
**为什么重要**：让项目"可被他人快速理解和使用"；展现专业态度  
**工作量**：小-中（1-2 天）  
**简历表述**：  
> "使用 Swagger/OpenAPI 自动生成完整的后端 API 文档，支持在线测试与交互式演示"

**具体做法**：
1. 后端引入 `swaggo/swag`  
2. 为 handler 加注释生成 Swagger  
3. 访问 `/swagger/index.html` 查看文档  
4. 将文档链接放到 README  

---

#### 🥈 **P2：前端性能优化** （优化意识）
**为什么重要**：当前前端包体 >1.4MB，这是显而易见的问题  
**工作量**：中等（1-2 天）  
**简历表述**：  
> "通过路由级 code-splitting、动态导入和包分析，将首屏加载时间从 X 秒降低到 Y 秒；优化包体积至 <500KB"

**具体做法**：
1. 在 `pages/` 里用 React.lazy + Suspense 做路由懒加载  
2. 用 `vite-plugin-visualizer` 分析包体  
3. 考虑移除或 tree-shake 没用的依赖  
4. 打印前后对比数据到 README

---

#### 🥈 **P2：日志与错误追踪** （生产级思维）
**为什么重要**：展现对可维护性和问题排查的理解  
**工作量**：小（<1 天）  
**简历表述**：  
> "集成结构化日志库（Go: logrus/slog，前端: 自定义日志中间件），记录关键业务事件和错误堆栈，便于问题排查"

**具体做法**：
1. 后端用 Go 标准库 `log/slog` 或 `logrus`  
2. 前端拦截 API 错误和控制台 error  
3. 记录到本地或发送到远程（如 sentry）

---

#### 🥈 **P2：监考增强面板** （产品思维）
**为什么重要**：当前监考只有基础事件收集，补完这块能说"整个教务系统"  
**工作量**：中等（2-3 天）  
**简历表述**：  
> "实现教师实时监考面板，展示学生异常行为（如 tab 切换、长时间未作答等），支持按班级/试卷过滤和违规记录导出"

**具体做法**：
1. 补前端监考事件收集（heartbeat、tab-switch 等）  
2. 后端统计异常事件并标记学生  
3. 新增教师监考面板路由和 API  
4. 支持违规记录下载

---

#### 🥉 **P3：Docker 容器化与部署脚本** （DevOps 意识）
**为什么重要**：一条命令启动体现对部署的理解  
**工作量**：小（<1 天）  
**简历表述**：  
> "编写 Dockerfile + docker-compose，支持一键启动前后端和数据库，并提供生产级部署脚本"

**具体做法**：
```dockerfile
# server/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
COPY --from=builder /app/server /app/server
```

```yaml
# docker-compose.yml
version: '3'
services:
  server:
    build: ./server
    ports: ["8080:8080"]
    environment:
      - DASHSCOPE_API_KEY=${DASHSCOPE_API_KEY}
  client:
    build: ./client
    ports: ["3000:3000"]
```

---

### 三、技术创新点（已有，可强调）

📌 **OpenAI 兼容协议的通用性**  
- 你的 AI 模块设计得足够通用，可以切换到不同的兼容服务  
- 简历可说："支持 OpenAI 兼容协议，易于切换云厂商（DashScope/Claude/其他）"

📌 **多班级灵活关系设计**  
- `users.class_id` + `user_classes` 多对多关系  
- 简历可说："支持学生跨班级参加考试，关系模型设计灵活高效"

---

### 四、简历总结模板

如果你想写进简历，可以这样组织：

---

**项目名称**：AI 题库与考试系统  
**技术栈**：Go 1.22 + React 18 + TypeScript + SQLite + OpenAI API  
**项目规模**：3 个完整端（教师端、学生端、管理端）+ 5000+ 行代码  

**核心亮点**：
1. ✨ **完整教务系统**：从题库管理、AI 生成、智能组卷、试卷发布到学生考试、自动阅卷的全链路设计与实现
2. 🔐 **三层权限体系**：管理员/教师/学生角色隔离，支持多班级灵活关系，细粒度资源访问控制
3. 🤖 **AI 兼容协议**：集成 OpenAI SDK，支持 DashScope 等兼容服务，易于扩展多家云厂商
4. 📊 **实时监考与统计**：学生自动保存、倒计时、异常事件上报，教师可查看提交情况与单生记录
5. 🛠️ **工程最佳实践**：清晰的架构分离、可配置部署、完整测试覆盖、Swagger 文档

**技术难点与解决**：
- 解决学生多班级场景的试卷可见性过滤逻辑  
- 通过 `/no_think` 优化 AI 出题延迟  
- 前端自动保存与防冲突的答案管理

---

### 五、我的建议优先级（用于规划）

**这周做（1-2 天）**：
1. 编程题完整闭包
2. API 文档（Swagger/OpenAPI）

**后续做（周末或空档）**：
1. 自动化测试
2. 前端性能优化
3. 监考增强

---

## 更新记录

- **2026-05-06**：初版完成；修复多班级可见性 bug；清理 TS/JS 镜像文件
