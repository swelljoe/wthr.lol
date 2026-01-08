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

func TestHandleAppInterest_MethodNotAllowed(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	req := httptest.NewRequest("GET", "/api/app-interest", nil)
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status MethodNotAllowed, got %v", resp.StatusCode)
	}
}

func TestHandleAppInterest_InvalidJSON(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":true,"invalid"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest for invalid JSON, got %v", resp.StatusCode)
	}
}

func TestHandleAppInterest_EmptyEmail(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	payload := `{"email":"","android":true,"ios":false,"country":"US"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest for empty email, got %v", resp.StatusCode)
	}
}

func TestHandleAppInterest_MissingPlatformSelection(t *testing.T) {
	mock := &mockDB{}
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":false,"ios":false,"country":"US"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest for missing platform selection, got %v", resp.StatusCode)
	}
}

// Helper function to create a mock DB that captures saved parameters
func createAppInterestMockDB() (*mockDB, *string, *bool, *bool, *string) {
	savedEmail := ""
	savedAndroid := false
	savedIOS := false
	savedCountry := ""

	mock := &mockDB{
		saveAppInterestFunc: func(email string, android bool, ios bool, country string) error {
			savedEmail = email
			savedAndroid = android
			savedIOS = ios
			savedCountry = country
			return nil
		},
	}

	return mock, &savedEmail, &savedAndroid, &savedIOS, &savedCountry
}

func TestHandleAppInterest_SuccessAndroidOnly(t *testing.T) {
	mock, savedEmail, savedAndroid, savedIOS, savedCountry := createAppInterestMockDB()
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":true,"ios":false,"country":"US"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}

	if *savedEmail != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", *savedEmail)
	}
	if !*savedAndroid {
		t.Errorf("expected android to be true, got %v", *savedAndroid)
	}
	if *savedIOS {
		t.Errorf("expected ios to be false, got %v", *savedIOS)
	}
	if *savedCountry != "US" {
		t.Errorf("expected country US, got %s", *savedCountry)
	}
}

func TestHandleAppInterest_SuccessIOSOnly(t *testing.T) {
	mock, savedEmail, savedAndroid, savedIOS, savedCountry := createAppInterestMockDB()
	h := &Handlers{db: mock}

	payload := `{"email":"ios@example.com","android":false,"ios":true,"country":"CA"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	if *savedEmail != "ios@example.com" {
		t.Errorf("expected email ios@example.com, got %s", *savedEmail)
	}
	if *savedAndroid {
		t.Errorf("expected android to be false, got %v", *savedAndroid)
	}
	if !*savedIOS {
		t.Errorf("expected ios to be true, got %v", *savedIOS)
	}
	if *savedCountry != "CA" {
		t.Errorf("expected country CA, got %s", *savedCountry)
	}
}

func TestHandleAppInterest_SuccessBothPlatforms(t *testing.T) {
	mock, savedEmail, savedAndroid, savedIOS, savedCountry := createAppInterestMockDB()
	h := &Handlers{db: mock}

	payload := `{"email":"both@example.com","android":true,"ios":true,"country":"UK"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	if *savedEmail != "both@example.com" {
		t.Errorf("expected email both@example.com, got %s", *savedEmail)
	}
	if !*savedAndroid {
		t.Errorf("expected android to be true, got %v", *savedAndroid)
	}
	if !*savedIOS {
		t.Errorf("expected ios to be true, got %v", *savedIOS)
	}
	if *savedCountry != "UK" {
		t.Errorf("expected country UK, got %s", *savedCountry)
	}
}

func TestHandleAppInterest_DatabaseError(t *testing.T) {
	mock := &mockDB{
		saveAppInterestFunc: func(email string, android bool, ios bool, country string) error {
			return errors.New("database connection failed")
		},
	}
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":true,"ios":false,"country":"US"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status InternalServerError, got %v", resp.StatusCode)
	}
}

func TestHandleAppInterest_NoDatabaseDevelopmentMode(t *testing.T) {
	// Test with nil database (development mode)
	h := &Handlers{db: nil}

	payload := `{"email":"dev@example.com","android":true,"ios":true,"country":"FR"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK in development mode, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}
}

func TestHandleAppInterest_EmptyCountry(t *testing.T) {
	// Test that empty country is accepted (country is not validated as required)
	mock, savedEmail, savedAndroid, savedIOS, savedCountry := createAppInterestMockDB()
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":true,"ios":false,"country":""}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	// Currently the handler doesn't validate country as required, so this should succeed
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK for empty country, got %v", resp.StatusCode)
	}

	if *savedEmail != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", *savedEmail)
	}
	if !*savedAndroid {
		t.Errorf("expected android to be true, got %v", *savedAndroid)
	}
	if *savedIOS {
		t.Errorf("expected ios to be false, got %v", *savedIOS)
	}
	if *savedCountry != "" {
		t.Errorf("expected empty country, got %s", *savedCountry)
	}
}

func TestHandleAppInterest_UnknownFields(t *testing.T) {
	// Test that unknown fields are rejected due to DisallowUnknownFields
	mock := &mockDB{}
	h := &Handlers{db: mock}

	payload := `{"email":"test@example.com","android":true,"ios":false,"country":"US","unknownField":"value"}`
	req := httptest.NewRequest("POST", "/api/app-interest", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAppInterest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest for unknown fields, got %v", resp.StatusCode)
	}
}
