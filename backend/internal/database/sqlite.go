package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	// Blank import for the file source driver used by golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// Blank import for the CGo-based SQLite driver.
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database connection, enables WAL mode, and applies all
// pending database migrations. It's the single entry point for database setup.
func InitDB(dataSourceName string) (*sql.DB, error) {
	// Ensure the parent directory for the database file exists.
	dir := filepath.Dir(dataSourceName)
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

	// Enable Write-Ahead Logging (WAL) mode for better concurrency.
	// This allows read operations to proceed while write operations are in progress.
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		slog.Warn("Failed to enable WAL mode for SQLite, continuing without it.", "error", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations orchestrates the database schema migration process. It ensures the
// database schema is always up-to-date with the version defined in the SQL files.
func runMigrations(db *sql.DB) error {
	// Create a migration driver instance for SQLite.
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create sqlite migration driver: %w", err)
	}

	// Reliably locate the migrations directory regardless of the execution context.
	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return err
	}

	// Initialize the migrate instance with the file source and database driver.
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	slog.Info("Applying database migrations...")
	// The `Up` command is idempotent; it applies only the migrations that haven't
	// been applied yet. `migrate.ErrNoChange` is not a critical error.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Log the final state of the database schema for visibility.
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	slog.Info("Database migration process complete", "version", version, "is_dirty", dirty)
	if dirty {
		slog.Error("DATABASE IS DIRTY. This indicates a failed migration and requires manual intervention.")
	}
	return nil
}

// getMigrationsPath dynamically finds the path to the migrations directory.
// This robust approach handles different execution contexts: running from source
// via `go run`, running tests via `go test`, or running in the final Docker container.
func getMigrationsPath() (string, error) {
	// `runtime.Caller(0)` returns information about the function that calls it.
	// Here, it gives us the path to this source file (`sqlite.go`).
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot determine current file path for migration discovery")
	}
	// The directory containing this file is `.../internal/database`.
	basepath := filepath.Dir(b)

	// The migrations are located in a sibling directory.
	localPath := filepath.Join(basepath, "migrations")

	// This path will resolve correctly for local development and testing.
	if _, err := os.Stat(localPath); err == nil {
		return "file://" + localPath, nil
	}

	// This is a fallback for the production Docker container, where migrations
	// are copied to a specific, known location.
	containerPath := "/app/migrations"
	if _, err := os.Stat(containerPath); err == nil {
		return "file://" + containerPath, nil
	}

	return "", fmt.Errorf("migrations directory not found: tried %s and %s", localPath, containerPath)
}
