package main

import (
	"fmt"
	"log"

	"bazi/internal/config"
	"bazi/migrations"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initDatabase opens a GORM database connection based on config.
// Uses SQLite when DB_HOST is empty, MySQL otherwise.
func initDatabase(cfg *config.Config) *gorm.DB {
	var dialector gorm.Dialector

	if cfg.UseSQLite {
		log.Printf("Using SQLite database: %s", cfg.SQLitePath)
		dialector = sqlite.Open(cfg.SQLitePath)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		log.Printf("Using MySQL database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := migrations.Apply(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
