package weather

import (
	"math"
	"testing"
)

// TestFormatHourlyLabel_ValidTime tests formatting with a valid RFC3339 timestamp
func TestFormatHourlyLabel_ValidTime(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		fallback  string
		expected  string
	}{
		{
			name:      "afternoon time",
			startTime: "2024-01-15T15:00:00-08:00",
			fallback:  "Fallback",
			expected:  "11 PM UTC",
		},
		{
			name:      "morning time",
			startTime: "2024-01-15T09:30:00Z",
			fallback:  "Fallback",
			expected:  "9 AM UTC",
		},
		{
			name:      "midnight UTC",
			startTime: "2024-01-15T00:00:00Z",
			fallback:  "Fallback",
			expected:  "12 AM UTC",
		},
		{
			name:      "noon UTC",
			startTime: "2024-01-15T12:00:00Z",
			fallback:  "Fallback",
			expected:  "12 PM UTC",
		},
		{
			name:      "with timezone offset",
			startTime: "2024-01-15T14:00:00-05:00",
			fallback:  "Fallback",
			expected:  "7 PM UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHourlyLabel(tt.startTime, tt.fallback)
			if result != tt.expected {
				t.Errorf("formatHourlyLabel(%q, %q) = %q, want %q", tt.startTime, tt.fallback, result, tt.expected)
			}
		})
	}
}

// TestFormatHourlyLabel_EmptyString tests that empty string returns fallback
func TestFormatHourlyLabel_EmptyString(t *testing.T) {
	fallback := "Original Label"
	result := formatHourlyLabel("", fallback)
	if result != fallback {
		t.Errorf("formatHourlyLabel(\"\", %q) = %q, want %q", fallback, result, fallback)
	}
}

// TestFormatHourlyLabel_InvalidFormat tests that invalid time format returns fallback
func TestFormatHourlyLabel_InvalidFormat(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		fallback  string
	}{
		{
			name:      "invalid format",
			startTime: "not-a-date",
			fallback:  "Fallback Label",
		},
		{
			name:      "wrong date format",
			startTime: "2024/01/15 15:00:00",
			fallback:  "Default",
		},
		{
			name:      "partial timestamp",
			startTime: "2024-01-15",
			fallback:  "Fallback",
		},
		{
			name:      "malformed RFC3339",
			startTime: "2024-01-15T15:00:00",
			fallback:  "Expected Label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHourlyLabel(tt.startTime, tt.fallback)
			if result != tt.fallback {
				t.Errorf("formatHourlyLabel(%q, %q) = %q, want %q (fallback)", tt.startTime, tt.fallback, result, tt.fallback)
			}
		})
	}
}

// TestObservationTemperature_NilObservation tests that nil observation returns false
func TestObservationTemperature_NilObservation(t *testing.T) {
	temp, unit, ok := observationTemperature(nil)
	if ok {
		t.Errorf("observationTemperature(nil) returned ok=true, want false")
	}
	if temp != 0 {
		t.Errorf("observationTemperature(nil) returned temp=%d, want 0", temp)
	}
	if unit != "" {
		t.Errorf("observationTemperature(nil) returned unit=%q, want empty string", unit)
	}
}

// TestObservationTemperature_NilTemperatureValue tests that nil temperature value returns false
func TestObservationTemperature_NilTemperatureValue(t *testing.T) {
	obs := createMockObservation(nil, "wmoUnit:degC", "Test")
	temp, unit, ok := observationTemperature(&obs)
	if ok {
		t.Errorf("observationTemperature with nil temperature returned ok=true, want false")
	}
	if temp != 0 {
		t.Errorf("observationTemperature with nil temperature returned temp=%d, want 0", temp)
	}
	if unit != "" {
		t.Errorf("observationTemperature with nil temperature returned unit=%q, want empty string", unit)
	}
}

// TestObservationTemperature_NaNValue tests that NaN temperature value returns false
func TestObservationTemperature_NaNValue(t *testing.T) {
	nanValue := math.NaN()
	obs := createMockObservation(&nanValue, "wmoUnit:degC", "Test")
	temp, unit, ok := observationTemperature(&obs)
	if ok {
		t.Errorf("observationTemperature with NaN temperature returned ok=true, want false")
	}
	if temp != 0 {
		t.Errorf("observationTemperature with NaN temperature returned temp=%d, want 0", temp)
	}
	if unit != "" {
		t.Errorf("observationTemperature with NaN temperature returned unit=%q, want empty string", unit)
	}
}

