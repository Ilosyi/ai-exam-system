# 课程文档管理与学生首页改造设计

## 背景

当前系统已经具备教师后台、题库、组卷、考试发布、学生考试列表、答题和结果查看能力。新需求希望先跳过既有迭代计划，补齐课程 Markdown 文档管理能力，并把学生入口从单一考试列表升级为类似个人中心的 `/home` 页面。

本设计基于已确认的需求边界：

- 文档根目录为项目根目录下的 `course-docs/`，与 `client/`、`server/` 同级。
- 文档按课程分组，每个课程下包含若干 `.md` 文档。
- 管理员和所有老师都能管理课程和文档，不区分角色、班级或资源归属。
- 登录用户都能查看全部课程资料；学生在 `/home` 查看课程与考试。
- 不做图片或附件上传，Markdown 中的图片暂时使用外链或后续单独扩展。
- 文档打开后进入独立阅读页，布局接近参考图 3。
- `/exam` 保留兼容跳转到 `/home`；答题和结果页继续使用 `/exam/:id/take` 与 `/exam/:id/result`。

## 目标

1. 新增后台“文档管理”界面，延续现有 Ant Design 后台风格。
2. 新增文件型课程文档模块，支持课程和 Markdown 文档的创建、编辑、删除和排序。
3. 将学生默认入口改为 `/home`，展示个人信息、课程列表和考试列表。
4. 新增 Markdown 阅读页，提供课程文档导航、标题目录和正文渲染。
5. 更新项目文档，说明新模块、目录约定、路由变化和回归验证点。

## 非目标

- 不新增图片、附件上传或资源管理。
- 不新增用户资料表；个人信息卡片先基于现有 `AuthUser` 映射，保留未来扩展入口。
- 不改动考试答题、自动保存、交卷、阅卷等核心后端流程。
- 不做课程与班级、教师归属或学生权限绑定；课程资料当前全员可见。

## 文件存储结构

课程文档使用文件系统作为主存储：

```text
course-docs/
  <course-slug>/
    course.json
    <doc-slug>.md
```

`course.json` 保存课程元数据和文档列表：

```json
{
  "id": "backend-go",
  "title": "服务端训练营",
  "description": "Go 后端基础课",
  "order": 1,
  "documents": [
    {
      "id": "day01-gin",
      "title": "DAY01 - Gin 基础",
      "order": 1
    }
  ]
}
```

约束：

- `id` 使用 slug，只允许小写字母、数字、短横线和下划线。
- 课程目录名与课程 `id` 一致。
- Markdown 文件名与文档 `id` 一致。
- 后端对所有文件路径做清理和根目录校验，禁止 `../` 路径穿越。
- 课程和文档排序使用 `order` 字段；同值时按标题或 id 稳定排序。

## 后端设计

新增文档模块，保持现有分层：

- `server/handlers/document_handler.go`
  负责请求参数校验、角色判断后的响应组织。
- `server/services/document_service.go`
  负责 slug 校验、业务规则、课程/文档增删改流程。
- `server/repositories/document_repository.go`
  负责 `course-docs/` 下的文件读写、目录扫描和原子化保存。

建议 API：

- 登录用户可读：
  - `GET /api/documents/courses`
  - `GET /api/documents/courses/:courseId`
  - `GET /api/documents/courses/:courseId/docs/:docId`
- 教师/管理员可管理：
  - `POST /api/documents/courses`
  - `PUT /api/documents/courses/:courseId`
  - `DELETE /api/documents/courses/:courseId`
  - `POST /api/documents/courses/:courseId/docs`
  - `PUT /api/documents/courses/:courseId/docs/:docId`
  - `DELETE /api/documents/courses/:courseId/docs/:docId`

响应格式遵循项目约定：

- 列表：`{"data": [...], "total": n}`
- 单对象：`{"data": {...}}`
- 动作成功：`{"message": "..."}`
- 错误：`{"message": "...", "error": "..."}`

权限：

- 所有已登录用户都能读取课程与文档详情。
- `admin` 和 `teacher` 能管理课程与文档。
- `student` 不能访问写接口。

错误处理：

- 文档根目录不存在时自动创建。
- 课程不存在、文档不存在返回 404。
- slug 非法、slug 冲突、标题为空返回 400。
- 删除课程前端二次确认；后端执行时删除该课程目录。

