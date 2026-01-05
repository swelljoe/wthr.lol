package weather

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/swelljoe/wthr.lol/internal/db"
)

// Service handles weather business logic and caching
type Service struct {
	client *Client
	db     *db.DB
}

// NewService creates a new weather service
func NewService(db *db.DB) *Service {
	return &Service{
		client: NewClient(),
		db:     db,
	}
}

// GetWeather returns weather data for a given location, utilizing caching
func (s *Service) GetWeather(lat, lon float64) (*WeatherData, error) {
	// 1. Round coordinates to 2 decimal places (approx 1.1km precision)
	// This reduces the number of unique cache entries and API hits
	const precision = 100.0
	rLat := math.Round(lat*precision) / precision
	rLon := math.Round(lon*precision) / precision

	// 2. Check cache
	cached, err := s.db.GetCachedWeather(rLat, rLon)
	if err != nil {
		log.Printf("Cache error: %v", err)
		// Proceed to fetch fresh data on cache error
	}

	if cached != nil {
		var wd WeatherData
		if err := json.Unmarshal([]byte(cached.Data), &wd); err == nil {
			wd.CachedAt = time.Now() // Indicate it's fresh from cache? Or original time?
			// Ideally we want to know when it expires.
			wd.ExpiresAt = cached.ExpiresAt
			return &wd, nil
		} else {
			log.Printf("Cache unmarshal error: %v", err)
		}
	}

	// 3. Fetch fresh data
	wd, err := s.fetchFreshWeather(rLat, rLon)
	if err != nil {
		return nil, err
	}

	// 4. Update cache
	data, err := json.Marshal(wd)
	if err == nil {
		if err := s.db.SetCachedWeather(rLat, rLon, string(data), 1*time.Hour); err != nil {
			log.Printf("Failed to update cache: %v", err)
		}
	}

	return wd, nil
}

func (s *Service) fetchFreshWeather(lat, lon float64) (*WeatherData, error) {
	// A. Get Point Metadata to find Forecast URL
	pt, err := s.client.GetPointMetadata(lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to get point metadata: %w", err)
	}

	// B. Get Forecast
	fc, err := s.client.GetForecast(pt.Properties.Forecast)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	// C. Get Alerts
	al, err := s.client.GetAlerts(lat, lon)
	if err != nil {
		// Log error but don't fail entire request?
		// User wants "Display severe weather alerts... if any".
		// If fails, we assume no alerts or partial failure.
		log.Printf("Failed to get alerts: %v", err)
		al = &AlertsResponse{} // Empty alerts
	}

	// D. Transform to internal structure
	return transform(fc, al)
}

func transform(fc *ForecastResponse, al *AlertsResponse) (*WeatherData, error) {
	wd := &WeatherData{
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Forecast:  make([]DailyForecast, 0),
		Alerts:    make([]Alert, 0),
	}

	periods := fc.Properties.Periods
	if len(periods) > 0 {
		// Current Weather (Use first period)
		curr := periods[0]
		wd.Current = CurrentCondition{
			Temperature:     curr.Temperature,
			TemperatureUnit: curr.TemperatureUnit,
			ShortForecast:   curr.ShortForecast,
			Precipitation:   curr.ProbabilityOfPrecipitation.Value,
			WindSpeed:       curr.WindSpeed,
			WindDirection:   curr.WindDirection,
			Icon:            mapIcon(curr.Icon, curr.IsDaytime),
		}

		// Calculate High/Low for "Today" (Current Day)
		high := curr.Temperature
		low := curr.Temperature
		if len(periods) > 1 {
			next := periods[1]
			if next.Temperature > high {
				high = next.Temperature
			}
			if next.Temperature < low {
				low = next.Temperature
			}
		}
		wd.Current.HighTemp = high
		wd.Current.LowTemp = low

		// Process Forecast
		processedDays := 0
		i := 0
		for i < len(periods) {
			p := periods[i]

			// Create a new day entry
			day := DailyForecast{
				Name:            p.Name,
				TemperatureUnit: p.TemperatureUnit,
				Icon:            mapIcon(p.Icon, p.IsDaytime),
				ShortForecast:   p.ShortForecast,
				PrecipChance:    p.ProbabilityOfPrecipitation.Value,
				HighTemp:        p.Temperature,
				LowTemp:         p.Temperature,
			}

			// Is this a "Day" part or "Night" part?
			if p.IsDaytime {
				day.HighTemp = p.Temperature
				// Look ahead for night
				if i+1 < len(periods) {
					next := periods[i+1]
					if !next.IsDaytime {
						day.LowTemp = next.Temperature
						// maximize precip chance?
						if next.ProbabilityOfPrecipitation.Value > day.PrecipChance {
							day.PrecipChance = next.ProbabilityOfPrecipitation.Value
						}
						i++ // Consume next period
					}
				}
			} else {
				// Standalone Night
				day.LowTemp = p.Temperature
				day.HighTemp = p.Temperature
			}

			wd.Forecast = append(wd.Forecast, day)
			processedDays++
			i++

			if processedDays >= 5 {
				break
			}
		}
	}

	// Alerts
	for _, feature := range al.Features {
		wd.Alerts = append(wd.Alerts, Alert{
			Event:       feature.Properties.Event,
			Headline:    feature.Properties.Headline,
			Description: feature.Properties.Description,
			Severity:    feature.Properties.Severity,
			AreaDesc:    feature.Properties.AreaDesc,
		})
	}

	return wd, nil
}

// mapIcon maps NWS icon URL or forecast description to Material Symbol name
func mapIcon(iconURL string, isDaytime bool) string {
	// Basic mapping based on keywords
	if contains(iconURL, "/skc") || contains(iconURL, "/few") {
		if !isDaytime {
			return "clear_night"
		}
		return "sunny" // Clear/Sunny
	}
	if contains(iconURL, "/sct") || contains(iconURL, "/bkn") {
		if !isDaytime {
			return "partly_cloudy_night"
		}
		return "partly_cloudy_day"
	}
	if contains(iconURL, "/ovc") {
		return "cloud" // Overcast
	}
	if contains(iconURL, "/rain") || contains(iconURL, "/showers") {
		return "rainy"
	}
	if contains(iconURL, "/tsra") {
		return "thunderstorm"
	}
	if contains(iconURL, "/snow") {
		return "weather_snowy"
	}
	if contains(iconURL, "/fog") {
		return "foggy"
	}
	if contains(iconURL, "/wind") {
		return "air"
	}

	return "thermostat"
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Geocode resolves a location string to coordinates
func (s *Service) Geocode(query string) (float64, float64, error) {
	return s.client.Geocode(query)
}
