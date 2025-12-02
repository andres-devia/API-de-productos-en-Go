package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteItemRepository implements the ItemRepository interface using SQLite.
// This file only handles repository creation, schema initialization and cleanup.
type SQLiteItemRepository struct {
	DB *sql.DB
}

// NewSQLiteItemRepository creates a new SQLite repository instance
// and initializes the database schema.
func NewSQLiteItemRepository(dbPath string) (*SQLiteItemRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &SQLiteItemRepository{DB: db}

	// Initialize the DB schema
	if err := repo.initSchema(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates the items table if it does not exist.
func (r *SQLiteItemRepository) initSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			image_url TEXT NOT NULL,
			description TEXT NOT NULL,
			price REAL NOT NULL,
			rating REAL NOT NULL,
			specifications TEXT NOT NULL
		)
	`

	if _, err := r.DB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Close closes the repository database connection.
func (r *SQLiteItemRepository) Close() error {
	return r.DB.Close()
}
