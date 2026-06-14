// ============================================================================
// bootstrap/router.go - 路由注册与静态资源服务
// ============================================================================
//
// 本文件负责：
// 1. 创建 Gin 路由引擎
// 2. 注册 CORS 中间件（跨域资源共享）
// 3. 注册所有 API 路由（按权限分组）
// 4. 挂载前端静态资源（SPA 应用）
//
// 路由分组策略（按权限从低到高）：
//   /api/auth/login, /api/auth/register, /api/auth/refresh  → 公开接口（无需登录）
//   /api/auth/me, /api/auth/logout, ...                     → 已登录接口
//   /api/users/*                                            → 仅管理员
//   /api/questions/*, /api/papers/*, /api/classes/*          → 教师 + 管理员
//   /api/exam/*                                             → 学生 + 管理员
//
// 学习要点：
// - Gin 的路由分组 (Group) 和中间件 (Use) 机制
// - CORS 是什么，为什么需要它
// - SPA（单页应用）的静态资源服务策略
// ============================================================================

package bootstrap

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
)

// buildRouter 构建完整的 HTTP 路由树。
//
// 这是路由注册的入口函数，按顺序完成：
// 1. 创建 Gin 默认引擎（自带 Logger 和 Recovery 中间件）
// 2. 注册 CORS 中间件（允许前端跨域请求）
// 3. 注册所有 API 路由
// 4. 挂载前端静态资源
//
// 参数：
//   - projectRoot: 项目根目录，用于定位前端构建产物 (client/dist)
//   - deps: 依赖对象集合，包含所有 Handler
//
// 返回值：
//   - *gin.Engine: 配置好的路由引擎，可以直接启动监听
func buildRouter(projectRoot string, deps *dependencies) *gin.Engine {
	// gin.Default() 创建一个带默认中间件的引擎：
	// - Logger: 打印请求日志（方法、路径、状态码、耗时）
	// - Recovery: 捕获 panic，返回 500 而不是让程序崩溃
	r := gin.Default()

	// 注册 CORS 中间件，允许前端跨域请求
	r.Use(corsMiddleware())

	// 注册所有 API 路由
	registerAPIRoutes(r, deps)

	// 挂载前端静态资源（SPA 应用的 index.html、JS、CSS 等）
	serveStatic(r, filepath.Join(projectRoot, "client", "dist"))

	return r
}