// TestObservationTemperature_CelsiusConversion tests Celsius to Fahrenheit conversion
func TestObservationTemperature_CelsiusConversion(t *testing.T) {
	tests := []struct {
		name         string
		celsius      float64
		unitCode     string
		expectedTemp int
		expectedUnit string
	}{
		{
			name:         "freezing point",
			celsius:      0.0,
			unitCode:     "degC",
			expectedTemp: 32,
			expectedUnit: "F",
		},
		{
			name:         "room temperature",
			celsius:      20.0,
			unitCode:     "degC",
			expectedTemp: 68,
			expectedUnit: "F",
		},
		{
			name:         "body temperature",
			celsius:      37.0,
			unitCode:     "degC",
			expectedTemp: 99,
			expectedUnit: "F",
		},
		{
			name:         "hot summer day",
			celsius:      35.0,
			unitCode:     "degC",
			expectedTemp: 95,
			expectedUnit: "F",
		},
		{
			name:         "below freezing",
			celsius:      -10.0,
			unitCode:     "degC",
			expectedTemp: 14,
			expectedUnit: "F",
		},
		{
			name:         "compound unit code",
			celsius:      25.0,
			unitCode:     "wmoUnit:degC",
			expectedTemp: 77,
			expectedUnit: "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := createMockObservation(&tt.celsius, tt.unitCode, "Test")
			temp, unit, ok := observationTemperature(&obs)
			if !ok {
				t.Errorf("observationTemperature returned ok=false, want true")
			}
			if temp != tt.expectedTemp {
				t.Errorf("observationTemperature(%f°C) = %d°F, want %d°F", tt.celsius, temp, tt.expectedTemp)
			}
			if unit != tt.expectedUnit {
				t.Errorf("observationTemperature returned unit=%q, want %q", unit, tt.expectedUnit)
			}
		})
	}
}

// TestObservationTemperature_FahrenheitPassthrough tests Fahrenheit values pass through
func TestObservationTemperature_FahrenheitPassthrough(t *testing.T) {
	tests := []struct {
		name         string
		fahrenheit   float64
		unitCode     string
		expectedTemp int
		expectedUnit string
	}{
		{
			name:         "freezing point",
			fahrenheit:   32.0,
			unitCode:     "degF",
			expectedTemp: 32,
			expectedUnit: "F",
		},
		{
			name:         "room temperature",
			fahrenheit:   68.5,
			unitCode:     "degF",
			expectedTemp: 69,
			expectedUnit: "F",
		},
		{
			name:         "hot day",
			fahrenheit:   95.3,
			unitCode:     "degF",
			expectedTemp: 95,
			expectedUnit: "F",
		},
		{
			name:         "compound unit code",
			fahrenheit:   75.0,
			unitCode:     "wmoUnit:degF",
			expectedTemp: 75,
			expectedUnit: "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := createMockObservation(&tt.fahrenheit, tt.unitCode, "Test")
			temp, unit, ok := observationTemperature(&obs)
			if !ok {
				t.Errorf("observationTemperature returned ok=false, want true")
			}
			if temp != tt.expectedTemp {
				t.Errorf("observationTemperature(%f°F) = %d°F, want %d°F", tt.fahrenheit, temp, tt.expectedTemp)
			}
			if unit != tt.expectedUnit {
				t.Errorf("observationTemperature returned unit=%q, want %q", unit, tt.expectedUnit)
			}
		})
	}
}

