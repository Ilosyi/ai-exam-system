// ============================================================================
// bootstrap/dependencies.go - 依赖注入与初始化
// ============================================================================
//
// 本文件负责创建和组装所有业务对象（Repository、Service、Handler），
// 这种模式叫做"依赖注入"（Dependency Injection）。
//
// 什么是依赖注入？
// 简单说就是"不自己创建依赖，而是从外部传入"。例如：
// - AuthHandler 不自己创建 UserRepository，而是通过构造函数接收
// - 这样在测试时可以传入 mock 对象，而不必连接真实数据库
//
// 本文件还负责"种子用户"（seed users）的创建：
// - 系统首次启动时，自动创建 admin、teacher01、student01 三个默认账号
// - 如果这些账号已存在，则跳过（幂等操作）
//
// 学习要点：
// - 依赖注入是 Go 工程中最常用的解耦手段
// - context.Background() 是空上下文，用于不需要超时/取消的场景
// - 密码哈希使用 bcrypt 算法，不可逆
// ============================================================================

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"week05/homework/server/handlers"
	"week05/homework/server/models"
	"week05/homework/server/repositories"
	"week05/homework/server/services"
)

// dependencies 是一个内部结构体，持有所有业务层的依赖对象。
//
// 它不在包外暴露，只在 bootstrap 包内部使用。
// 所有 Handler 都通过这个结构体传递给路由注册函数。
//
// 依赖关系图：
//
//	repositories（数据访问层）
//	     ↑
//	 services（业务逻辑层）
//	     ↑
//	 handlers（HTTP 处理层）
//	     ↑
//	   router（路由层）
type dependencies struct {
	authService     *services.AuthService        // 认证服务：JWT 生成/解析、密码哈希
	userRepo        *repositories.UserRepository // 用户数据访问
	questionHandler *handlers.QuestionHandler    // 题库管理接口
	paperHandler    *handlers.PaperHandler       // 试卷管理接口
	examHandler     *handlers.ExamHandler        // 考试接口
	aiHandler       *handlers.AIHandler          // AI 出题接口
	noteHandler     *handlers.NoteHandler        // 学习笔记接口
	authHandler     *handlers.AuthHandler        // 认证接口
	classHandler    *handlers.ClassHandler       // 班级管理接口
	documentHandler *handlers.DocumentHandler    // 课程文档接口
}

