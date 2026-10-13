// Package migration moves data from SQLite sources to a PostgreSQL target.
package migration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Job struct {
	ID         string   `json:"id"`
	SourcePath string   `json:"source_path"`
	TargetDSN  string   `json:"target_dsn"`
	Tables     []string `json:"tables"`
	BatchSize  int      `json:"batch_size"`
	Workers    int      `json:"workers"`
	Status     string   `json:"status"`

	TotalRead    int64      `json:"total_read"`
	TotalWritten int64      `json:"total_written"`
	Errors       int64      `json:"errors"`
	LastError    string     `json:"last_error,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type Manager struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func NewManager() *Manager {
	return &Manager{jobs: make(map[string]*Job)}
}

func (m *Manager) Start(j *Job) error {
	m.mu.Lock()
	if _, exists := m.jobs[j.ID]; exists {
		m.mu.Unlock()
		return fmt.Errorf("job %s already exists", j.ID)
	}
	m.jobs[j.ID] = j
	m.mu.Unlock()
	go m.run(j)
	return nil
}

func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	return j, ok
}

func (m *Manager) List() []*Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, j)
	}
	return out
}

func (m *Manager) run(j *Job) {
	j.Status = "running"
	j.StartedAt = time.Now().UTC()

	src, err := gorm.Open(sqlite.Open(j.SourcePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		j.Status = "failed"
		j.LastError = "open source: " + err.Error()
		return
	}

	dst, err := gorm.Open(postgres.Open(j.TargetDSN), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: false,
	})
	if err != nil {
		j.Status = "failed"
		j.LastError = "open target: " + err.Error()
		return
	}

	log.Printf("[migrate] job %s: start tables=%v", j.ID, j.Tables)

	for _, table := range j.Tables {
		if j.Status == "cancelled" {
			return
		}
		if err := migrateTable(j, src, dst, table); err != nil {
			j.Status = "failed"
			j.LastError = fmt.Sprintf("table %s: %v", table, err)
			log.Printf("[migrate] job %s FAILED: %s", j.ID, j.LastError)
			return
		}
	}

	now := time.Now().UTC()
	j.FinishedAt = &now
	j.Status = "completed"
	log.Printf("[migrate] job %s: completed read=%d written=%d errors=%d",
		j.ID, j.TotalRead, j.TotalWritten, j.Errors)
}

func migrateTable(j *Job, src, dst *gorm.DB, table string) error {
	var total int64
	if err := src.Table(table).Count(&total).Error; err != nil {
		return fmt.Errorf("count: %w", err)
	}
	atomic.AddInt64(&j.TotalRead, total)

	if total == 0 {
		log.Printf("[migrate] job %s: table %s empty", j.ID, table)
		return nil
	}

	log.Printf("[migrate] job %s: table %s rows=%d", j.ID, table, total)

	chunk := j.BatchSize
	if chunk <= 0 {
		chunk = 500
	}

	// First, read one row to know the columns.
	var first []map[string]any
	if err := src.Table(table).Limit(1).Find(&first).Error; err != nil {
		return fmt.Errorf("read first row: %w", err)
	}
	if len(first) == 0 {
		return nil
	}

	// Drop and recreate the target table with all TEXT columns.
	if err := dst.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %q`, table)).Error; err != nil {
		return fmt.Errorf("drop target: %w", err)
	}
	ddl := buildDDL(table, first[0])
	log.Printf("[migrate] job %s: DDL for %s: %s", j.ID, table, ddl)
	if err := dst.Exec(ddl).Error; err != nil {
		return fmt.Errorf("create target: %w", err)
	}

	// Now iterate all rows and insert.
	offset := 0
	written := 0
	for offset < int(total) {
		var batch []map[string]any
		err := src.Table(table).Limit(chunk).Offset(offset).Find(&batch).Error
		if err != nil {
			log.Printf("[migrate] job %s: read offset=%d: %v", j.ID, offset, err)
			atomic.AddInt64(&j.Errors, 1)
			offset += chunk
			continue
		}
		if len(batch) == 0 {
			break
		}

		// Stringify all values (SQLite gives time.Time, int64, bool, etc.).
		for i := range batch {
			batch[i] = stringifyRow(batch[i])
		}

		// Insert with retry.
		var insertErr error
		for attempt := 1; attempt <= 3; attempt++ {
			insertErr = dst.Table(table).CreateInBatches(batch, len(batch)).Error
			if insertErr == nil {
				break
			}
			log.Printf("[migrate] job %s: insert %s offset=%d attempt=%d: %v",
				j.ID, table, offset, attempt, insertErr)
			time.Sleep(200 * time.Millisecond)
		}
		if insertErr != nil {
			atomic.AddInt64(&j.Errors, 1)
			j.LastError = fmt.Sprintf("insert %s offset=%d: %v", table, offset, insertErr)
			// Continue with next batch rather than abort entire job.
		} else {
			written += len(batch)
			atomic.AddInt64(&j.TotalWritten, int64(len(batch)))
		}

		offset += len(batch)
	}

	log.Printf("[migrate] job %s: table %s done, wrote %d/%d", j.ID, table, written, total)
	return nil
}

// buildDDL creates a CREATE TABLE statement where every column is TEXT,
// using the first row's keys.
func buildDDL(table string, sample map[string]any) string {
	cols := make([]string, 0, len(sample))
	for k := range sample {
		cols = append(cols, fmt.Sprintf("%q TEXT", k))
	}
	ddl := fmt.Sprintf("CREATE TABLE %q (", table)
	for i, c := range cols {
		if i > 0 {
			ddl += ", "
		}
		ddl += c
	}
	ddl += ")"
	return ddl
}

// stringifyRow converts every value to a string or nil, so it fits TEXT.
func stringifyRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for k, v := range row {
		switch x := v.(type) {
		case nil:
			out[k] = nil
		case string:
			out[k] = x
		case []byte:
			out[k] = string(x)
		case time.Time:
			out[k] = x.UTC().Format(time.RFC3339Nano)
		case bool:
			if x {
				out[k] = "true"
			} else {
				out[k] = "false"
			}
		default:
			out[k] = fmt.Sprintf("%v", x)
		}
	}
	return out
}

var _ = context.Background
