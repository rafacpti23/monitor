package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	log.Println("Database successfully initialized/migrated.")
	return nil
}

// LogSystemAction logs system events easily.
func LogSystemAction(userID int64, action, details, ip string) {
	_, err := DB.Exec("INSERT INTO system_logs (user_id, action, details, ip) VALUES (?, ?, ?, ?)", userID, action, details, ip)
	if err != nil {
		log.Printf("Failed to log system action: %v", err)
	}
}
