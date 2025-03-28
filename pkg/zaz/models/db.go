package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver that doesn't require CGO
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection that can be used by all models
var DB *gorm.DB
var dbOnce sync.Once
var dbMutex sync.RWMutex

// InitDB initializes the database connection
// It uses SQLite for now but can be changed to PostgreSQL in the future
func InitDB(dbPath string) (*gorm.DB, error) {
	var err error
	
	dbOnce.Do(func() {
		// Create directory if it doesn't exist
		dir := filepath.Dir(dbPath)
		if err = os.MkdirAll(dir, 0755); err != nil {
			err = fmt.Errorf("failed to create database directory: %w", err)
			return
		}

		// Configure GORM logger
		gormLogger := logger.Default.LogMode(logger.Info)

		// Open database connection
		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			err = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// Initialize tables by auto-migrating the models
		if err = autoMigrate(); err != nil {
			err = fmt.Errorf("failed to migrate tables: %w", err)
			return
		}

		log.Println("Database initialized successfully")
	})

	return DB, err
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	return DB
}

// autoMigrate automatically creates and updates database tables based on model structs
func autoMigrate() error {
	// Auto migrate all models
	return DB.AutoMigrate(
		&User{},
		&Company{},
		&Shareholder{},
		&BoardMeeting{},
		&Attendee{},
		&Vote{},
		&VoteOption{},
		&Ballot{},
		&Product{},
		&Sale{},
		&SaleItem{},
	)
}