// registerAPIRoutes 注册所有 API 路由。
//
// 路由组织结构：
//
//	/api
//	├── /auth
//	│   ├── POST /login          (公开) 用户登录
//	│   ├── POST /register       (公开) 用户注册
//	│   ├── POST /refresh        (公开) 刷新 token
//	│   ├── GET  /me             (已登录) 获取当前用户信息
//	│   ├── POST /logout         (已登录) 退出登录
//	│   └── POST /change-password(已登录) 修改密码
//	├── /users                   (仅管理员)
//	│   ├── GET  /               用户列表
//	│   ├── PUT  /:id            更新用户
//	│   └── DELETE /:id          删除用户
//	├── /questions               (教师+管理员)
//	│   ├── GET  /               题目列表
//	│   ├── POST /               创建题目
//	│   ├── PUT  /:id            更新题目
//	│   ├── DELETE /:id          删除题目
//	│   └── DELETE /             批量删除
//	├── /papers                  (教师+管理员)
//	│   ├── POST /generate       智能组卷
//	│   ├── POST /               保存试卷
//	│   ├── GET  /               试卷列表
//	│   ├── GET  /:id            试卷详情
//	│   ├── PUT  /:id            更新试卷
//	│   ├── DELETE /:id          删除试卷
//	│   └── ...                  (发布、取消发布、提交统计等)
//	├── /classes                 (教师+管理员)
//	│   ├── GET  /               班级列表
//	│   ├── POST /               创建班级
//	│   ├── PUT  /:id            更新班级
//	│   ├── DELETE /:id          删除班级
//	│   └── /:id/students/*      班级学生管理
//	├── /ai                      (教师+管理员)
//	│   ├── POST /generate       AI 出题
//	│   └── POST /test           测试 AI 连接
//	└── /exam                    (学生+管理员)
//	    ├── GET  /published      已发布考试与历史记录
//	    ├── POST /papers/:id/start  开始答题
//	    └── /attempts/:id/*      答题相关操作
func registerAPIRoutes(r *gin.Engine, deps *dependencies) {
	// 创建 /api 路由组，所有接口都以 /api 开头
	api := r.Group("/api")
	{
		// ---- 公开接口（无需登录） ----
		// 这些接口任何人都可以访问
		api.POST("/auth/login", deps.authHandler.Login)       // 用户登录：验证用户名密码，返回 JWT
		api.POST("/auth/register", deps.authHandler.Register) // 用户注册：创建新用户
		api.POST("/auth/refresh", deps.authHandler.Refresh)   // 刷新 token：旧 token 换新 token

		// ---- 已登录接口（需要有效的 JWT） ----
		// RequireAuth 中间件会验证 JWT 的有效性，并将用户信息写入请求上下文
		protected := api.Group("")
		protected.Use(middleware.RequireAuth(deps.authService, deps.userRepo))
		protected.GET("/auth/me", deps.authHandler.Me)                           // 获取当前登录用户信息
		protected.POST("/auth/logout", deps.authHandler.Logout)                  // 退出登录（前端清除 token）
		protected.POST("/auth/change-password", deps.authHandler.ChangePassword) // 修改密码

		// 课程文档读取：所有已登录用户都可以查看课程和 Markdown 文档
		protected.GET("/documents/courses", deps.documentHandler.ListCourses)
		protected.GET("/documents/courses/:courseId", deps.documentHandler.GetCourse)
		protected.GET("/documents/courses/:courseId/docs/:docId", deps.documentHandler.GetDocument)

		// ---- 管理员接口（需要 admin 角色） ----
		// RequireRoles 中间件会检查当前用户的角色是否在允许列表中
		adminRoutes := protected.Group("")
		adminRoutes.Use(middleware.RequireRoles("admin"))
		adminRoutes.GET("/users", deps.authHandler.ListUsers)         // 用户列表（支持分页、筛选）
		adminRoutes.PUT("/users/:id", deps.authHandler.UpdateUser)    // 更新用户信息（角色、状态、班级）
		adminRoutes.DELETE("/users/:id", deps.authHandler.DeleteUser) // 删除用户

		// ---- 教师/管理员接口 ----
		// 教师和管理员都可以访问题库、试卷、班级管理等功能
		teacherRoutes := protected.Group("")
		teacherRoutes.Use(middleware.RequireRoles("admin", "teacher"))

		// 学习笔记
		teacherRoutes.GET("/notes", deps.noteHandler.Get) // 获取 README.md 内容

		// 题库管理（CRUD + 批量删除）
		teacherRoutes.GET("/questions", deps.questionHandler.List)          // 题目列表
		teacherRoutes.POST("/questions", deps.questionHandler.Create)       // 创建题目
		teacherRoutes.PUT("/questions/:id", deps.questionHandler.Update)    // 更新题目
		teacherRoutes.DELETE("/questions/:id", deps.questionHandler.Delete) // 删除单个题目
		teacherRoutes.DELETE("/questions", deps.questionHandler.DeleteMany) // 批量删除题目

		// AI 出题
		teacherRoutes.POST("/ai/generate", deps.aiHandler.Generate)   // AI 生成题目
		teacherRoutes.POST("/ai/test", deps.aiHandler.TestConnection) // 测试 AI 连接

		// 试卷管理
		teacherRoutes.POST("/papers/generate", deps.paperHandler.Generate)                    // 智能组卷（随机选题）
		teacherRoutes.POST("/papers", deps.paperHandler.Create)                               // 保存试卷
		teacherRoutes.GET("/papers", deps.paperHandler.List)                                  // 试卷列表
		teacherRoutes.GET("/papers/:id", deps.paperHandler.Get)                               // 试卷详情
		teacherRoutes.PUT("/papers/:id", deps.paperHandler.Update)                            // 更新试卷
		teacherRoutes.DELETE("/papers/:id", deps.paperHandler.Delete)                         // 删除试卷
		teacherRoutes.POST("/papers/:id/replace-question", deps.paperHandler.ReplaceQuestion) // 替换试卷中的题目
		teacherRoutes.DELETE("/papers/:id/items/:itemId", deps.paperHandler.DeleteItem)       // 删除试卷中的题目项
		teacherRoutes.POST("/papers/:id/publish", deps.paperHandler.Publish)                  // 发布试卷
		teacherRoutes.POST("/papers/:id/unpublish", deps.paperHandler.Unpublish)              // 取消发布
		teacherRoutes.GET("/papers/:id/submissions", deps.paperHandler.GetSubmissions)        // 查看提交统计

		// 班级管理
		teacherRoutes.GET("/classes", deps.classHandler.List)                                          // 班级列表
		teacherRoutes.POST("/classes", deps.classHandler.Create)                                       // 创建班级
		teacherRoutes.PUT("/classes/:id", deps.classHandler.Update)                                    // 更新班级
		teacherRoutes.DELETE("/classes/:id", deps.classHandler.Delete)                                 // 删除班级
		teacherRoutes.GET("/classes/:id/students", deps.classHandler.ListStudents)                     // 班级学生列表
		teacherRoutes.POST("/classes/:id/students/batch-edit", deps.classHandler.BatchEditStudents)    // 批量加入/移出学生
		teacherRoutes.GET("/classes/:id/students/:studentId/exams", deps.classHandler.GetStudentExams) // 学生考试记录

		// 课程文档管理
		teacherRoutes.POST("/documents/courses", deps.documentHandler.CreateCourse)
		teacherRoutes.PUT("/documents/courses/:courseId", deps.documentHandler.UpdateCourse)
		teacherRoutes.DELETE("/documents/courses/:courseId", deps.documentHandler.DeleteCourse)
		teacherRoutes.POST("/documents/courses/:courseId/docs", deps.documentHandler.CreateDocument)
		teacherRoutes.PUT("/documents/courses/:courseId/docs/:docId", deps.documentHandler.UpdateDocument)
		teacherRoutes.DELETE("/documents/courses/:courseId/docs/:docId", deps.documentHandler.DeleteDocument)

		// ---- 学生/管理员接口 ----
		// 学生可以查看已发布考试与历史记录、答题、查看结果
		studentRoutes := protected.Group("")
		studentRoutes.Use(middleware.RequireRoles("admin", "student"))
		studentRoutes.GET("/exam/published", deps.examHandler.ListPublished)            // 已发布考试与历史记录
		studentRoutes.GET("/exam/papers/:id/detail", deps.examHandler.GetPaperDetail)   // 查看试卷详情
		studentRoutes.POST("/exam/papers/:id/start", deps.examHandler.StartAttempt)     // 开始答题
		studentRoutes.GET("/exam/attempts/:id", deps.examHandler.GetAttempt)            // 获取答题详情
		studentRoutes.PUT("/exam/attempts/:id/answers", deps.examHandler.SaveAnswers)   // 保存答案（自动保存）
		studentRoutes.POST("/exam/attempts/:id/submit", deps.examHandler.SubmitAttempt) // 交卷
		studentRoutes.GET("/exam/attempts/:id/result", deps.examHandler.GetResult)      // 查看考试结果
		studentRoutes.POST("/exam/attempts/:id/events", deps.examHandler.RecordEvent)   // 记录监考事件
	}
}

