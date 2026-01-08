package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// TestStaticCountriesServed verifies that the static file server serves
// the countries JSON file at /static/countries.json and that it contains
// expected content (e.g., includes "United States").
func TestStaticCountriesServed(t *testing.T) {
	// Serve files from the repo's static directory (relative to cmd/wthr)
	staticDir := filepath.Join("..", "..", "static")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/countries.json")
	if err != nil {
		t.Fatalf("failed to GET countries.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("unexpected Content-Type: %s", ct)
	}

	var countries []string
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		t.Fatalf("failed to decode countries.json: %v", err)
	}

	if len(countries) == 0 {
		t.Fatalf("expected at least one country in countries.json")
	}

	found := false
	for _, c := range countries {
		if c == "United States" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("'United States' not found in countries.json")
	}
}
