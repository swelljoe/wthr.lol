package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockRoundTripper is a custom RoundTripper for testing HTTP clients.
// It intercepts HTTP requests and passes them to a test handler, allowing
// us to mock API responses without starting a real HTTP server or modifying
// the Client's get method. This approach keeps tests isolated and fast.
type mockRoundTripper struct {
	handler http.Handler
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	resp := rec.Result()
	return resp, nil
}

// createMockObservation is a helper function to create ObservationResponse instances
// for testing, reducing code duplication.
func createMockObservation(tempValue *float64, unitCode, description string) ObservationResponse {
	return ObservationResponse{
		Properties: struct {
			Temperature struct {
				Value    *float64 `json:"value"`
				UnitCode string   `json:"unitCode"`
			} `json:"temperature"`
			TextDescription string `json:"textDescription"`
		}{
			Temperature: struct {
				Value    *float64 `json:"value"`
				UnitCode string   `json:"unitCode"`
			}{
				Value:    tempValue,
				UnitCode: unitCode,
			},
			TextDescription: description,
		},
	}
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

// TestGetObservationStations_Success tests successful retrieval of observation stations
func TestGetObservationStations_Success(t *testing.T) {
	mockResponse := ObservationStationsResponse{
		ObservationStations: []string{
			"https://api.weather.gov/stations/KSFO",
			"https://api.weather.gov/stations/KOAK",
			"https://api.weather.gov/stations/KHWD",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	stations, err := client.GetObservationStations("https://api.weather.gov/gridpoints/MTR/85,105/stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stations) != 3 {
		t.Errorf("expected 3 stations, got %d", len(stations))
	}

	expectedFirst := "https://api.weather.gov/stations/KSFO"
	if stations[0] != expectedFirst {
		t.Errorf("expected first station %q, got %q", expectedFirst, stations[0])
	}
}

// TestGetObservationStations_EmptyList tests handling of empty station list
func TestGetObservationStations_EmptyList(t *testing.T) {
	mockResponse := ObservationStationsResponse{
		ObservationStations: []string{},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	stations, err := client.GetObservationStations("https://api.weather.gov/gridpoints/MTR/85,105/stations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stations) != 0 {
		t.Errorf("expected 0 stations, got %d", len(stations))
	}
}

// TestGetObservationStations_APIError tests error handling when API returns an error
func TestGetObservationStations_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.GetObservationStations("https://api.weather.gov/gridpoints/MTR/85,105/stations")
	if err == nil {
		t.Fatal("expected error for API error, got nil")
	}
}

// TestGetObservationStations_InvalidJSON tests error handling for invalid JSON response
func TestGetObservationStations_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		w.Write([]byte("invalid json {"))
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.GetObservationStations("https://api.weather.gov/gridpoints/MTR/85,105/stations")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestGetLatestObservation_Success tests successful retrieval of latest observation
func TestGetLatestObservation_Success(t *testing.T) {
	tempValue := 20.5
	mockResponse := createMockObservation(&tempValue, "wmoUnit:degC", "Partly Cloudy")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the URL includes /observations/latest
		if !strings.Contains(r.URL.Path, "/observations/latest") {
			t.Errorf("expected URL to contain /observations/latest, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/geo+json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	obs, err := client.GetLatestObservation("https://api.weather.gov/stations/KSFO")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if obs.Properties.Temperature.Value == nil {
		t.Fatal("expected temperature value, got nil")
	}

	if *obs.Properties.Temperature.Value != tempValue {
		t.Errorf("expected temperature %v, got %v", tempValue, *obs.Properties.Temperature.Value)
	}

	if obs.Properties.Temperature.UnitCode != "wmoUnit:degC" {
		t.Errorf("expected unit code %q, got %q", "wmoUnit:degC", obs.Properties.Temperature.UnitCode)
	}

	if obs.Properties.TextDescription != "Partly Cloudy" {
		t.Errorf("expected description %q, got %q", "Partly Cloudy", obs.Properties.TextDescription)
	}
}

// TestGetLatestObservation_NullTemperature tests handling of null temperature value
func TestGetLatestObservation_NullTemperature(t *testing.T) {
	mockResponse := createMockObservation(nil, "wmoUnit:degC", "Clear")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	obs, err := client.GetLatestObservation("https://api.weather.gov/stations/KSFO")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if obs.Properties.Temperature.Value != nil {
		t.Errorf("expected nil temperature value, got %v", *obs.Properties.Temperature.Value)
	}

	if obs.Properties.TextDescription != "Clear" {
		t.Errorf("expected description %q, got %q", "Clear", obs.Properties.TextDescription)
	}
}

// TestGetLatestObservation_URLTrimming tests that trailing slashes are properly handled
func TestGetLatestObservation_URLTrimming(t *testing.T) {
	tempValue := 15.0
	mockResponse := createMockObservation(&tempValue, "wmoUnit:degC", "Sunny")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the URL doesn't have double slashes before /observations
		if strings.Contains(r.URL.Path, "//observations") {
			t.Error("URL should not contain double slashes")
		}

		w.Header().Set("Content-Type", "application/geo+json")
		json.NewEncoder(w).Encode(mockResponse)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	// Test with trailing slash
	obs, err := client.GetLatestObservation("https://api.weather.gov/stations/KSFO/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if obs.Properties.Temperature.Value == nil {
		t.Fatal("expected temperature value, got nil")
	}

	if *obs.Properties.Temperature.Value != tempValue {
		t.Errorf("expected temperature %v, got %v", tempValue, *obs.Properties.Temperature.Value)
	}
}

// TestGetLatestObservation_APIError tests error handling when API returns an error
func TestGetLatestObservation_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.GetLatestObservation("https://api.weather.gov/stations/INVALID")
	if err == nil {
		t.Fatal("expected error for API error, got nil")
	}
}

// TestGetLatestObservation_InvalidJSON tests error handling for invalid JSON response
func TestGetLatestObservation_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		w.Write([]byte("invalid json {"))
	})

	client := &Client{
		UserAgent: "test-agent",
		HTTPClient: &http.Client{
			Transport: &mockRoundTripper{handler: handler},
		},
	}

	_, err := client.GetLatestObservation("https://api.weather.gov/stations/KSFO")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
