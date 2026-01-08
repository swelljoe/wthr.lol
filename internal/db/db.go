package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
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
	if err != nil {
		return err
	}

	placesQuery := `
	CREATE TABLE IF NOT EXISTS places (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		state TEXT NOT NULL,
		zip TEXT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		population INTEGER DEFAULT 0
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS places_fts USING fts5(
		name,
		state,
		zip,
		details,
		content='places',
		content_rowid='id',
		tokenize='porter ascii'
	);

	CREATE TRIGGER IF NOT EXISTS places_ai AFTER INSERT ON places BEGIN
		INSERT INTO places_fts(rowid, name, state, zip, details)
		VALUES (new.id, new.name, new.state, new.zip, new.name || ', ' || new.state || ' ' || COALESCE(new.zip, ''));
	END;

	CREATE TRIGGER IF NOT EXISTS places_ad AFTER DELETE ON places BEGIN
		INSERT INTO places_fts(places_fts, rowid, name, state, zip, details)
		VALUES('delete', old.id, old.name, old.state, old.zip, old.name || ', ' || old.state || ' ' || COALESCE(old.zip, ''));
	END;

	CREATE TRIGGER IF NOT EXISTS places_au AFTER UPDATE ON places BEGIN
		INSERT INTO places_fts(places_fts, rowid, name, state, zip, details)
		VALUES('delete', old.id, old.name, old.state, old.zip, old.name || ', ' || old.state || ' ' || COALESCE(old.zip, ''));
		INSERT INTO places_fts(rowid, name, state, zip, details)
		VALUES (new.id, new.name, new.state, new.zip, new.name || ', ' || new.state || ' ' || COALESCE(new.zip, ''));
	END;
	`
	_, err = db.Exec(placesQuery)
	if err != nil {
		return err
	}

	appQuery := `
	CREATE TABLE IF NOT EXISTS app_interest (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		android INTEGER NOT NULL DEFAULT 0,
		ios INTEGER NOT NULL DEFAULT 0,
		country TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(appQuery)
	if err != nil {
		return err
	}

	return nil
}

// CacheEntry represents a cached weather response
type CacheEntry struct {
	Data      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// GetCachedWeather retrieves weather data if valid
func (db *DB) GetCachedWeather(lat, lon float64) (*CacheEntry, error) {
	// Round to 2 decimal places to match key generation
	key := fmt.Sprintf("%.2f,%.2f", lat, lon)

	var data string
	var expiresAt, createdAt time.Time

	err := db.QueryRow("SELECT data, expires_at, created_at FROM weather_cache WHERE id = ? AND expires_at > ?", key, time.Now()).Scan(&data, &expiresAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	return &CacheEntry{
		Data:      data,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
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

// Place represents a search result
type Place struct {
	Name      string  `json:"name"`
	State     string  `json:"state"`
	Zip       string  `json:"zip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SearchPlaces searches for places matching the query
func (db *DB) SearchPlaces(query string) ([]Place, error) {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return nil, nil
	}

	// Construct FTS5 query: simple prefix matching
	// e.g. "San Fran" -> "San* AND Fran*"
	var ftsParts []string
	for _, term := range terms {
		// Sanitize term: escape double quotes and remove characters that could break FTS5 syntax
		// FTS5 special characters include: " ( ) AND OR NOT NEAR
		sanitized := sanitizeFTSTerm(term)
		if sanitized == "" {
			continue // Skip empty terms after sanitization
		}
		ftsParts = append(ftsParts, sanitized+"*") // Prefix match
	}

	if len(ftsParts) == 0 {
		return nil, nil
	}

	ftsQuery := strings.Join(ftsParts, " AND ")

	q := `
	SELECT p.name, p.state, p.zip, p.latitude, p.longitude
	FROM places p
	JOIN places_fts ON p.id = places_fts.rowid
	WHERE places_fts MATCH ?
	ORDER BY rank
	LIMIT 10;
	`

	rows, err := db.Query(q, ftsQuery)
	if err != nil {
		// Provide more context about the error, especially for FTS5 query issues
		return nil, fmt.Errorf("failed to execute search query (query: %q): %w", ftsQuery, err)
	}
	defer rows.Close()

	var places []Place
	for rows.Next() {
		var p Place
		var zip sql.NullString
		if err := rows.Scan(&p.Name, &p.State, &zip, &p.Latitude, &p.Longitude); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		p.Zip = zip.String
		places = append(places, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return places, nil
}

// sanitizeFTSTerm sanitizes a search term for use in FTS5 queries
// It removes or escapes characters that have special meaning in FTS5
func sanitizeFTSTerm(term string) string {
	// Remove FTS5 operators and special characters that could break the query
	// Keep alphanumeric, spaces, and some common punctuation
	var result strings.Builder
	for _, r := range term {
		// Skip FTS5 special characters that have syntactic meaning
		if r == '"' || r == '(' || r == ')' || r == '*' || r == '^' {
			continue
		}
		// Keep letters, digits, spaces, hyphens, and periods
		// Using direct comparisons for ASCII punctuation while supporting Unicode letters/digits
		if r == ' ' || r == '-' || r == '.' {
			result.WriteRune(r)
		} else if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			result.WriteRune(r)
		}
		// Note: We intentionally keep ASCII-only to avoid issues with Unicode characters in FTS5
	}
	return strings.TrimSpace(result.String())
}

// SaveAppInterest inserts a new record into the app_interest table
func (db *DB) SaveAppInterest(email string, android bool, ios bool, country string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	a := 0
	if android {
		a = 1
	}
	i := 0
	if ios {
		i = 1
	}

	_, err := db.Exec(`
		INSERT INTO app_interest (email, android, ios, country)
		VALUES (?, ?, ?, ?)
	`, email, a, i, country)

	return err
}
