package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Connect initializes a SQLite database (stored under server/data/questions.db).
func Connect(baseDir string, migrate func(db *gorm.DB) error) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		dataDir := filepath.Join(baseDir, "data")
		if mkErr := os.MkdirAll(dataDir, 0o755); mkErr != nil {
			err = fmt.Errorf("create data dir failed: %w", mkErr)
			return
		}
		dsn := filepath.Join(dataDir, "questions.db")
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			err = fmt.Errorf("open sqlite failed: %w", err)
			return
		}
		if migrate != nil {
			if mgErr := migrate(db); mgErr != nil {
				err = fmt.Errorf("auto migrate failed: %w", mgErr)
			}
		}
	})
	return db, err
}
