package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/swelljoe/wthr.lol/internal/db"
	"github.com/swelljoe/wthr.lol/internal/weather"
)

// Database defines the interface for database operations needed by handlers
type Database interface {
	SearchPlaces(query string) ([]db.Place, error)
	Ping() error
}

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	db        Database
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
		if _, err = fmt.Sscanf(latStr, "%f", &lat); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("<div class='error'>Invalid latitude</div>"))
			return
		}
		if _, err = fmt.Sscanf(lonStr, "%f", &lon); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("<div class='error'>Invalid longitude</div>"))
			return
		}
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

// HandleSearch performs location autocomplete
func (h *Handlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	places, err := h.db.SearchPlaces(q)
	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if places == nil {
		places = []db.Place{}
	}

	data, err := json.Marshal(places)
	if err != nil {
		log.Printf("JSON encode error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Response write error: %v", err)
	}
}
