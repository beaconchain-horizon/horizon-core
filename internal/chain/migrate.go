package chain

import (
	"gorm.io/gorm"
	"log"
)

func AutoMigrateAll(db *gorm.DB) error {
	log.Println("[chain] migrating Block")
	if err := db.AutoMigrate(&Block{}); err != nil {
		return err
	}
	log.Println("[chain] migrating Transaction")
	if err := db.AutoMigrate(&Transaction{}); err != nil {
		return err
	}
	log.Println("[chain] migrating MerkleRoot")
	if err := db.AutoMigrate(&MerkleRoot{}); err != nil {
		return err
	}
	log.Println("[chain] migrating Genesis")
	if err := db.AutoMigrate(&Genesis{}); err != nil {
		return err
	}
	return nil
}
