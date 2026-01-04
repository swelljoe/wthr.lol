# wthr.lol

Because I'm tired of ad-filled weather sites and apps

A simple, ad-free weather application built with Go, PostgreSQL, HTML, Datastar, and Pico.css.

## Features

- **Clean UI**: Modern, responsive design using Pico.css
- **No Ads**: Just weather information, nothing else
- **Privacy-focused**: No tracking or data collection
- **Fast**: Built with Go for optimal performance

## Tech Stack

- **Backend**: Go + PostgreSQL
- **Frontend**: HTML + Datastar
- **Styling**: Pico.css

## Project Structure

```
wthr.lol/
├── cmd/
│   └── wthr/           # Main application entry point
├── internal/
│   ├── db/             # Database connection and queries
│   └── handlers/       # HTTP request handlers
├── static/             # Static assets (CSS, JS, images)
├── templates/          # HTML templates
├── Makefile            # Build automation
└── go.mod              # Go module dependencies
```

## Getting Started

### Prerequisites

- Go 1.24 or later
- PostgreSQL (optional for development)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/swelljoe/wthr.lol.git
cd wthr.lol
```

2. Install dependencies:
```bash
make mod-download
```

### Configuration

The application can be configured using environment variables:

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- Or individual database settings:
  - `DB_HOST`: Database host (default: localhost)
  - `DB_PORT`: Database port (default: 5432)
  - `DB_USER`: Database user (default: postgres)
  - `DB_PASSWORD`: Database password (default: postgres)
  - `DB_NAME`: Database name (default: wthr)

### Development

Build the application:
```bash
make build
```

Run the application:
```bash
make run
```

Run tests:
```bash
make test
```

Run static analysis (formatting, vet, staticcheck):
```bash
make check
```

Format code:
```bash
make fmt
```

Clean build artifacts:
```bash
make clean
```

View all available commands:
```bash
make help
```

## License

See [LICENSE](LICENSE) for details.
