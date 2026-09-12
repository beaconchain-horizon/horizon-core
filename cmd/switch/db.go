package main

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// configureSQLite sets high-performance pragmas for SQLite.
// Must be called right after opening the DB.
func configureSQLite(db *gorm.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",       // write-ahead logging
		"PRAGMA synchronous = NORMAL",     // balance speed/safety
		"PRAGMA busy_timeout = 10000",     // wait up to 10s if locked
		"PRAGMA cache_size = -64000",      // 64MB cache
		"PRAGMA temp_store = MEMORY",      // temp tables in RAM
		"PRAGMA mmap_size = 268435456",    // 256MB mmap
		"PRAGMA foreign_keys = OFF",       // faster inserts
		"PRAGMA wal_autocheckpoint = 1000",
	}

	for _, p := range pragmas {
		if err := db.Exec(p).Error; err != nil {
			log.Printf("⚠️  pragma failed: %s (%v)", p, err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)        // SQLite: single writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ SQLite tuned: WAL, single-writer pool")
	return nil
}
