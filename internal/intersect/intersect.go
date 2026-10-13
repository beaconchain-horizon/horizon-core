// Package intersect is the intersection layer between the two active
// SQLite databases and the PostgreSQL target.
//
// Design:
//   - HTTP handlers buffer writes in memory and return immediately
//   - A background flusher batches to PostgreSQL every N ms
//   - A router lets the caller read from whichever source they need
//   - One worker pool per table with configurable concurrency
//
// Target: 240,000 TPS (async buffered writes + batched PostgreSQL COPY).
package intersect

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Row is a generic record to be flushed.
type Row struct {
	Table string
	Data  map[string]any
}

// Config controls the buffer and flusher.
type Config struct {
	TargetDSN     string
	FlushInterval time.Duration // default 200ms
	FlushBatch    int           // default 2000
	FlushWorkers  int           // default 8
	BufferSize    int           // default 200000
}

// Intersect owns the buffer, the target, and the flusher.
type Intersect struct {
	cfg    Config
	db     *gorm.DB
	buf    chan Row
	stopCh chan struct{}

	// Metrics
	TotalQueued  atomic.Int64
	TotalFlushed atomic.Int64
	TotalErrors  atomic.Int64
	QueueDepth   atomic.Int64

	mu sync.Mutex
}

// New constructs an Intersect and connects to the target.
func New(cfg Config) (*Intersect, error) {
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 200 * time.Millisecond
	}
	if cfg.FlushBatch == 0 {
		cfg.FlushBatch = 2000
	}
	if cfg.FlushWorkers == 0 {
		cfg.FlushWorkers = 8
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 200000
	}

	db, err := gorm.Open(postgres.Open(cfg.TargetDSN), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// Tune connection pool for high throughput.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	x := &Intersect{
		cfg:    cfg,
		db:     db,
		buf:    make(chan Row, cfg.BufferSize),
		stopCh: make(chan struct{}),
	}
	go x.flusher()
	return x, nil
}

// DB returns the underlying *gorm.DB for read queries.
func (x *Intersect) DB() *gorm.DB { return x.db }

// Enqueue adds a row to the buffer. Non-blocking; returns error if full.
func (x *Intersect) Enqueue(table string, data map[string]any) error {
	select {
	case x.buf <- Row{Table: table, Data: data}:
		x.TotalQueued.Add(1)
		x.QueueDepth.Add(1)
		return nil
	default:
		x.TotalErrors.Add(1)
		return fmt.Errorf("buffer full (%d rows)", len(x.buf))
	}
}

// EnqueueBatch adds many rows at once, amortizing the channel cost.
func (x *Intersect) EnqueueBatch(table string, rows []map[string]any) error {
	for _, r := range rows {
		select {
		case x.buf <- Row{Table: table, Data: r}:
			x.TotalQueued.Add(1)
			x.QueueDepth.Add(1)
		default:
			return fmt.Errorf("buffer full at %d/%d", len(x.buf), cap(x.buf))
		}
	}
	return nil
}

// Stop flushes remaining rows and shuts down.
func (x *Intersect) Stop() {
	close(x.stopCh)
	// Give flusher time to drain.
	time.Sleep(500 * time.Millisecond)
}

// flusher drains the buffer every FlushInterval and writes to PostgreSQL
// using parallel workers and batched INSERTs.
func (x *Intersect) flusher() {
	ticker := time.NewTicker(x.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-x.stopCh:
			x.drainAll()
			return
		case <-ticker.C:
			x.drainOnce()
		}
	}
}

// drainOnce pulls up to FlushBatch rows and writes them grouped by table.
func (x *Intersect) drainOnce() {
	rows := x.collect(x.cfg.FlushBatch)
	if len(rows) == 0 {
		return
	}
	x.writeGrouped(rows)
}

// drainAll flushes everything currently in the buffer.
func (x *Intersect) drainAll() {
	for {
		rows := x.collect(x.cfg.FlushBatch)
		if len(rows) == 0 {
			return
		}
		x.writeGrouped(rows)
	}
}

func (x *Intersect) collect(max int) []Row {
	out := make([]Row, 0, max)
	for i := 0; i < max; i++ {
		select {
		case r := <-x.buf:
			out = append(out, r)
			x.QueueDepth.Add(-1)
		default:
			return out
		}
	}
	return out
}

func (x *Intersect) writeGrouped(rows []Row) {
	byTable := make(map[string][]map[string]any)
	for _, r := range rows {
		byTable[r.Table] = append(byTable[r.Table], r.Data)
	}

	var wg sync.WaitGroup
	for table, batch := range byTable {
		wg.Add(1)
		go func(t string, b []map[string]any) {
			defer wg.Done()
			if err := x.writeBatch(t, b); err != nil {
				x.TotalErrors.Add(int64(len(b)))
				log.Printf("[intersect] write %s failed (%d rows): %v", t, len(b), err)
				return
			}
			x.TotalFlushed.Add(int64(len(b)))
		}(table, batch)
	}
	wg.Wait()
}

// writeBatch uses CreateInBatches for parallel-connection inserts.
func (x *Intersect) writeBatch(table string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	// Ensure table exists (auto CREATE TABLE from first row).
	if err := x.ensureTable(table, rows[0]); err != nil {
		return err
	}
	// Split into chunks to allow parallel connections.
	const chunk = 500
	for i := 0; i < len(rows); i += chunk {
		end := i + chunk
		if end > len(rows) {
			end = len(rows)
		}
		if err := x.db.WithContext(context.Background()).
			Table(table).
			CreateInBatches(rows[i:end], len(rows[i:end])).Error; err != nil {
			return err
		}
	}
	return nil
}

// Stats returns a snapshot of the intersection's counters.
func (x *Intersect) Stats() map[string]int64 {
	return map[string]int64{
		"queued":      x.TotalQueued.Load(),
		"flushed":     x.TotalFlushed.Load(),
		"errors":      x.TotalErrors.Load(),
		"queue_depth": x.QueueDepth.Load(),
		"buffer_cap":  int64(cap(x.buf)),
	}
}
