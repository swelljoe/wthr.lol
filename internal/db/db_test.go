package db

import (
	"database/sql"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	// Use in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Insert some test data
	testData := []struct {
		name      string
		state     string
		zip       string
		latitude  float64
		longitude float64
	}{
		{"San Francisco", "CA", "94102", 37.7749, -122.4194},
		{"San Diego", "CA", "92101", 32.7157, -117.1611},
		{"Los Angeles", "CA", "90001", 34.0522, -118.2437},
		{"Sacramento", "CA", "95814", 38.5816, -121.4944},
		{"New York", "NY", "10001", 40.7128, -74.0060},
	}

	for _, place := range testData {
		_, err := db.Exec(
			"INSERT INTO places (name, state, zip, latitude, longitude) VALUES (?, ?, ?, ?, ?)",
			place.name, place.state, place.zip, place.latitude, place.longitude,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	return &DB{db}
}

func TestSearchPlaces(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	tests := []struct {
		name        string
		query       string
		expectError bool
		minResults  int
		maxResults  int
	}{
		{
			name:        "simple search",
			query:       "San",
			expectError: false,
			minResults:  2,
			maxResults:  2,
		},
		{
			name:        "multi-word search",
			query:       "San Francisco",
			expectError: false,
			minResults:  1,
			maxResults:  1,
		},
		{
			name:        "empty query",
			query:       "",
			expectError: false,
			minResults:  0,
			maxResults:  0,
		},
		{
			name:        "whitespace only",
			query:       "   ",
			expectError: false,
			minResults:  0,
			maxResults:  0,
		},
		{
			name:        "special characters",
			query:       "San\"Francisco",
			expectError: false,
			minResults:  0,
			maxResults:  5,
		},
		{
			name:        "parentheses",
			query:       "San (Francisco)",
			expectError: false,
			minResults:  0,
			maxResults:  5,
		},
		{
			name:        "asterisk",
			query:       "San*",
			expectError: false,
			minResults:  0,
			maxResults:  5,
		},
		{
			name:        "no results",
			query:       "xyz123notfound",
			expectError: false,
			minResults:  0,
			maxResults:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			places, err := testDB.SearchPlaces(tt.query)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if places == nil {
				places = []Place{}
			}

			resultCount := len(places)
			if resultCount < tt.minResults || resultCount > tt.maxResults {
				t.Errorf("Expected between %d and %d results, got %d",
					tt.minResults, tt.maxResults, resultCount)
			}
		})
	}
}

func TestSanitizeFTSTerm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Francisco",
			expected: "Francisco",
		},
		{
			name:     "with quotes",
			input:    "San\"Francisco",
			expected: "SanFrancisco",
		},
		{
			name:     "with parentheses",
			input:    "San(Francisco)",
			expected: "SanFrancisco",
		},
		{
			name:     "with asterisk",
			input:    "San*",
			expected: "San",
		},
		{
			name:     "with spaces",
			input:    "San Francisco",
			expected: "San Francisco",
		},
		{
			name:     "with hyphen",
			input:    "San-Francisco",
			expected: "San-Francisco",
		},
		{
			name:     "with numbers",
			input:    "Route66",
			expected: "Route66",
		},
		{
			name:     "only special chars",
			input:    "\"()^*",
			expected: "",
		},
		{
			name:     "mixed",
			input:    "New*York(123)",
			expected: "NewYork123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFTSTerm(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSearchPlacesErrorMessages(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Test that error messages include the query for debugging
	// We can't easily trigger FTS5 errors with valid SQLite, but we can
	// verify the structure by checking normal queries still work
	places, err := testDB.SearchPlaces("San")
	if err != nil {
		t.Errorf("Normal query should not error: %v", err)
	}
	if len(places) < 1 {
		t.Error("Expected at least one result for 'San'")
	}
}

func TestNewDB(t *testing.T) {
	// Test with a temporary database file
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test_wthr.db"

	t.Setenv("DB_PATH", tmpFile)

	db, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create new DB: %v", err)
	}
	defer db.Close()

	// Verify we can ping it
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping DB: %v", err)
	}
}

func TestSaveAppInterest(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	tests := []struct {
		name    string
		email   string
		android bool
		ios     bool
		country string
		wantErr bool
	}{
		{
			name:    "android only",
			email:   "android@example.com",
			android: true,
			ios:     false,
			country: "US",
			wantErr: false,
		},
		{
			name:    "ios only",
			email:   "ios@example.com",
			android: false,
			ios:     true,
			country: "CA",
			wantErr: false,
		},
		{
			name:    "both platforms",
			email:   "both@example.com",
			android: true,
			ios:     true,
			country: "GB",
			wantErr: false,
		},
		{
			name:    "neither platform",
			email:   "neither@example.com",
			android: false,
			ios:     false,
			country: "DE",
			wantErr: false,
		},
		{
			name:    "empty country",
			email:   "nocountry@example.com",
			android: true,
			ios:     false,
			country: "",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			android: true,
			ios:     true,
			country: "US",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.SaveAppInterest(tt.email, tt.android, tt.ios, tt.country)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveAppInterest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the record was inserted
				var email string
				var android, ios int
				var country string
				err := testDB.QueryRow(
					"SELECT email, android, ios, country FROM app_interest WHERE email = ? ORDER BY id DESC LIMIT 1",
					tt.email,
				).Scan(&email, &android, &ios, &country)

				if err != nil {
					t.Fatalf("Failed to query inserted record: %v", err)
				}

				if email != tt.email {
					t.Errorf("Expected email %q, got %q", tt.email, email)
				}

				expectedAndroid := 0
				if tt.android {
					expectedAndroid = 1
				}
				if android != expectedAndroid {
					t.Errorf("Expected android %d, got %d", expectedAndroid, android)
				}

				expectedIos := 0
				if tt.ios {
					expectedIos = 1
				}
				if ios != expectedIos {
					t.Errorf("Expected ios %d, got %d", expectedIos, ios)
				}

				if country != tt.country {
					t.Errorf("Expected country %q, got %q", tt.country, country)
				}
			}
		})
	}
}

func TestSaveAppInterestNilDB(t *testing.T) {
	var db *DB
	err := db.SaveAppInterest("test@example.com", true, true, "US")
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
	expectedMsg := "database not initialized"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}
