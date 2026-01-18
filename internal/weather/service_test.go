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