// TestObservationTemperature_UnrecognizedUnitCode tests fallback for unrecognized units
func TestObservationTemperature_UnrecognizedUnitCode(t *testing.T) {
	tests := []struct {
		name         string
		temperature  float64
		unitCode     string
		expectedTemp int
		expectedUnit string
	}{
		{
			name:         "Kelvin",
			temperature:  273.15,
			unitCode:     "K",
			expectedTemp: 273,
			expectedUnit: "K",
		},
		{
			name:         "unknown unit",
			temperature:  25.0,
			unitCode:     "unknown",
			expectedTemp: 25,
			expectedUnit: "unknown",
		},
		{
			name:         "compound unknown unit",
			temperature:  30.0,
			unitCode:     "wmoUnit:degK",
			expectedTemp: 30,
			expectedUnit: "degK",
		},
		{
			name:         "empty unit code",
			temperature:  20.0,
			unitCode:     "",
			expectedTemp: 20,
			expectedUnit: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := createMockObservation(&tt.temperature, tt.unitCode, "Test")
			temp, unit, ok := observationTemperature(&obs)
			if !ok {
				t.Errorf("observationTemperature returned ok=false, want true")
			}
			if temp != tt.expectedTemp {
				t.Errorf("observationTemperature(%f %s) = %d, want %d", tt.temperature, tt.unitCode, temp, tt.expectedTemp)
			}
			if unit != tt.expectedUnit {
				t.Errorf("observationTemperature returned unit=%q, want %q", unit, tt.expectedUnit)
			}
		})
	}
}

// TestObservationTemperature_Rounding tests proper rounding of temperature values
func TestObservationTemperature_Rounding(t *testing.T) {
	tests := []struct {
		name         string
		temperature  float64
		unitCode     string
		expectedTemp int
	}{
		{
			name:         "round down",
			temperature:  20.4,
			unitCode:     "degC",
			expectedTemp: 69, // 68.72 rounds to 69
		},
		{
			name:         "round up",
			temperature:  20.6,
			unitCode:     "degC",
			expectedTemp: 69, // 69.08 rounds to 69
		},
		{
			name:         "exact half rounds up",
			temperature:  68.5,
			unitCode:     "degF",
			expectedTemp: 69, // math.Round rounds half away from zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := createMockObservation(&tt.temperature, tt.unitCode, "Test")
			temp, _, ok := observationTemperature(&obs)
			if !ok {
				t.Errorf("observationTemperature returned ok=false, want true")
			}
			if temp != tt.expectedTemp {
				t.Errorf("observationTemperature(%f %s) = %d, want %d", tt.temperature, tt.unitCode, temp, tt.expectedTemp)
			}
		})
	}
}

// Helper functions for creating test data

func createMockForecastResponse(periods []struct {
	Name        string
	StartTime   string
	IsDaytime   bool
	Temperature int
	Unit        string
	WindSpeed   string
	WindDir     string
	Icon        string
	ShortFcst   string
	PrecipValue int
}) *ForecastResponse {
	fc := &ForecastResponse{}
	for _, p := range periods {
		fc.Properties.Periods = append(fc.Properties.Periods, struct {
			Name                       string `json:"name"`
			StartTime                  string `json:"startTime"`
			IsDaytime                  bool   `json:"isDaytime"`
			Temperature                int    `json:"temperature"`
			TemperatureUnit            string `json:"temperatureUnit"`
			ProbabilityOfPrecipitation struct {
				Value int `json:"value"`
			} `json:"probabilityOfPrecipitation"`
			WindSpeed        string `json:"windSpeed"`
			WindDirection    string `json:"windDirection"`
			Icon             string `json:"icon"`
			ShortForecast    string `json:"shortForecast"`
			DetailedForecast string `json:"detailedForecast"`
		}{
			Name:            p.Name,
			StartTime:       p.StartTime,
			IsDaytime:       p.IsDaytime,
			Temperature:     p.Temperature,
			TemperatureUnit: p.Unit,
			WindSpeed:       p.WindSpeed,
			WindDirection:   p.WindDir,
			Icon:            p.Icon,
			ShortForecast:   p.ShortFcst,
			ProbabilityOfPrecipitation: struct {
				Value int `json:"value"`
			}{Value: p.PrecipValue},
		})
	}
	return fc
}

func createMockAlertsResponse() *AlertsResponse {
	return &AlertsResponse{
		Features: []struct {
			Properties struct {
				Event       string `json:"event"`
				Headline    string `json:"headline"`
				Description string `json:"description"`
				Severity    string `json:"severity"`
				AreaDesc    string `json:"areaDesc"`
			} `json:"properties"`
		}{},
	}
}

