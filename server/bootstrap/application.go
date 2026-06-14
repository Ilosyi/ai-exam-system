// ============================================================================
// bootstrap/application.go - 应用启动与初始化
// ============================================================================
//
// 本文件是整个后端的"启动引擎"，负责将所有零件组装在一起并启动服务。
//
// 职责：
// 1. 解析项目根目录（用于定位配置文件、数据库文件、静态资源等）
// 2. 加载配置文件 (config.json)
// 3. 初始化 SQLite 数据库连接并执行表结构迁移
// 4. 创建所有依赖对象（Repository、Service、Handler）
// 5. 构建 HTTP 路由树
// 6. 启动 HTTP 服务器
//
// 设计模式说明：
// - Application 结构体封装了路由器和监听地址，是整个应用的"外壳"
// - Start() 函数是对外暴露的唯一入口，内部编排了完整的启动流程
// - 每个 initXxx 函数只负责一个关注点，便于测试和理解
//
// 学习要点：
// - Go 的错误包装：fmt.Errorf("xxx failed: %w", err) 保留原始错误链
// - filepath 的跨平台路径拼接
// - GORM AutoMigrate 自动建表/加字段
// ============================================================================

package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin" // Gin 是 Go 最流行的 HTTP 框架，类似 Express.js
	"gorm.io/gorm"             // GORM 是 Go 最流行的 ORM 库，类似 SQLAlchemy

	"week05/homework/server/config"
	"week05/homework/server/database"
	"week05/homework/server/models"
)

// Application 是应用的核心结构体，持有 Gin 路由引擎和服务器监听地址。
// 通过将 router 和 addr 封装在一起，我们可以：
// 1. 在测试中创建 Application 实例而不实际启动服务器
// 2. 方便地获取应用状态
type Application struct {
	router *gin.Engine // Gin 路由引擎，处理所有 HTTP 请求的分发
	addr   string      // 服务器监听地址，格式为 ":端口号"，如 ":8080"
}

// Start 是应用的启动入口函数。
//
// 参数：
//   - addr: 服务器监听地址，传空字符串 "" 表示使用配置文件中的端口
//
// 返回值：
//   - error: 如果启动失败（如端口被占用），返回错误
//
// 工作流程：
// 1. 调用 NewApplication 创建应用实例（包括所有初始化）
// 2. 打印启动日志
// 3. 调用 router.Run() 开始监听 HTTP 请求（这是一个阻塞调用）
func Start(addr string) error {
	// 创建应用实例，内部完成所有初始化工作
	app, err := NewApplication(addr)
	if err != nil {
		return err
	}

	// 打印服务器启动信息，方便调试
	log.Printf("Server listening on %s", app.addr)

	// router.Run() 会阻塞当前 goroutine，持续监听并处理 HTTP 请求
	// 只有在发生致命错误时才会返回
	return app.router.Run(app.addr)
}

// NewApplication 创建并初始化一个完整的 Application 实例。
//
// 这是整个启动流程的核心函数，按顺序完成以下步骤：
// 1. 解析项目根目录
// 2. 加载配置文件
// 3. 初始化数据库
// 4. 创建所有依赖（Repository、Service、Handler）
// 5. 构建路由树
//
// 将这个函数独立出来（而不是放在 Start 里）的好处是：
// - 测试时可以创建 Application 实例而不启动服务器
// - 可以在启动前检查应用状态
func NewApplication(addr string) (*Application, error) {
	// 第一步：确定项目根目录
	// 所有配置文件、数据库文件、静态资源都相对于这个目录
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return nil, err
	}

	// 第二步：如果调用者没有指定监听地址，则从配置文件读取
	if addr == "" {
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return nil, err
		}
		// 将端口号转换为 ":端口" 格式，如 ":8080"
		addr = ":" + strconv.Itoa(cfg.ServerPort)
	}

	// 第三步：初始化数据库连接并执行表结构迁移
	db, err := initDatabase(projectRoot)
	if err != nil {
		return nil, err
	}

	// 第四步：创建所有依赖对象（依赖注入）
	deps, err := initDependencies(projectRoot, db)
	if err != nil {
		return nil, err
	}

	// 第五步：构建路由树
	router := buildRouter(projectRoot, deps)

	// 返回组装好的应用实例
	return &Application{router: router, addr: addr}, nil
}

// resolveProjectRoot 解析项目根目录的绝对路径。
//
// 当前实现是取"当前工作目录的父目录"，因为服务器程序通常在 server/ 目录下运行，
// 所以父目录就是项目根目录（week05/homework/）。
//
// 返回值示例："/Users/xxx/CodeHub/kejunxiang/week05/homework"
//
// 学习要点：
// - os.Getwd() 获取当前工作目录
// - filepath.Dir() 获取父目录路径
// - Go 的错误包装使用 fmt.Errorf + %w 动词
func resolveProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd failed: %w", err)
	}
	// filepath.Dir 返回路径的父目录
	// 例如："/Users/xxx/week05/homework/server" → "/Users/xxx/week05/homework"
	return filepath.Dir(wd), nil
}

// initDatabase 初始化 SQLite 数据库连接并执行自动迁移。
//
// 参数：
//   - projectRoot: 项目根目录，数据库文件将存储在 <projectRoot>/server/data/questions.db
//
// 返回值：
//   - *gorm.DB: 数据库连接实例，后续所有数据库操作都通过这个对象
//   - error: 初始化失败时返回错误
//
// AutoMigrate 会自动完成以下工作：
// - 如果表不存在，创建表
// - 如果表缺少字段，添加字段
// - 如果字段类型变化，尝试修改（但不会删除字段）
//
// 注意：AutoMigrate 不会删除已有的字段或表，这是一种安全的迁移策略。
func initDatabase(projectRoot string) (*gorm.DB, error) {
	// database.Connect 内部使用 sync.Once 保证全局只有一个数据库连接
	// 第二个参数是迁移函数，会在数据库连接成功后自动执行
	db, err := database.Connect(filepath.Join(projectRoot, "server"), func(db *gorm.DB) error {
		// AutoMigrate 传入所有需要迁移的模型
		// GORM 会根据模型的 gorm 标签自动创建/更新对应的数据库表
		return db.AutoMigrate(
			&models.Question{},          // 题库题目表
			&models.Paper{},             // 试卷表
			&models.PaperItem{},         // 试卷题目项表
			&models.PaperPublication{},  // 试卷发布记录表
			&models.ExamAttempt{},       // 考试答题记录表
			&models.ExamAnswer{},        // 考试答案表
			&models.User{},              // 用户表
			&models.Class{},             // 班级表
			&models.UserClass{},         // 学生-班级关联表
			&models.ProctorEvent{},      // 监考事件表
		)
	})
	if err != nil {
		return nil, fmt.Errorf("database init failed: %w", err)
	}
	return db, nil
}
