package weather

import "time"

// WeatherData aggregates all weather info
type WeatherData struct {
	Current   CurrentCondition `json:"current"`
	Forecast  []DailyForecast  `json:"forecast"`
	Alerts    []Alert          `json:"alerts"`
	CachedAt  time.Time        `json:"cached_at"`
	ExpiresAt time.Time        `json:"expires_at"`
	Location  string           `json:"location,omitempty"`
}

type CurrentCondition struct {
	Temperature     int    `json:"temperature"`
	TemperatureUnit string `json:"temperature_unit"`
	ShortForecast   string `json:"short_forecast"`
	Precipitation   int    `json:"precipitation_chance"`
	WindSpeed       string `json:"wind_speed"`
	WindDirection   string `json:"wind_direction"`
	Icon            string `json:"icon"`
	HighTemp        int    `json:"high_temp"`
	LowTemp         int    `json:"low_temp"`
}

type DailyForecast struct {
	Name            string `json:"name"` // e.g., "Monday"
	HighTemp        int    `json:"high_temp"`
	LowTemp         int    `json:"low_temp"`
	TemperatureUnit string `json:"temperature_unit"`
	ShortForecast   string `json:"short_forecast"`
	Icon            string `json:"icon"`
	PrecipChance    int    `json:"precip_chance"`
}

type Alert struct {
	Event       string `json:"event"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	AreaDesc    string `json:"area_desc"`
}
