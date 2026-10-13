package chain

import "gorm.io/gorm"

// AutoMigrateAll creates or updates all chain-related tables.
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		&Block{},
		&Transaction{},
		&MerkleRoot{},
		&Genesis{},
	)
}
