package main

import (
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var txQueue chan *Transaction

func initTxQueue() {
	txQueue = make(chan *Transaction, 500000)
}

func queueTransaction(tx *Transaction) {
	txQueue <- tx
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

// flushLedger — OPTIMIZED: single SELECT + bulk insert
func flushLedger(l *Ledger) {
	snap := l.Snapshot()
	if len(snap) == 0 {
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		ids := make([]string, 0, len(snap))
		for id := range snap {
			ids = append(ids, id)
		}

		var existing []Account
		if err := tx.Where("bank_id IN ?", ids).Find(&existing).Error; err != nil {
			return err
		}

		existMap := make(map[string]float64, len(existing))
		for _, a := range existing {
			existMap[a.BankID] = a.Balance
		}

		var toInsert []Account
		var toUpdate []Account
		for id, bal := range snap {
			if _, ok := existMap[id]; !ok {
				toInsert = append(toInsert, Account{BankID: id, Balance: bal})
			} else if existMap[id] != bal {
				toUpdate = append(toUpdate, Account{BankID: id, Balance: bal})
			}
		}

		if len(toInsert) > 0 {
			if err := tx.CreateInBatches(toInsert, 500).Error; err != nil {
				return err
			}
		}

		for _, a := range toUpdate {
			if err := tx.Model(&Account{}).Where("bank_id = ?", a.BankID).
				Update("balance", a.Balance).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("ledger flush error: %v", err)
	}
}

func txWriterLoop(stopCh chan struct{}) {
	batch := make([]*Transaction, 0, 5000)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(batch, 1000).Error; err != nil {
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
			if len(batch) >= 5000 {
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
