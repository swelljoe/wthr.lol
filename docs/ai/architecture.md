# Architecture

## Overview
The application is a web server written in Go. It serves HTML pages and keeps a local cache of weather data.

## Components
- **Backend**: Go (stdlib mostly, SQLite for caching)
- **Frontend**: HTML/CSS (Vanilla), JS (Vanilla)
- **External APIs**: National Weather Service (NWS) API

## Data Flow
1. User requests page.
2. Browser requests location access.
3. User sends lat/lon to server.
4. Server checks cache for valid weather data for rounded lat/lon.
5. If miss, Server queries NWS API (Points -> Forecast).
6. Server caches result.
7. Server returns data to Frontend.

## Design Decisions
- **SQLite**: Used for caching to keep deployment simple (single file database) vs PostgreSQL.
- **Rounding Lat/Lon**: To increase cache hit rate and reduce NWS API load.
- **NWS API**: Free, reliable US weather data source.