// TestTransform_HourlyNil tests transform when hourly forecast (hc) is nil
func TestTransform_HourlyNil(t *testing.T) {
	fc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Today", IsDaytime: true, Temperature: 75, Unit: "F", WindSpeed: "10 mph", WindDir: "N", Icon: "https://api.weather.gov/icons/land/day/sunny", ShortFcst: "Sunny", PrecipValue: 10},
	})
	al := createMockAlertsResponse()
	tempValue := 72.0
	obs := createMockObservation(&tempValue, "wmoUnit:degC", "Clear")

	wd, err := transform(fc, nil, al, &obs)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// When hc is nil, Hourly should be empty
	if len(wd.Hourly) != 0 {
		t.Errorf("Expected Hourly to be empty when hc is nil, got %d items", len(wd.Hourly))
	}

	// Current should be populated from fc (forecast) fallback
	if wd.Current.Temperature != 162 { // 72°C = ~162°F
		t.Errorf("Expected Current.Temperature to be 162 from observation, got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}
}

// TestTransform_ObservationNil tests transform when observation (obs) is nil
func TestTransform_ObservationNil(t *testing.T) {
	hc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Now", StartTime: "2024-01-15T15:00:00Z", IsDaytime: true, Temperature: 68, Unit: "F", WindSpeed: "5 mph", WindDir: "NE", Icon: "https://api.weather.gov/icons/land/day/cloudy", ShortFcst: "Cloudy", PrecipValue: 20},
		{Name: "+1h", StartTime: "2024-01-15T16:00:00Z", IsDaytime: true, Temperature: 69, Unit: "F", WindSpeed: "5 mph", WindDir: "NE", Icon: "https://api.weather.gov/icons/land/day/cloudy", ShortFcst: "Cloudy", PrecipValue: 20},
	})
	fc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Today", IsDaytime: true, Temperature: 75, Unit: "F", WindSpeed: "10 mph", WindDir: "N", Icon: "https://api.weather.gov/icons/land/day/sunny", ShortFcst: "Sunny", PrecipValue: 10},
	})
	al := createMockAlertsResponse()

	wd, err := transform(fc, hc, al, nil)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Hourly should be populated from hc
	if len(wd.Hourly) != 2 {
		t.Errorf("Expected Hourly to have 2 items, got %d", len(wd.Hourly))
	}

	// Current should be populated from hc, not overridden by observation
	if wd.Current.Temperature != 68 {
		t.Errorf("Expected Current.Temperature to be 68 from hc, got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}
}

// TestTransform_BothHourlyAndObservationPresent tests transform when both hc and obs are present
func TestTransform_BothHourlyAndObservationPresent(t *testing.T) {
	hc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Now", StartTime: "2024-01-15T15:00:00Z", IsDaytime: true, Temperature: 68, Unit: "F", WindSpeed: "5 mph", WindDir: "NE", Icon: "https://api.weather.gov/icons/land/day/cloudy", ShortFcst: "Cloudy", PrecipValue: 20},
		{Name: "+1h", StartTime: "2024-01-15T16:00:00Z", IsDaytime: true, Temperature: 69, Unit: "F", WindSpeed: "6 mph", WindDir: "NE", Icon: "https://api.weather.gov/icons/land/day/cloudy", ShortFcst: "Cloudy", PrecipValue: 25},
		{Name: "+2h", StartTime: "2024-01-15T17:00:00Z", IsDaytime: true, Temperature: 70, Unit: "F", WindSpeed: "7 mph", WindDir: "E", Icon: "https://api.weather.gov/icons/land/day/rain", ShortFcst: "Light Rain", PrecipValue: 30},
	})
	fc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Today", IsDaytime: true, Temperature: 75, Unit: "F", WindSpeed: "10 mph", WindDir: "N", Icon: "https://api.weather.gov/icons/land/day/sunny", ShortFcst: "Sunny", PrecipValue: 10},
	})
	al := createMockAlertsResponse()
	tempValue := 20.0 // 20°C = 68°F
	obs := createMockObservation(&tempValue, "wmoUnit:degC", "Clear")

	wd, err := transform(fc, hc, al, &obs)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Hourly should be populated from hc with first 5 periods
	if len(wd.Hourly) != 3 {
		t.Errorf("Expected Hourly to have 3 items, got %d", len(wd.Hourly))
	}

	// Verify hourly data
	if wd.Hourly[0].Temperature != 68 {
		t.Errorf("Expected Hourly[0].Temperature to be 68, got %d", wd.Hourly[0].Temperature)
	}
	if wd.Hourly[0].Name != "3 PM UTC" {
		t.Errorf("Expected Hourly[0].Name to be '3 PM UTC', got %s", wd.Hourly[0].Name)
	}

	// Current temperature should be overridden by observation
	if wd.Current.Temperature != 68 {
		t.Errorf("Expected Current.Temperature to be 68 (from observation 20°C), got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}

	// Other current fields should still come from hc
	if wd.Current.ShortForecast != "Cloudy" {
		t.Errorf("Expected Current.ShortForecast to be 'Cloudy' from hc, got %s", wd.Current.ShortForecast)
	}
	if wd.Current.WindSpeed != "5 mph" {
		t.Errorf("Expected Current.WindSpeed to be '5 mph' from hc, got %s", wd.Current.WindSpeed)
	}
}

// TestTransform_CurrentFromHourly tests that current condition is populated from hourly when available
func TestTransform_CurrentFromHourly(t *testing.T) {
	hc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Now", StartTime: "2024-01-15T15:00:00Z", IsDaytime: true, Temperature: 65, Unit: "F", WindSpeed: "8 mph", WindDir: "SW", Icon: "https://api.weather.gov/icons/land/day/partly-cloudy", ShortFcst: "Partly Cloudy", PrecipValue: 15},
	})
	al := createMockAlertsResponse()

	wd, err := transform(nil, hc, al, nil)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Current should be populated from hc (first period)
	if wd.Current.Temperature != 65 {
		t.Errorf("Expected Current.Temperature to be 65 from hc, got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}
	if wd.Current.ShortForecast != "Partly Cloudy" {
		t.Errorf("Expected Current.ShortForecast to be 'Partly Cloudy', got %s", wd.Current.ShortForecast)
	}
	if wd.Current.WindSpeed != "8 mph" {
		t.Errorf("Expected Current.WindSpeed to be '8 mph', got %s", wd.Current.WindSpeed)
	}
	if wd.Current.WindDirection != "SW" {
		t.Errorf("Expected Current.WindDirection to be 'SW', got %s", wd.Current.WindDirection)
	}
	if wd.Current.Precipitation != 15 {
		t.Errorf("Expected Current.Precipitation to be 15, got %d", wd.Current.Precipitation)
	}
}

// TestTransform_CurrentFallbackToForecast tests that current condition falls back to forecast when hourly is not available
func TestTransform_CurrentFallbackToForecast(t *testing.T) {
	fc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Today", IsDaytime: true, Temperature: 72, Unit: "F", WindSpeed: "12 mph", WindDir: "NW", Icon: "https://api.weather.gov/icons/land/day/sunny", ShortFcst: "Sunny", PrecipValue: 5},
	})
	al := createMockAlertsResponse()

	wd, err := transform(fc, nil, al, nil)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Current should be populated from fc (forecast) since hc is nil
	if wd.Current.Temperature != 72 {
		t.Errorf("Expected Current.Temperature to be 72 from fc, got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}
	if wd.Current.ShortForecast != "Sunny" {
		t.Errorf("Expected Current.ShortForecast to be 'Sunny', got %s", wd.Current.ShortForecast)
	}
	if wd.Current.WindSpeed != "12 mph" {
		t.Errorf("Expected Current.WindSpeed to be '12 mph', got %s", wd.Current.WindSpeed)
	}
}

// TestTransform_HourlyLimitsFiveItems tests that hourly forecast is limited to 5 items
func TestTransform_HourlyLimitsFiveItems(t *testing.T) {
	hc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "h1", StartTime: "2024-01-15T15:00:00Z", IsDaytime: true, Temperature: 65, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon1", ShortFcst: "F1", PrecipValue: 10},
		{Name: "h2", StartTime: "2024-01-15T16:00:00Z", IsDaytime: true, Temperature: 66, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon2", ShortFcst: "F2", PrecipValue: 15},
		{Name: "h3", StartTime: "2024-01-15T17:00:00Z", IsDaytime: true, Temperature: 67, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon3", ShortFcst: "F3", PrecipValue: 20},
		{Name: "h4", StartTime: "2024-01-15T18:00:00Z", IsDaytime: true, Temperature: 68, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon4", ShortFcst: "F4", PrecipValue: 25},
		{Name: "h5", StartTime: "2024-01-15T19:00:00Z", IsDaytime: true, Temperature: 69, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon5", ShortFcst: "F5", PrecipValue: 30},
		{Name: "h6", StartTime: "2024-01-15T20:00:00Z", IsDaytime: false, Temperature: 64, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon6", ShortFcst: "F6", PrecipValue: 35},
		{Name: "h7", StartTime: "2024-01-15T21:00:00Z", IsDaytime: false, Temperature: 63, Unit: "F", WindSpeed: "5 mph", WindDir: "N", Icon: "icon7", ShortFcst: "F7", PrecipValue: 40},
	})
	al := createMockAlertsResponse()

	wd, err := transform(nil, hc, al, nil)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Should only have first 5 hourly items
	if len(wd.Hourly) != 5 {
		t.Errorf("Expected Hourly to be limited to 5 items, got %d", len(wd.Hourly))
	}

	// Verify we got the first 5, not the last 5
	if wd.Hourly[4].Temperature != 69 {
		t.Errorf("Expected 5th hourly item to have temperature 69, got %d", wd.Hourly[4].Temperature)
	}
}

// TestTransform_ObservationOverridesCurrentTemperature tests that observation temperature overrides current
func TestTransform_ObservationOverridesCurrentTemperature(t *testing.T) {
	hc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Now", StartTime: "2024-01-15T15:00:00Z", IsDaytime: true, Temperature: 70, Unit: "F", WindSpeed: "10 mph", WindDir: "E", Icon: "https://api.weather.gov/icons/land/day/cloudy", ShortFcst: "Cloudy", PrecipValue: 20},
	})
	fc := createMockForecastResponse([]struct {
		Name        string
		StartTime   string
		IsDaytime   bool
		Temperature int
		Unit        string
		WindSpeed   string
		WindDir     string
		Icon        string
		ShortFcst   string
		PrecipValue int
	}{
		{Name: "Today", IsDaytime: true, Temperature: 75, Unit: "F", WindSpeed: "10 mph", WindDir: "N", Icon: "https://api.weather.gov/icons/land/day/sunny", ShortFcst: "Sunny", PrecipValue: 10},
	})
	al := createMockAlertsResponse()
	tempValue := 25.0 // 25°C = 77°F
	obs := createMockObservation(&tempValue, "wmoUnit:degC", "Mostly Sunny")

	wd, err := transform(fc, hc, al, &obs)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Temperature and unit should be from observation
	if wd.Current.Temperature != 77 {
		t.Errorf("Expected Current.Temperature to be 77 (from observation 25°C), got %d", wd.Current.Temperature)
	}
	if wd.Current.TemperatureUnit != "F" {
		t.Errorf("Expected Current.TemperatureUnit to be F, got %s", wd.Current.TemperatureUnit)
	}

	// Other fields should still come from hc
	if wd.Current.ShortForecast != "Cloudy" {
		t.Errorf("Expected Current.ShortForecast to remain 'Cloudy' from hc, got %s", wd.Current.ShortForecast)
	}
	if wd.Current.WindSpeed != "10 mph" {
		t.Errorf("Expected Current.WindSpeed to remain '10 mph' from hc, got %s", wd.Current.WindSpeed)
	}
}

// TestTransform_ObservationSetsHighLowWhenNoForecasts tests observation sets high/low when no forecast data
func TestTransform_ObservationSetsHighLowWhenNoForecasts(t *testing.T) {
	al := createMockAlertsResponse()
	tempValue := 22.0 // 22°C = ~72°F
	obs := createMockObservation(&tempValue, "wmoUnit:degC", "Fair")

	wd, err := transform(nil, nil, al, &obs)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// When both hc and fc are nil, observation should set high/low to avoid misleading 0° values
	if wd.Current.HighTemp != 72 {
		t.Errorf("Expected Current.HighTemp to be 72 from observation, got %d", wd.Current.HighTemp)
	}
	if wd.Current.LowTemp != 72 {
		t.Errorf("Expected Current.LowTemp to be 72 from observation, got %d", wd.Current.LowTemp)
	}
}