// corsMiddleware 创建 CORS（跨域资源共享）中间件。
//
// 什么是 CORS？
// 当前端和后端运行在不同的端口（如前端 :3000，后端 :8080）时，
// 浏览器出于安全考虑会阻止跨域请求。CORS 通过设置 HTTP 头来告诉浏览器：
// "这个请求是允许的，不要阻止它"。
//
// 本中间件做了以下事情：
// 1. 设置 Access-Control-Allow-Origin 为请求的 Origin（允许任何来源）
// 2. 允许携带 Cookie（Credentials）
// 3. 允许常见的 HTTP 方法和请求头
// 4. 对 OPTIONS 预检请求直接返回 204
//
// 学习要点：
// - 浏览器的"同源策略"和 CORS 机制
// - OPTIONS 预检请求是什么
// - 为什么开发环境需要 CORS，生产环境通常不需要（同源）
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许任何来源的请求（生产环境应该限制为具体域名）
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		// 允许携带 Cookie
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// 允许的 HTTP 方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

		// 处理 OPTIONS 预检请求
		// 浏览器在发送跨域请求前，会先发送一个 OPTIONS 请求来确认服务器是否允许
		// 这里直接返回 204 No Content，表示"允许"
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 继续处理后续的中间件和路由处理器
		c.Next()
	}
}

