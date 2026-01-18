package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client handles NWS API interactions
type Client struct {
	UserAgent  string
	HTTPClient *http.Client
}

// NewClient creates a new NWS API client
func NewClient() *Client {
	userAgent := os.Getenv("NWS_USER_AGENT")
	if userAgent == "" {
		userAgent = "wthr.lol/1.0 (contact@wthr.lol)"
	}

	return &Client{
		UserAgent: userAgent,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/geo+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NWS API error: %d %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// PointResponse represents the NWS /points/ response
type PointResponse struct {
	Properties struct {
		GridId               string `json:"gridId"`
		GridX                int    `json:"gridX"`
		GridY                int    `json:"gridY"`
		Forecast             string `json:"forecast"`
		ForecastHourly       string `json:"forecastHourly"`
		ObservationStations  string `json:"observationStations"`
		County               string `json:"county"` // URL to county
	} `json:"properties"`
}

// GetPointMetadata fetches metadata for a lat/lon
func (c *Client) GetPointMetadata(lat, lon float64) (*PointResponse, error) {
	url := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)
	data, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var pt PointResponse
	if err := json.Unmarshal(data, &pt); err != nil {
		return nil, err
	}
	return &pt, nil
}

// ForecastResponse represents the NWS /gridpoints/.../forecast response
type ForecastResponse struct {
	Properties struct {
		Periods []struct {
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
		} `json:"periods"`
	} `json:"properties"`
}

// GetForecast fetches forecast data from a provided URL
func (c *Client) GetForecast(url string) (*ForecastResponse, error) {
	data, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var fc ForecastResponse
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

// AlertsResponse represents the NWS /alerts/active response
type AlertsResponse struct {
	Features []struct {
		Properties struct {
			Event       string `json:"event"`
			Headline    string `json:"headline"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			AreaDesc    string `json:"areaDesc"`
		} `json:"properties"`
	} `json:"features"`
}

// GetAlerts fetches active alerts for a lat/lon
func (c *Client) GetAlerts(lat, lon float64) (*AlertsResponse, error) {
	url := fmt.Sprintf("https://api.weather.gov/alerts/active?point=%.4f,%.4f", lat, lon)
	data, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var al AlertsResponse
	if err := json.Unmarshal(data, &al); err != nil {
		return nil, err
	}
	return &al, nil
}

// StationFeature represents a single station feature in the GeoJSON FeatureCollection
type StationFeature struct {
	ID string `json:"id"`
}

// ObservationStationsResponse represents the /points/.../stations response as GeoJSON FeatureCollection
type ObservationStationsResponse struct {
	Features []StationFeature `json:"features"`
}

// ObservationResponse represents the /stations/.../observations/latest response
type ObservationResponse struct {
	Properties struct {
		Temperature struct {
			Value    *float64 `json:"value"`
			UnitCode string   `json:"unitCode"`
		} `json:"temperature"`
		TextDescription string `json:"textDescription"`
	} `json:"properties"`
}

// GetObservationStations fetches observation station URLs for a point
func (c *Client) GetObservationStations(stationsURL string) ([]string, error) {
	data, err := c.get(stationsURL)
	if err != nil {
		return nil, err
	}

	var resp ObservationStationsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	
	// Extract station IDs from features
	stations := make([]string, 0, len(resp.Features))
	for _, feature := range resp.Features {
		if feature.ID != "" {
			stations = append(stations, feature.ID)
		}
	}
	return stations, nil
}

// GetLatestObservation fetches the latest observation for a station URL
func (c *Client) GetLatestObservation(stationURL string) (*ObservationResponse, error) {
	obsURL := strings.TrimRight(stationURL, "/") + "/observations/latest"
	data, err := c.get(obsURL)
	if err != nil {
		return nil, err
	}

	var obs ObservationResponse
	if err := json.Unmarshal(data, &obs); err != nil {
		return nil, err
	}
	return &obs, nil
}

// GeocodeResponse represents Nominatim response
type GeocodeResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Geocode fetches coordinates for a location string using OpenStreetMap
func (c *Client) Geocode(query string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", "1")
	requestURL := baseURL + "?" + params.Encode()

	data, err := c.get(requestURL)
	if err != nil {
		return 0, 0, err
	}

	var resp GeocodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, 0, err
	}

	if len(resp) == 0 {
		return 0, 0, fmt.Errorf("location not found")
	}

	var lat, lon float64
	fmt.Sscanf(resp[0].Lat, "%f", &lat)
	fmt.Sscanf(resp[0].Lon, "%f", &lon)

	return lat, lon, nil
}

// ReverseResponse represents Nominatim reverse response
type ReverseResponse struct {
	DisplayName string `json:"display_name"`
	Address     struct {
		City    string `json:"city"`
		Town    string `json:"town"`
		Village string `json:"village"`
		State   string `json:"state"`
		County  string `json:"county"`
	} `json:"address"`
}

// ReverseGeocode fetches a human-friendly location name for given coords using OpenStreetMap
func (c *Client) ReverseGeocode(lat, lon float64) (string, error) {
	baseURL := "https://nominatim.openstreetmap.org/reverse"
	params := url.Values{}
	params.Set("format", "json")
	params.Set("lat", fmt.Sprintf("%.6f", lat))
	params.Set("lon", fmt.Sprintf("%.6f", lon))
	params.Set("zoom", "10")
	params.Set("addressdetails", "1")
	requestURL := baseURL + "?" + params.Encode()

	data, err := c.get(requestURL)
	if err != nil {
		return "", err
	}

	var resp ReverseResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	// Prefer city/town/village and append state if available
	place := ""
	if resp.Address.City != "" {
		place = resp.Address.City
	} else if resp.Address.Town != "" {
		place = resp.Address.Town
	} else if resp.Address.Village != "" {
		place = resp.Address.Village
	}
	if place != "" {
		if resp.Address.State != "" {
			return fmt.Sprintf("%s, %s", place, resp.Address.State), nil
		}
		return place, nil
	}

	if resp.Address.County != "" {
		if resp.Address.State != "" {
			return fmt.Sprintf("%s, %s", resp.Address.County, resp.Address.State), nil
		}
		return resp.Address.County, nil
	}

	if resp.DisplayName != "" {
		return resp.DisplayName, nil
	}

	return "", fmt.Errorf("location not found")
}
