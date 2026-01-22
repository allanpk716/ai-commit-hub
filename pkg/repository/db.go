package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/allanpk716/ai-commit-hub/pkg/models"
)

var (
	db   *gorm.DB
	once sync.Once
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string
}

// NewDatabaseConfig creates a new database config
func NewDatabaseConfig() *DatabaseConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}

	configDir := filepath.Join(homeDir, ".ai-commit-hub")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create config directory: %v", err))
	}

	return &DatabaseConfig{
		Path: filepath.Join(configDir, "ai-commit-hub.db"),
	}
}

// InitializeDatabase initializes the database connection
func InitializeDatabase(config *DatabaseConfig) error {
	var initErr error
	once.Do(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(config.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Auto migrate schemas
		if err := db.AutoMigrate(&models.GitProject{}); err != nil {
			initErr = fmt.Errorf("failed to migrate database: %w", err)
			return
		}

		fmt.Println("Database initialized:", config.Path)
	})

	return initErr
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
