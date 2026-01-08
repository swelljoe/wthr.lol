package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/swelljoe/wthr.lol/internal/db"
)

func TestHandleHealth(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.HandleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", contentType)
	}
}

func TestHandleIndex(t *testing.T) {
	h := &Handlers{}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}
}

func TestHandleIndexNotFound(t *testing.T) {
	h := &Handlers{}

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status NotFound, got %v", resp.StatusCode)
	}
}

// mockDB is a mock implementation of the database for testing
type mockDB struct {
	searchPlacesFunc    func(query string) ([]db.Place, error)
	pingFunc            func() error
	saveAppInterestFunc func(email string, android bool, ios bool, country string) error
}

func (m *mockDB) SearchPlaces(query string) ([]db.Place, error) {
	if m.searchPlacesFunc != nil {
		return m.searchPlacesFunc(query)
	}
	return nil, nil
}

func (m *mockDB) Ping() error {
	if m.pingFunc != nil {
		return m.pingFunc()
	}
	return nil
}

func (m *mockDB) SaveAppInterest(email string, android bool, ios bool, country string) error {
	if m.saveAppInterestFunc != nil {
		return m.saveAppInterestFunc(email, android, ios, country)
	}
	return nil
}

func TestHandleSearch_QueryTooShort(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	tests := []struct {
		name  string
		query string
	}{
		{"empty query", ""},
		{"single character", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/search?q="+tt.query, nil)
			w := httptest.NewRecorder()

			h.HandleSearch(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status OK, got %v", resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %v", contentType)
			}

			// Should return empty array
			var result []db.Place
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Errorf("failed to decode response: %v", err)
			}
			if len(result) != 0 {
				t.Errorf("expected empty array, got %d items", len(result))
			}
		})
	}
}

func TestHandleSearch_SuccessWithResults(t *testing.T) {
	expectedPlaces := []db.Place{
		{
			Name:      "San Francisco",
			State:     "CA",
			Zip:       "94102",
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
		{
			Name:      "San Jose",
			State:     "CA",
			Zip:       "95110",
			Latitude:  37.3382,
			Longitude: -121.8863,
		},
	}

	mock := &mockDB{
		searchPlacesFunc: func(query string) ([]db.Place, error) {
			if query == "San" {
				return expectedPlaces, nil
			}
			return nil, nil
		},
	}

	h := &Handlers{db: mock}

	req := httptest.NewRequest("GET", "/search?q=San", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", contentType)
	}

	var result []db.Place
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if len(result) != len(expectedPlaces) {
		t.Errorf("expected %d places, got %d", len(expectedPlaces), len(result))
	}

	for i, place := range result {
		if place.Name != expectedPlaces[i].Name {
			t.Errorf("expected name %s, got %s", expectedPlaces[i].Name, place.Name)
		}
		if place.State != expectedPlaces[i].State {
			t.Errorf("expected state %s, got %s", expectedPlaces[i].State, place.State)
		}
		if place.Zip != expectedPlaces[i].Zip {
			t.Errorf("expected zip %s, got %s", expectedPlaces[i].Zip, place.Zip)
		}
		if place.Latitude != expectedPlaces[i].Latitude {
			t.Errorf("expected latitude %f, got %f", expectedPlaces[i].Latitude, place.Latitude)
		}
		if place.Longitude != expectedPlaces[i].Longitude {
			t.Errorf("expected longitude %f, got %f", expectedPlaces[i].Longitude, place.Longitude)
		}
	}
}

func TestHandleSearch_EmptyResults(t *testing.T) {
	mock := &mockDB{
		searchPlacesFunc: func(query string) ([]db.Place, error) {
			return nil, nil
		},
	}

	h := &Handlers{db: mock}

	req := httptest.NewRequest("GET", "/search?q=NonexistentPlace", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", contentType)
	}

	var result []db.Place
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestHandleSearch_DatabaseError(t *testing.T) {
	mock := &mockDB{
		searchPlacesFunc: func(query string) ([]db.Place, error) {
			return nil, errors.New("database connection failed")
		},
	}

	h := &Handlers{db: mock}

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status InternalServerError, got %v", resp.StatusCode)
	}
}

func TestHandleSearch_QueryValidation(t *testing.T) {
	mock := &mockDB{
		searchPlacesFunc: func(query string) ([]db.Place, error) {
			return []db.Place{
				{Name: "Test City", State: "TX", Latitude: 30.0, Longitude: -97.0},
			}, nil
		},
	}

	h := &Handlers{db: mock}

	tests := []struct {
		name          string
		query         string
		expectResults bool
	}{
		{"minimum valid query", "ab", true},
		{"typical query", "Austin", true},
		{"query with spaces", "San+Francisco", true},
		{"query with numbers", "New+York+10001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/search?q="+tt.query, nil)
			w := httptest.NewRecorder()

			h.HandleSearch(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status OK, got %v", resp.StatusCode)
			}

			var result []db.Place
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Errorf("failed to decode response: %v", err)
			}

			if tt.expectResults && len(result) == 0 {
				t.Errorf("expected results for query %q, got none", tt.query)
			}
		})
	}
}

func TestHandleAppInterest_ValidEmail(t *testing.T) {
	mock := &mockDB{
		saveAppInterestFunc: func(email string, android bool, ios bool, country string) error {
			return nil
		},
	}

	h := &Handlers{db: mock}

	tests := []struct {
		name  string
		email string
	}{
		{"simple email", "user@example.com"},
		{"email with subdomain", "user@mail.example.com"},
		{"email with plus", "user+tag@example.com"},
		{"email with dots", "first.last@example.com"},
		{"email with hyphen", "user@my-domain.com"},
		{"email with numbers", "user123@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := `{"email":"` + tt.email + `","android":true,"ios":false,"country":"US"}`
			req := httptest.NewRequest("POST", "/app-interest", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.HandleAppInterest(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status OK for email %q, got %v", tt.email, resp.StatusCode)
			}
		})
	}
}

func TestHandleAppInterest_InvalidEmail(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	tests := []struct {
		name  string
		email string
	}{
		{"missing @", "userexample.com"},
		{"missing domain", "user@"},
		{"missing local part", "@example.com"},
		{"double @", "user@@example.com"},
		{"spaces", "user name@example.com"},
		{"invalid characters", "user<>@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := `{"email":"` + tt.email + `","android":true,"ios":false,"country":"US"}`
			req := httptest.NewRequest("POST", "/app-interest", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.HandleAppInterest(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status BadRequest for email %q, got %v", tt.email, resp.StatusCode)
			}
		})
	}
}

func TestHandleAppInterest_EmptyEmail(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	payload := `{"email":"","android":true,"ios":false,"country":"US"}`
	req := httptest.NewRequest("POST", "/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest for empty email, got %v", resp.StatusCode)
	}
}
