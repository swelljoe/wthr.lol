package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockRoundTripper is a custom RoundTripper for testing
type mockRoundTripper struct {
	handler http.Handler
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	resp := rec.Result()
	return resp, nil
}

// TestReverseGeocode_CityWithState tests successful reverse geocoding with city and state
func TestReverseGeocode_CityWithState(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "San Francisco, California, United States",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{
			City:  "San Francisco",
			State: "California",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("expected format=json, got %s", r.URL.Query().Get("format"))
		}
		if r.URL.Query().Get("lat") != "37.774900" {
			t.Errorf("expected lat=37.774900, got %s", r.URL.Query().Get("lat"))
		}
		if r.URL.Query().Get("lon") != "-122.419400" {
			t.Errorf("expected lon=-122.419400, got %s", r.URL.Query().Get("lon"))
		}
		if r.URL.Query().Get("zoom") != "10" {
			t.Errorf("expected zoom=10, got %s", r.URL.Query().Get("zoom"))
		}
		if r.URL.Query().Get("addressdetails") != "1" {
			t.Errorf("expected addressdetails=1, got %s", r.URL.Query().Get("addressdetails"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(37.7749, -122.4194)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "San Francisco, California"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestReverseGeocode_TownWithState tests reverse geocoding with town and state
func TestReverseGeocode_TownWithState(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "Smalltown, Texas, United States",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{
			Town:  "Smalltown",
			State: "Texas",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(30.0, -97.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Smalltown, Texas"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestReverseGeocode_VillageOnly tests reverse geocoding with village (no state)
func TestReverseGeocode_VillageOnly(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "Rural Village, Country",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{
			Village: "Rural Village",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(45.0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Rural Village"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestReverseGeocode_CountyFallback tests fallback to county when no city/town/village
func TestReverseGeocode_CountyFallback(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "Some County, State, Country",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{
			County: "Some County",
			State:  "Nevada",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(38.0, -117.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Some County, Nevada"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestReverseGeocode_DisplayNameFallback tests fallback to display name
func TestReverseGeocode_DisplayNameFallback(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "Some Remote Location",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(0.0, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Some Remote Location"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestReverseGeocode_LocationNotFound tests error when location not found
func TestReverseGeocode_LocationNotFound(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.ReverseGeocode(999.0, 999.0)
	if err == nil {
		t.Fatal("expected error for location not found, got nil")
	}

	expectedError := "location not found"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

// TestReverseGeocode_APIError tests error handling when API returns an error
func TestReverseGeocode_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.ReverseGeocode(37.7749, -122.4194)
	if err == nil {
		t.Fatal("expected error for API error, got nil")
	}
}

// TestReverseGeocode_InvalidJSON tests error handling for invalid JSON response
func TestReverseGeocode_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json {"))
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.ReverseGeocode(37.7749, -122.4194)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestReverseGeocode_CityPriorityOverTown tests that city is preferred over town
func TestReverseGeocode_CityPriorityOverTown(t *testing.T) {
	mockResponse := ReverseResponse{
		DisplayName: "Should not use this",
		Address: struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
			County  string `json:"county"`
		}{
			City:  "Big City",
			Town:  "Small Town",
			State: "California",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	result, err := client.ReverseGeocode(37.7749, -122.4194)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Big City, California"
	if result != expected {
		t.Errorf("expected %q (city should be preferred over town), got %q", expected, result)
	}
}
