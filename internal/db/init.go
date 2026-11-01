package db

import (
	"log"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("horizon-core.db"), &gorm.Config{})
	if err != nil { log.Fatalf("DB Init Failed: %v", err) }
	
	err = DB.AutoMigrate(&Block{}, &Transaction{}, &Agent{}, &License{}, &SystemLaw{})
	if err != nil { log.Fatalf("Migration Failed: %v", err) }
	
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-001", Title: "Zero-Storage Private Key", Status: "MANDATORY"})
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-002", Title: "Hardware-Bound License", Status: "MANDATORY"})
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-003", Title: "All Internal Comms Require License", Status: "MANDATORY"})
	
	log.Println("✅ Database initialized.")
}
