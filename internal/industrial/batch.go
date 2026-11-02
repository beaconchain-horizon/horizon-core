package industrial

import (
	"gorm.io/gorm"
)

func InsertReadingsBulk(db *gorm.DB, readings []Reading) error {
	if len(readings) == 0 {
		return nil
	}
	return db.CreateInBatches(readings, len(readings)).Error
}
