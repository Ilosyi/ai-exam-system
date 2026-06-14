# 后端架构说明

## 分层约定

后端默认遵循 `bootstrap -> handlers -> repositories/services` 的结构：

- `server/bootstrap` 负责应用装配、依赖初始化、路由注册、静态资源挂载。
- `server/handlers` 负责 HTTP 参数校验、权限校验和响应组织。
- `server/repositories` 负责数据访问、查询逻辑和持久化细节。
- `server/services` 负责领域业务校验、外部服务封装和跨 repository 的业务编排。
- `server/models` 负责 GORM 模型定义。
- `server/database` 负责 SQLite 初始化和迁移。

## 课程文档模块

课程文档模块采用 file-backed 存储：`course-docs/` 位于项目根目录，用于存放课程目录、`course.json` 元数据和 Markdown 文档正文，不进入 SQLite 主业务表，也不参与 GORM `AutoMigrate`。

- `server/repositories/document_repository.go` 负责安全读写 `course-docs/`，必须限制文件操作在该根目录内，避免路径穿越。
- `server/services/document_service.go` 负责课程 slug、文档 slug、标题、排序、重复冲突等业务校验。
- `server/handlers/document_handler.go` 负责 HTTP 参数绑定、鉴权结果使用和统一响应。
- 读接口登录即可访问，包括 `/api/documents/courses`、`/api/documents/courses/:courseId`、`/api/documents/courses/:courseId/docs/:docId`。
- 写接口只允许 `admin` 和 `teacher` 访问，路由归属为 `/api/documents/courses*`。

当前课程文档不支持图片或附件上传，只管理课程元数据与 Markdown 文本。