// initDependencies 创建并组装所有依赖对象。
//
// 这是依赖注入的"组装工厂"，按照正确的顺序创建所有对象：
// 1. 先创建 Repository（数据访问层）—— 它们只依赖数据库连接
// 2. 再创建 Service（业务逻辑层）—— 它们可能依赖 Repository
// 3. 然后创建种子用户（如果不存在）
// 4. 最后创建 Handler（HTTP 处理层）—— 它们依赖 Repository 和 Service
//
// 参数：
//   - projectRoot: 项目根目录，用于加载配置文件和 .env
//   - db: 数据库连接实例
//
// 返回值：
//   - *dependencies: 包含所有依赖对象的结构体
//   - error: 初始化失败时返回错误
func initDependencies(projectRoot string, db *gorm.DB) (*dependencies, error) {
	// ---- 第一步：创建所有 Repository ----
	// Repository 是数据访问层，封装了所有数据库操作
	// 每个 Repository 对应一个业务实体（题目、试卷、考试等）
	questionRepo := repositories.NewQuestionRepository(db)                                        // 题库数据访问
	paperRepo := repositories.NewPaperRepository(db)                                              // 试卷数据访问
	examRepo := repositories.NewExamRepository(db)                                                // 考试数据访问
	userRepo := repositories.NewUserRepository(db)                                                // 用户数据访问
	classRepo := repositories.NewClassRepository(db)                                              // 班级数据访问
	documentRepo := repositories.NewDocumentRepository(filepath.Join(projectRoot, "course-docs")) // 课程文档文件仓库

	// ---- 第二步：创建 Service ----
	// Service 是业务逻辑层，处理跨 Repository 的复杂业务
	aiService, err := services.NewAIService(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("ai service init failed: %w", err)
	}

	// AuthService 处理 JWT 令牌和密码哈希
	// 参数说明：
	// - JWT_SECRET 环境变量：用于签名 JWT 的密钥
	// - 24*time.Hour：JWT 有效期为 24 小时
	authService := services.NewAuthService(os.Getenv("JWT_SECRET"), 24*time.Hour)
	documentService := services.NewDocumentService(documentRepo)

	// ---- 第三步：创建种子用户 ----
	// 系统首次启动时，自动创建默认的 admin、teacher、student 账号
	// 如果账号已存在，则跳过（幂等操作）
	if err := seedDefaultUsers(context.Background(), userRepo, authService); err != nil {
		return nil, fmt.Errorf("seed users failed: %w", err)
	}

	// ---- 第四步：创建所有 Handler ----
	// Handler 是 HTTP 处理层，负责解析请求、校验参数、调用 Service/Repository、返回响应
	// 每个 Handler 通过构造函数接收它所依赖的 Repository 或 Service
	return &dependencies{
		authService:     authService,
		userRepo:        userRepo,
		questionHandler: handlers.NewQuestionHandler(questionRepo, userRepo),                   // 题库 Handler 依赖题库和用户 Repository
		paperHandler:    handlers.NewPaperHandler(paperRepo, questionRepo, classRepo),          // 试卷 Handler 依赖试卷、题目、班级 Repository
		examHandler:     handlers.NewExamHandler(examRepo, paperRepo, questionRepo, classRepo), // 考试 Handler 依赖四个 Repository
		aiHandler:       handlers.NewAIHandler(aiService),                                      // AI Handler 依赖 AI Service
		noteHandler:     handlers.NewNoteHandler(projectRoot),                                  // 笔记 Handler 依赖项目根目录（读取 README）
		authHandler:     handlers.NewAuthHandler(userRepo, authService),                        // 认证 Handler 依赖用户 Repository 和认证 Service
		classHandler:    handlers.NewClassHandler(classRepo),                                   // 班级 Handler 依赖班级 Repository
		documentHandler: handlers.NewDocumentHandler(documentService),                          // 文档 Handler 依赖文档 Service
	}, nil
}

// seedDefaultUsers 创建系统默认用户（如果不存在）。
//
// 这是一个"幂等操作"——无论调用多少次，结果都是一样的：
// - 如果用户已存在，跳过
// - 如果用户不存在，创建
//
// 默认账号：
//   - admin / admin123     (角色：管理员)
//   - teacher01 / teacher123 (角色：教师)
//   - student01 / student123 (角色：学生)
//
// 安全提示：这些是开发/演示用的默认密码，生产环境应该修改。
//
// 参数：
//   - ctx: 上下文，用于控制超时/取消（这里用 Background 表示不需要控制）
//   - userRepo: 用户数据访问层
//   - authService: 认证服务，用于哈希密码
func seedDefaultUsers(ctx context.Context, userRepo *repositories.UserRepository, authService *services.AuthService) error {
	// 使用 bcrypt 算法对密码进行哈希
	// bcrypt 是一种单向哈希算法，无法从哈希值反推出原始密码
	// 即使数据库泄露，攻击者也无法获取用户的真实密码
	adminHash, err := authService.HashPassword("admin123")
	if err != nil {
		return err
	}
	teacherHash, err := authService.HashPassword("teacher123")
	if err != nil {
		return err
	}
	studentHash, err := authService.HashPassword("student123")
	if err != nil {
		return err
	}

	// EnsureDefaults 方法会检查每个用户是否已存在：
	// - 如果 username 已存在，跳过
	// - 如果 username 不存在，插入新记录
	return userRepo.EnsureDefaults(ctx, []models.User{
		{Username: "admin", Role: "admin", PasswordHash: adminHash, Status: "active"},
		{Username: "teacher01", Role: "teacher", PasswordHash: teacherHash, Status: "active"},
		{Username: "student01", Role: "student", PasswordHash: studentHash, Status: "active"},
	})
}
