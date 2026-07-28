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
