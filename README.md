![screenshot of wthr.lol showing weather in Los Angeles, CA](https://raw.githubusercontent.com/swelljoe/wthr.lol/refs/heads/main/docs/images/LA-wthr.png)

# wthr.lol

Because I'm tired of ad-filled weather sites and apps

A simple, ad-free weather application built with Go, SQLite, HTML, Datastar, and Pico.css.

## Features

- No bullshit, just weather.
- No ads, no tracking.
- That's right, just weather.

## Tech Stack

- **Backend**: Go + SQLite
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
- `NWS_USER_AGENT`: User-Agent to use when fetching place data from government sources (e.g. `example.tld/1.0 (contact@example.tld)`)

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


## Deployment

When installing via RPM, the configuration file is located at `/etc/default/wthr`.

You **must** configure the `NWS_USER_AGENT` in this file to comply with National Weather Service API usage guidelines.

```bash
# /etc/default/wthr
NWS_USER_AGENT=my-app.com/1.0 (me@example.com)
```

## License

See [LICENSE](LICENSE) for details.
