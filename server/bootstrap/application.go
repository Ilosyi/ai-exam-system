package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"week05/homework/server/config"
	"week05/homework/server/database"
	"week05/homework/server/models"
)

type Application struct {
	router *gin.Engine
	addr   string
}

func Start(addr string) error {
	app, err := NewApplication(addr)
	if err != nil {
		return err
	}

	log.Printf("Server listening on %s", app.addr)
	return app.router.Run(app.addr)
}

func NewApplication(addr string) (*Application, error) {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return nil, err
	}

	if addr == "" {
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return nil, err
		}
		addr = ":" + strconv.Itoa(cfg.ServerPort)
	}

	db, err := initDatabase(projectRoot)
	if err != nil {
		return nil, err
	}

	deps, err := initDependencies(projectRoot, db)
	if err != nil {
		return nil, err
	}

	router := buildRouter(projectRoot, deps)
	return &Application{router: router, addr: addr}, nil
}

func resolveProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd failed: %w", err)
	}
	return filepath.Dir(wd), nil
}

func initDatabase(projectRoot string) (*gorm.DB, error) {
	db, err := database.Connect(filepath.Join(projectRoot, "server"), func(db *gorm.DB) error {
		return db.AutoMigrate(
			&models.Question{},
			&models.Paper{},
			&models.PaperItem{},
			&models.PaperPublication{},
			&models.ExamAttempt{},
			&models.ExamAnswer{},
			&models.User{},
			&models.Class{},
			&models.UserClass{},
			&models.ProctorEvent{},
		)
	})
	if err != nil {
		return nil, fmt.Errorf("database init failed: %w", err)
	}
	return db, nil
}
