package db

import (
	"fmt"
	"log"
	"os"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" { nodeID = "main" }
	dbName := fmt.Sprintf("horizon-core-%s.db", nodeID)
	DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil { log.Fatalf("DB Init Failed: %v", err) }
	err = DB.AutoMigrate(&Block{}, &Transaction{}, &Agent{}, &License{}, &SystemLaw{}, &GlobalBank{}, &GlobalIndustry{})
	if err != nil { log.Fatalf("Migration Failed: %v", err) }
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-001", Title: "Zero-Storage Private Key", Status: "MANDATORY"})
	SeedGlobalEntities()
	log.Printf("✅ Database initialized: %s", dbName)
}
