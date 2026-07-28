package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("sqlite", dsn+"?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	return nil
}

func Migrate(schemaPath string) error {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema: %v", err)
	}
	_, err = DB.Exec(string(b))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	// Idempotent post-schema ALTERs. Each entry is applied once; errors that
	// indicate the column already exists are ignored.
	alters := []string{
		`ALTER TABLE servers ADD COLUMN interval_seconds INTEGER NOT NULL DEFAULT 0`,
	}

	// Idempotent CREATE TABLE for features added after initial deploy.
	// CREATE TABLE IF NOT EXISTS is safe to run every boot.
	createStmts := []string{
		`CREATE TABLE IF NOT EXISTS papi_panels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL DEFAULT '',
			base_url TEXT NOT NULL DEFAULT 'https://papi.api.br',
			panel_token TEXT NOT NULL DEFAULT '',
			check_interval_sec INTEGER NOT NULL DEFAULT 60,
			status TEXT NOT NULL DEFAULT 'pending',
			last_checked DATETIME,
			last_error TEXT NOT NULL DEFAULT '',
			total_instances INTEGER NOT NULL DEFAULT 0,
			connected_instances INTEGER NOT NULL DEFAULT 0,
			channels TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_papi_panels_user_id ON papi_panels(user_id)`,
		`CREATE TABLE IF NOT EXISTS papi_instances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			panel_id INTEGER NOT NULL REFERENCES papi_panels(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			instance_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			phone_number TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			last_seen DATETIME,
			updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
			UNIQUE(panel_id, instance_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_papi_instances_panel_id ON papi_instances(panel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_papi_instances_user_id ON papi_instances(user_id)`,
	}
	for _, stmt := range createStmts {
		if _, err := DB.Exec(stmt); err != nil {
			log.Printf("[migrate] create skipped or failed: %v", err)
		}
	}
	for _, stmt := range alters {
		if _, err := DB.Exec(stmt); err != nil {
			// SQLite returns "duplicate column name" when the column exists.
			if !isDuplicateColumnErr(err) {
				log.Printf("[migrate] alter skipped or failed: %s | %v", stmt, err)
			}
		} else {
			log.Printf("[migrate] applied: %s", stmt)
		}
	}

	log.Println("Database successfully initialized/migrated.")
	return nil
}

func isDuplicateColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate column name") || strings.Contains(msg, "already exists")
}

// LogSystemAction logs system events easily.
func LogSystemAction(userID int64, action, details, ip string) {
	_, err := DB.Exec("INSERT INTO system_logs (user_id, action, details, ip) VALUES (?, ?, ?, ?)", userID, action, details, ip)
	if err != nil {
		log.Printf("Failed to log system action: %v", err)
	}
}
