package main

import (
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openDB(dbPath string) (*gorm.DB, error) {
	if os.Getenv("DB_ENGINE") == "postgres" {
		dsn := os.Getenv("DB_DSN")
		if dsn == "" {
			log.Fatal("SECURITY FATAL: DB_DSN not set")
		}
		log.Println("🐘 PostgreSQL (online mode)")
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(200)
		sqlDB.SetMaxIdleConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return db, nil
	}
	
	// ساخت پوشه data در صورت عدم وجود
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Printf("⚠️  failed to create ./data dir: %v", err)
	}

	if dbPath == "" {
		dbPath = "./data/horizon-switch.db"
	}
	log.Println("💾 SQLite (offline mode):", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := configureSQLite(db); err != nil {
		log.Printf("⚠️ configureSQLite failed: %v", err)
	}
	return db, nil
}
