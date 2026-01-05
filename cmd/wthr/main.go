package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/swelljoe/wthr.lol/internal/db"
	"github.com/swelljoe/wthr.lol/internal/handlers"
	"github.com/swelljoe/wthr.lol/internal/weather"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection
	database, err := db.NewDB()
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Continuing without database connection...")
	} else {
		defer database.Close()
		log.Println("Database connected successfully")
	}

	// Initialize services
	wService := weather.NewService(database)

	// Setup routes
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve service worker from root for scope
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "static/sw.js")
	})

	// Serve .well-known for Digital Asset Links
	mux.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir("static/.well-known"))))

	// Setup handlers
	h := handlers.New(database, wService)
	mux.HandleFunc("/", h.HandleIndex)
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/api/weather", h.HandleWeatherAPI)
	mux.HandleFunc("/api/search", h.HandleSearch)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
