package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a database connection
type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	DSN string
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	// Use local sqlite file
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "wthr.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &DB{db}, nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS weather_cache (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	return err
}

// CacheEntry represents a cached weather response
type CacheEntry struct {
	Data      string
	ExpiresAt time.Time
}

// GetCachedWeather retrieves weather data if valid
func (db *DB) GetCachedWeather(lat, lon float64) (*CacheEntry, error) {
	// Round to 2 decimal places to match key generation
	key := fmt.Sprintf("%.2f,%.2f", lat, lon)
	
	var data string
	var expiresAt time.Time
	
	err := db.QueryRow("SELECT data, expires_at FROM weather_cache WHERE id = ? AND expires_at > ?", key, time.Now()).Scan(&data, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}
	
	return &CacheEntry{
		Data:      data,
		ExpiresAt: expiresAt,
	}, nil
}

// SetCachedWeather saves weather data
func (db *DB) SetCachedWeather(lat, lon float64, data string, duration time.Duration) error {
	key := fmt.Sprintf("%.2f,%.2f", lat, lon)
	expiresAt := time.Now().Add(duration)
	
	_, err := db.Exec(`
		INSERT INTO weather_cache (id, data, expires_at) 
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			data = excluded.data,
			expires_at = excluded.expires_at,
			created_at = CURRENT_TIMESTAMP
	`, key, data, expiresAt)
	
	return err
}
