package chain

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// PendingTxRecord persists PendingTx to disk so it survives a Backend restart.
type PendingTxRecord struct {
	TxID      string    `gorm:"primaryKey;column:tx_id"`
	TxType    string    `gorm:"column:tx_type"`
	Payload   string    `gorm:"column:payload;type:text"`
	Signature string    `gorm:"column:signature;type:text"`
	PublicKey string    `gorm:"column:public_key;type:text"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PendingTxRecord) TableName() string { return "pending_txs" }

var pendingDB *gorm.DB

// InitPendingStore migrates the pending_txs table and reloads any
// transactions that were in flight when the previous process stopped.
func InitPendingStore(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if err := db.AutoMigrate(&PendingTxRecord{}); err != nil {
		return err
	}
	pendingDB = db

	var recs []PendingTxRecord
	if err := db.Order("created_at asc").Find(&recs).Error; err != nil {
		return err
	}
	pendingMu.Lock()
	for _, r := range recs {
		pendingQueue = append(pendingQueue, PendingTx{
			TxID:      r.TxID,
			TxType:    r.TxType,
			Payload:   r.Payload,
			Signature: r.Signature,
			PublicKey: r.PublicKey,
		})
	}
	n := len(pendingQueue)
	pendingMu.Unlock()
	if n > 0 {
		log.Printf("[chain] pending store: reloaded %d tx from disk", n)
	}
	return nil
}

func persistPending(tx PendingTx) error {
	if pendingDB == nil {
		return nil
	}
	return pendingDB.Create(&PendingTxRecord{
		TxID:      tx.TxID,
		TxType:    tx.TxType,
		Payload:   tx.Payload,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
		CreatedAt: time.Now().UTC(),
	}).Error
}

func deletePendingByIDs(ids []string) {
	if pendingDB == nil || len(ids) == 0 {
		return
	}
	if err := pendingDB.Where("tx_id IN ?", ids).Delete(&PendingTxRecord{}).Error; err != nil {
		log.Printf("[chain] pending store: delete failed: %v", err)
	}
}
