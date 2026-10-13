package intersect

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// knownTables caches which tables we've already created in PostgreSQL.
var knownTables sync.Map // map[string]bool

// ensureTable creates the table if it doesn't exist. Columns are inferred
// from the first row's keys; all columns are TEXT to accept anything.
func (x *Intersect) ensureTable(table string, sample map[string]any) error {
	if _, ok := knownTables.Load(table); ok {
		return nil
	}

	// Sanitize table name.
	if !isValidIdent(table) {
		return fmt.Errorf("invalid table name: %q", table)
	}

	// Check if table exists.
	var exists int64
	err := x.db.Raw(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?`,
		table,
	).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("check table: %w", err)
	}

	if exists > 0 {
		knownTables.Store(table, true)
		return nil
	}

	// Build CREATE TABLE from sample keys.
	keys := make([]string, 0, len(sample))
	for k := range sample {
		if isValidIdent(k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return fmt.Errorf("cannot infer columns for table %q", table)
	}

	var cols []string
	for _, k := range keys {
		cols = append(cols, fmt.Sprintf("%q TEXT", k))
	}

	ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q (%s)`, table, strings.Join(cols, ", "))
	if err := x.db.Exec(ddl).Error; err != nil {
		return fmt.Errorf("create table %s: %w", table, err)
	}

	knownTables.Store(table, true)
	return nil
}

func isValidIdent(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				return false
			}
		} else {
			if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				return false
			}
		}
	}
	return true
}
