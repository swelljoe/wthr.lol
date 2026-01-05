package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/swelljoe/wthr.lol/internal/db"
)

const (
	placesURL = "https://www2.census.gov/geo/docs/maps-data/data/gazetteer/2023_Gazetteer/2023_Gaz_place_national.zip"
	zactasURL = "https://www2.census.gov/geo/docs/maps-data/data/gazetteer/2023_Gazetteer/2023_Gaz_zcta_national.zip"
	dataDir   = "data"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	// Initialize DB
	database, err := db.NewDB()
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer database.Close()

	// Download and process Places
	if err := processDataset(database.DB, placesURL, "places", importPlaces); err != nil {
		return fmt.Errorf("failed to process places: %w", err)
	}

	// Download and process ZCTAs
	if err := processDataset(database.DB, zactasURL, "zctas", importZCTAs); err != nil {
		return fmt.Errorf("failed to process zctas: %w", err)
	}

	return nil
}

type importFunc func(*sql.DB, io.Reader) error

func processDataset(db *sql.DB, url, name string, importer importFunc) error {
	zipPath := filepath.Join(dataDir, name+".zip")

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		fmt.Printf("Downloading %s...\n", name)
		if err := downloadFile(url, zipPath); err != nil {
			return err
		}
	} else {
		fmt.Printf("Using existing %s.zip\n", name)
	}

	fmt.Printf("Processing %s...\n", name)
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".txt") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			return importer(db, rc)
		}
	}
	return fmt.Errorf("no txt file found in %s", zipPath)
}

func downloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func importPlaces(db *sql.DB, r io.Reader) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO places (name, state, latitude, longitude) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return err
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip malformed lines
		}

		// 2023_gaz_place_national.txt format:
		// USPS(0)	GEOID(1)	ANSICODE(2)	NAME(3)	LSAD(4)	FUNCSTAT(5)	ALAND(6)	AWATER(7)	ALAND_SQMI(8)	AWATER_SQMI(9)	INTPTLAT(10)	INTPTLONG(11)

		if len(record) < 12 {
			continue
		}

		state := strings.TrimSpace(record[0])
		rawName := strings.TrimSpace(record[3])
		latStr := strings.TrimSpace(record[10])
		lonStr := strings.TrimSpace(record[11])

		// Clean name: remove " city", " town", etc.
		name := cleanPlaceName(rawName)

		// Parse and validate coordinates
		lat, lon, err := parseAndValidateCoordinates(latStr, lonStr)
		if err != nil {
			log.Printf("Error parsing coordinates for %s: %v", name, err)
			continue
		}

		_, err = stmt.Exec(name, state, lat, lon)
		if err != nil {
			log.Printf("Error inserting %s: %v", name, err)
			continue
		}
		count++
		if count%1000 == 0 {
			fmt.Printf("Imported %d places...\r", count)
		}
	}
	fmt.Printf("\nFinished importing %d places.\n", count)

	return tx.Commit()
}

func importZCTAs(db *sql.DB, r io.Reader) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO places (name, zip, state, latitude, longitude) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return err
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// 2023_gaz_zcta_national.txt format:
		// GEOID(0)	ALAND(1)	AWATER(2)	ALAND_SQMI(3)	AWATER_SQMI(4)	INTPTLAT(5)	INTPTLONG(6)

		if len(record) < 7 {
			continue
		}

		zipCode := strings.TrimSpace(record[0])
		latStr := strings.TrimSpace(record[5])
		lonStr := strings.TrimSpace(record[6])

		// Parse and validate coordinates
		lat, lon, err := parseAndValidateCoordinates(latStr, lonStr)
		if err != nil {
			log.Printf("Error parsing coordinates for ZIP %s: %v", zipCode, err)
			continue
		}

		_, err = stmt.Exec(zipCode, zipCode, "", lat, lon)
		if err != nil {
			log.Printf("Error inserting ZIP %s: %v", zipCode, err)
			continue
		}
		count++
		if count%1000 == 0 {
			fmt.Printf("Imported %d ZIPs...\r", count)
		}
	}
	fmt.Printf("\nFinished importing %d ZIPs.\n", count)

	return tx.Commit()
}

func cleanPlaceName(name string) string {
	suffixes := []string{" city", " town", " village", " CDP", " borough"}
	for _, s := range suffixes {
		if strings.HasSuffix(name, s) {
			return name[:len(name)-len(s)]
		}
	}
	return name
}

// parseAndValidateCoordinates parses and validates latitude and longitude strings
func parseAndValidateCoordinates(latStr, lonStr string) (float64, float64, error) {
	// Parse latitude
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %w", err)
	}
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude out of range: %f", lat)
	}

	// Parse longitude
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %w", err)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude out of range: %f", lon)
	}

	return lat, lon, nil
}
