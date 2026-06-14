// ============================================================================
// database/database.go - SQLite 数据库连接管理
// ============================================================================
//
// 本文件负责初始化 SQLite 数据库连接，并执行表结构迁移。
//
// 为什么用 SQLite？
// - 零配置：不需要安装数据库服务器，数据存储在单个文件中
// - 适合教学/演示：开箱即用，减少环境配置的复杂度
// - 轻量级：单文件数据库，便于备份和迁移
//
// 设计要点：
// - 使用 sync.Once 保证全局只有一个数据库连接（单例模式）
// - 自动创建 data 目录和数据库文件
// - 支持自定义迁移函数，在连接成功后自动执行
//
// 学习要点：
// - sync.Once 保证函数只执行一次（并发安全）
// - GORM 的基本用法：Open、AutoMigrate
// - SQLite 的 DSN（数据源名称）格式
// ============================================================================

package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite" // SQLite 驱动
	"gorm.io/gorm"          // GORM ORM 库
)

// 包级变量，用于实现单例模式
var (
	db   *gorm.DB   // 全局数据库连接实例
	once sync.Once  // 保证 Connect 函数只执行一次
)

// Connect 初始化 SQLite 数据库连接。
//
// 这是一个单例函数：无论调用多少次，都只创建一个数据库连接。
// 使用 sync.Once 保证并发安全——即使多个 goroutine 同时调用，
// 也只会有一个 goroutine 执行初始化逻辑。
//
// 参数：
//   - baseDir: 基础目录，通常是 server/ 目录的绝对路径
//   - migrate: 迁移函数，在数据库连接成功后执行，用于创建/更新表结构
//
// 返回值：
//   - *gorm.DB: 数据库连接实例
//   - error: 初始化失败时返回错误
//
// 工作流程：
// 1. 创建 data 目录（如果不存在）
// 2. 打开 SQLite 数据库文件（如果文件不存在会自动创建）
// 3. 执行迁移函数（创建/更新表结构）
//
// 数据库文件位置：<baseDir>/data/questions.db
func Connect(baseDir string, migrate func(db *gorm.DB) error) (*gorm.DB, error) {
	var err error

	// sync.Once 保证以下函数体只执行一次
	// 即使多个 goroutine 并发调用 Connect，也只会有一个执行初始化
	once.Do(func() {
		// 第一步：创建 data 目录
		// os.MkdirAll 会递归创建目录，如果目录已存在则不报错
		dataDir := filepath.Join(baseDir, "data")
		if mkErr := os.MkdirAll(dataDir, 0o755); mkErr != nil {
			err = fmt.Errorf("create data dir failed: %w", mkErr)
			return
		}

		// 第二步：打开 SQLite 数据库
		// DSN（数据源名称）就是数据库文件的路径
		// 如果文件不存在，SQLite 会自动创建一个新的空数据库
		dsn := filepath.Join(dataDir, "questions.db")
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			err = fmt.Errorf("open sqlite failed: %w", err)
			return
		}

		// 第三步：执行迁移函数
		// 迁移函数通常调用 db.AutoMigrate() 来创建/更新表结构
		if migrate != nil {
			if mgErr := migrate(db); mgErr != nil {
				err = fmt.Errorf("auto migrate failed: %w", mgErr)
			}
		}
	})

	return db, err
}
