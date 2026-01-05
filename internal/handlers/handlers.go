package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/swelljoe/wthr.lol/internal/db"
	"github.com/swelljoe/wthr.lol/internal/weather"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	db        *db.DB
	weather   *weather.Service
	templates *template.Template
}

// New creates a new Handlers instance
func New(database *db.DB, wService *weather.Service) *Handlers {
	// Parse templates
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
	}

	return &Handlers{
		db:        database,
		weather:   wService,
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

// HandleWeatherAPI handles weather data requests
func (h *Handlers) HandleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	var lat, lon float64
	var err error

	location := r.URL.Query().Get("location")
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if location != "" {
		lat, lon, err = h.weather.Geocode(location)
		if err != nil {
			// Return a nice error fragment? Or just text for now
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("<div class='error'>Location not found: %s</div>", template.HTMLEscapeString(err.Error()))))
			return
		}
	} else if latStr != "" && lonStr != "" {
		fmt.Sscanf(latStr, "%f", &lat)
		fmt.Sscanf(lonStr, "%f", &lon)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<div class='error'>Please provide a location</div>"))
		return
	}

	wd, err := h.weather.GetWeather(lat, lon)
	if err != nil {
		log.Printf("Weather error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<div class='error'>Failed to retrieve weather data</div>"))
		return
	}

	if err := h.templates.ExecuteTemplate(w, "weather_fragment", wd); err != nil {
		log.Printf("Template error: %v", err)
	}
}
