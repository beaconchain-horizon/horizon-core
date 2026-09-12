package main

import (
	"log"
	"time"

	"gorm.io/gorm"
)

var txQueue chan *Transaction

func initTxQueue() {
	txQueue = make(chan *Transaction, 500000)
}

func queueTransaction(tx *Transaction) {
	select {
	case txQueue <- tx:
	default:
		// Queue full - write synchronously as fallback
		if err := db.Create(tx).Error; err != nil {
			log.Printf("sync tx write error: %v", err)
		}
	}
}

func startLedgerLoops(l *Ledger, stopCh chan struct{}) {
	go balanceFlushLoop(l, stopCh)
	go txWriterLoop(stopCh)
}

func balanceFlushLoop(l *Ledger, stopCh chan struct{}) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var lastOps int64
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			ops := l.Ops()
			if ops == lastOps {
				continue
			}
			flushLedger(l)
			lastOps = ops
		}
	}
}

func flushLedger(l *Ledger) {
	snap := l.Snapshot()
	if len(snap) == 0 {
		return
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for bankID, bal := range snap {
			var acc Account
			err := tx.Where("bank_id = ?", bankID).First(&acc).Error
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&Account{BankID: bankID, Balance: bal}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if acc.Balance != bal {
				if err := tx.Model(&Account{}).Where("bank_id = ?", bankID).
					Update("balance", bal).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("ledger flush error: %v", err)
	}
}

func txWriterLoop(stopCh chan struct{}) {
	batch := make([]*Transaction, 0, 1000)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := db.CreateInBatches(batch, 500).Error; err != nil {
			log.Printf("tx batch write error: %v", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-stopCh:
			flush()
			return
		case tx := <-txQueue:
			batch = append(batch, tx)
			if len(batch) >= 1000 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func loadLedgerFromDB(l *Ledger) error {
	var accounts []Account
	if err := db.Find(&accounts).Error; err != nil {
		return err
	}
	m := make(map[string]float64, len(accounts))
	for _, a := range accounts {
		m[a.BankID] = a.Balance
	}
	l.LoadSnapshot(m)
	log.Printf("ledger: loaded %d accounts from DB", len(m))
	return nil
}
