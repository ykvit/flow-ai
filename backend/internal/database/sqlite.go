package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Import the sqlite3 driver.
)

// InitDB connects to the SQLite database and runs migrations.
func InitDB(dataSourceName string) (*sql.DB, error) {
	// Ensure the directory for the database file exists.
	dir := filepath.Dir(dataSourceName)
	// [FIX] G301: Use more restrictive directory permissions as recommended by gosec.
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable WAL mode for better concurrency.
	// This allows readers to not block writers.
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		// [FIX] Switched to slog for consistent, structured logging.
		slog.Warn("Failed to enable WAL mode for SQLite, continuing without it.", "error", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createTables executes the SQL statements to create the database schema.
func createTables(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS chats (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			model TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_chats_user_id_updated_at ON chats(user_id, updated_at DESC);

		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			chat_id TEXT NOT NULL,
			parent_id TEXT,
			role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
			status TEXT,
			content TEXT NOT NULL,
			model TEXT,
			timestamp DATETIME NOT NULL,
			metadata TEXT,
			context BLOB,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES messages(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_messages_chat_id_active_timestamp ON messages(chat_id, is_active, timestamp);
		CREATE INDEX IF NOT EXISTS idx_messages_parent_id ON messages(parent_id);

		CREATE TABLE IF NOT EXISTS attachments (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`
	_, err := db.Exec(schema)
	return err
}