## 前端设计

### 后台文档管理页

新增 `/documents` 页面，挂在现有 `AppLayout` 内，菜单新增“文档管理”，仅 `admin/teacher` 可见。

页面布局：

- 左侧：课程和文档树，支持选择课程或文档。
- 右侧：编辑区域。
  - 选择课程时编辑课程标题、描述、排序。
  - 选择文档时编辑文档标题、id、排序和 Markdown 正文。
- 操作区：新增课程、新增文档、保存、删除、预览。
- 预览：使用抽屉或右侧预览区渲染 Markdown。

编辑体验保持克制，不引入复杂 Markdown 编辑器；正文使用大文本框，确保教学项目可读、可维护。

### 学生首页 `/home`

学生默认入口从 `/exam` 改为 `/home`。`/home` 不使用教师后台 `AppLayout`，采用独立学生端样式，尽可能复刻参考图 1、2：

- 顶部：页面标题“个人中心”和退出登录。
- 个人信息卡片：
  - 圆形头像占位，显示用户名或首字。
  - 主标题显示用户名。
  - 基础信息显示 ID、角色、状态、班级 ID。
  - 实现上先映射现有 `AuthUser`，保留未来扩展为 profile 数据的展示模型。
- 课程列表：
  - 三列卡片网格。
  - 每张课程卡显示课程标题、描述、文档列表。
  - 点击文档进入阅读页。
- 我的考试：
  - 复用现有 `/api/exam/published` 数据。
  - 使用参考图 2 的卡片网格样式。
  - 当前在考试时间内显示进入答题；未开始或已结束显示不可进入状态。

兼容策略：

- `/exam` 自动跳转到 `/home`。
- `/exam/:id/take` 和 `/exam/:id/result` 保持不变。

### 文档阅读页

新增独立阅读页，路由建议为：

```text
/home/courses/:courseId/docs/:docId
```

页面结构接近参考图 3：

- 顶部标题栏：
  - 当前文档标题。
  - 返回首页或切换课件入口。
- 左侧栏：
  - 当前课程下的文档列表。
  - 当前 Markdown 正文解析出的标题目录。
- 右侧正文：
  - 使用 `react-markdown` 渲染。
  - 开启 `remark-gfm`、`rehype-highlight`。
  - 使用 `github-markdown-css` 保证 Markdown 基础排版。

## 数据流

`/home` 首屏并行加载：

- 当前用户：来自 `useAuth` 已有状态。
- 课程列表：`GET /api/documents/courses`。
- 考试列表：`GET /api/exam/published`。

阅读页加载：

1. 根据路由参数请求课程详情或课程列表，得到课程文档导航。
2. 请求文档详情，得到 Markdown 正文。
3. 前端从 Markdown 中提取标题，生成左侧目录。

后台管理页加载：

1. 请求课程列表。
2. 选择课程或文档后加载详情。
3. 保存时调用对应写接口。
4. 保存成功后刷新课程树，并保留当前选择。

## 测试与验证

自动化验证：

```bash
cd server
go build ./...
```

```bash
cd client
pnpm build
```

重点手工验证：

1. `admin/teacher` 能看到并进入“文档管理”。
2. `student` 默认登录后进入 `/home`。
3. `/exam` 跳转到 `/home`。
4. `/home` 展示个人信息、课程列表、考试列表。
5. 点击课程文档能进入阅读页并渲染 Markdown。
6. 后台新增、编辑、删除课程和文档后，学生端能看到变化。
7. 非法 slug、重复 slug、读取不存在文档时有明确错误提示。

## 文档更新范围

实现完成后同步更新：

- `AGENTS.md`
  - 增加文档管理模块说明。
  - 更新前端主要页面，加入 `/home`、`/documents` 和文档阅读页。
  - 更新路由权限说明。
  - 增加 `course-docs/` 目录约定。
- `docs/`
  - 补充文档模块架构与回归验证点，或更新现有后端架构说明。

## 待实施拆分

获批后进入实现计划阶段，建议按以下顺序拆分：

1. 后端文件型文档 repository/service/handler/API。
2. 前端 documents API 与类型定义。
3. 后台文档管理页。
4. 学生 `/home` 页面与 `/exam` 兼容跳转。
5. Markdown 阅读页。
6. 文档更新、构建验证和浏览器验收。
