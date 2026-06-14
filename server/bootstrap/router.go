package bootstrap

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"week05/homework/server/middleware"
)

func buildRouter(projectRoot string, deps *dependencies) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	registerAPIRoutes(r, deps)
	serveStatic(r, filepath.Join(projectRoot, "client", "dist"))

	return r
}

func registerAPIRoutes(r *gin.Engine, deps *dependencies) {
	api := r.Group("/api")
	{
		api.POST("/auth/login", deps.authHandler.Login)
		api.POST("/auth/register", deps.authHandler.Register)
		api.POST("/auth/refresh", deps.authHandler.Refresh)

		protected := api.Group("")
		protected.Use(middleware.RequireAuth(deps.authService, deps.userRepo))
		protected.GET("/auth/me", deps.authHandler.Me)
		protected.POST("/auth/logout", deps.authHandler.Logout)
		protected.POST("/auth/change-password", deps.authHandler.ChangePassword)
		protected.GET("/documents/courses", deps.documentHandler.ListCourses)
		protected.GET("/documents/courses/:courseId", deps.documentHandler.GetCourse)
		protected.GET("/documents/courses/:courseId/docs/:docId", deps.documentHandler.GetDocument)

		adminRoutes := protected.Group("")
		adminRoutes.Use(middleware.RequireRoles("admin"))
		adminRoutes.GET("/users", deps.authHandler.ListUsers)
		adminRoutes.PUT("/users/:id", deps.authHandler.UpdateUser)
		adminRoutes.DELETE("/users/:id", deps.authHandler.DeleteUser)

		teacherRoutes := protected.Group("")
		teacherRoutes.Use(middleware.RequireRoles("admin", "teacher"))
		teacherRoutes.GET("/notes", deps.noteHandler.Get)
		teacherRoutes.GET("/questions", deps.questionHandler.List)
		teacherRoutes.POST("/questions", deps.questionHandler.Create)
		teacherRoutes.PUT("/questions/:id", deps.questionHandler.Update)
		teacherRoutes.DELETE("/questions/:id", deps.questionHandler.Delete)
		teacherRoutes.DELETE("/questions", deps.questionHandler.DeleteMany)
		teacherRoutes.POST("/ai/generate", deps.aiHandler.Generate)
		teacherRoutes.POST("/ai/test", deps.aiHandler.TestConnection)
		teacherRoutes.POST("/papers/generate", deps.paperHandler.Generate)
		teacherRoutes.POST("/papers", deps.paperHandler.Create)
		teacherRoutes.GET("/papers", deps.paperHandler.List)
		teacherRoutes.GET("/papers/:id", deps.paperHandler.Get)
		teacherRoutes.PUT("/papers/:id", deps.paperHandler.Update)
		teacherRoutes.DELETE("/papers/:id", deps.paperHandler.Delete)
		teacherRoutes.POST("/papers/:id/replace-question", deps.paperHandler.ReplaceQuestion)
		teacherRoutes.DELETE("/papers/:id/items/:itemId", deps.paperHandler.DeleteItem)
		teacherRoutes.POST("/papers/:id/publish", deps.paperHandler.Publish)
		teacherRoutes.POST("/papers/:id/unpublish", deps.paperHandler.Unpublish)
		teacherRoutes.GET("/papers/:id/submissions", deps.paperHandler.GetSubmissions)
		teacherRoutes.GET("/classes", deps.classHandler.List)
		teacherRoutes.POST("/classes", deps.classHandler.Create)
		teacherRoutes.PUT("/classes/:id", deps.classHandler.Update)
		teacherRoutes.DELETE("/classes/:id", deps.classHandler.Delete)
		teacherRoutes.GET("/classes/:id/students", deps.classHandler.ListStudents)
		teacherRoutes.POST("/classes/:id/students/batch-edit", deps.classHandler.BatchEditStudents)
		teacherRoutes.GET("/classes/:id/students/:studentId/exams", deps.classHandler.GetStudentExams)
		teacherRoutes.POST("/documents/courses", deps.documentHandler.CreateCourse)
		teacherRoutes.PUT("/documents/courses/:courseId", deps.documentHandler.UpdateCourse)
		teacherRoutes.DELETE("/documents/courses/:courseId", deps.documentHandler.DeleteCourse)
		teacherRoutes.POST("/documents/courses/:courseId/docs", deps.documentHandler.CreateDocument)
		teacherRoutes.PUT("/documents/courses/:courseId/docs/:docId", deps.documentHandler.UpdateDocument)
		teacherRoutes.DELETE("/documents/courses/:courseId/docs/:docId", deps.documentHandler.DeleteDocument)

		studentRoutes := protected.Group("")
		studentRoutes.Use(middleware.RequireRoles("admin", "student"))
		studentRoutes.GET("/exam/published", deps.examHandler.ListPublished)
		studentRoutes.POST("/exam/papers/:id/start", deps.examHandler.StartAttempt)
		studentRoutes.GET("/exam/attempts/:id", deps.examHandler.GetAttempt)
		studentRoutes.PUT("/exam/attempts/:id/answers", deps.examHandler.SaveAnswers)
		studentRoutes.POST("/exam/attempts/:id/submit", deps.examHandler.SubmitAttempt)
		studentRoutes.GET("/exam/attempts/:id/result", deps.examHandler.GetResult)
		studentRoutes.POST("/exam/attempts/:id/events", deps.examHandler.RecordEvent)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func serveStatic(r *gin.Engine, distPath string) {
	if _, err := os.Stat(distPath); err != nil {
		r.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "前端尚未构建，请先在 client 目录执行 npm run build"})
		})
		return
	}

	assetsPath := filepath.Join(distPath, "assets")
	if _, err := os.Stat(assetsPath); err == nil {
		r.Static("/assets", assetsPath)
	}

	docsPath := filepath.Join(distPath, "docs")
	if _, err := os.Stat(docsPath); err == nil {
		r.Static("/docs", docsPath)
	}

	indexFile := filepath.Join(distPath, "index.html")
	r.StaticFile("/", indexFile)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.URL.Path == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}

		requested := filepath.Clean(c.Request.URL.Path)
		target := filepath.Join(distPath, requested)
		if !strings.HasPrefix(target, distPath) {
			c.File(indexFile)
			return
		}
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			c.File(target)
			return
		}
		c.File(indexFile)
	})
}
