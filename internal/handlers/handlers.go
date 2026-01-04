package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/swelljoe/wthr.lol/internal/db"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	db        *db.DB
	templates *template.Template
}

// New creates a new Handlers instance
func New(database *db.DB) *Handlers {
	// Parse templates
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
	}

	return &Handlers{
		db:        database,
		templates: tmpl,
	}
}

// HandleIndex handles the main page
func (h *Handlers) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if h.templates != nil {
		err := h.templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>wthr.lol</title>
</head>
<body>
	<h1>wthr.lol</h1>
	<p>Weather application - templates not loaded</p>
</body>
</html>`))
	}
}

// HandleHealth handles health check endpoint
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			status = "degraded"
		}
	} else {
		status = "no_database"
	}

	w.Write([]byte(`{"status":"` + status + `"}`))
}
