package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	h := New(nil)

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
	h := New(nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}
}

func TestHandleIndexNotFound(t *testing.T) {
	h := New(nil)

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status NotFound, got %v", resp.StatusCode)
	}
}