// serveStatic 挂载前端静态资源。
//
// 本项目是一个前后端分离的 SPA（单页应用）：
// - 前端使用 React + Vite 构建，产物在 client/dist 目录
// - 后端使用 Gin 提供 API 接口
// - 生产环境下，后端同时负责托管前端静态资源
//
// 静态资源服务策略：
// 1. /assets/* → 直接返回 Vite 构建的 JS/CSS/图片等资源
// 2. /docs/*   → 直接返回文档资源
// 3. /         → 返回 index.html（SPA 入口）
// 4. 其他路径  → 先尝试返回静态文件，找不到则返回 index.html（SPA 路由）
//
// 为什么找不到文件要返回 index.html？
// 因为 SPA 的路由是在前端处理的（如 /papers/1/edit），
// 后端并没有对应的物理文件，但前端的 React Router 会根据 URL 渲染正确的页面。
//
// 参数：
//   - r: Gin 路由引擎
//   - distPath: 前端构建产物目录的绝对路径
func serveStatic(r *gin.Engine, distPath string) {
	// 检查前端构建产物目录是否存在
	if _, err := os.Stat(distPath); err != nil {
		// 如果不存在，说明前端还没有构建，返回提示信息
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "前端尚未构建，请先在 client 目录执行 npm run build"})
		})
		return
	}

	// 挂载 /assets 目录（Vite 构建的 JS、CSS、图片等）
	assetsPath := filepath.Join(distPath, "assets")
	if _, err := os.Stat(assetsPath); err == nil {
		r.Static("/assets", assetsPath)
	}

	// 挂载 /docs 目录（文档资源）
	docsPath := filepath.Join(distPath, "docs")
	if _, err := os.Stat(docsPath); err == nil {
		r.Static("/docs", docsPath)
	}

	// 将根路径 / 指向 index.html
	indexFile := filepath.Join(distPath, "index.html")
	r.StaticFile("/", indexFile)

	// NoRoute 处理所有未匹配的路由
	// 这是 SPA 路由的关键：前端路由如 /papers/1/edit 在后端没有对应的处理函数，
	// 但前端的 React Router 会根据 URL 渲染正确的页面，
	// 所以我们需要返回 index.html 让前端接管路由。
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是 /api/* 路径，返回 404 JSON 错误
		// （API 路径不应该被前端路由接管）
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}

		// 清理请求路径，防止目录遍历攻击（如 ../../etc/passwd）
		requested := filepath.Clean(c.Request.URL.Path)
		target := filepath.Join(distPath, requested)

		// 安全检查：确保目标路径在 distPath 目录内
		if !strings.HasPrefix(target, distPath) {
			// 路径越界，返回 index.html
			c.File(indexFile)
			return
		}

		// 如果请求的是一个真实存在的文件，直接返回
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			c.File(target)
			return
		}

		// 否则返回 index.html，让前端路由接管
		c.File(indexFile)
	})
}
