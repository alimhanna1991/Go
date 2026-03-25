# Webpage Analyzer

A production-ready web application built with Go that analyzes webpages and provides comprehensive insights about their structure, content, and accessibility.

## Features

- **HTML Version Detection**: Identifies HTML5, HTML 4.01, XHTML versions
- **Title Extraction**: Retrieves the page title
- **Heading Analysis**: Counts all heading levels (h1-h6)
- **Link Analysis**: Distinguishes between internal and external links
- **Broken Link Detection**: Checks accessibility of all links
- **Login Form Detection**: Identifies pages with login forms
- **Error Handling**: Graceful handling of unreachable URLs with HTTP status codes

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker (optional)
- Docker Compose (optional)

### Run Locally

```bash
# Clone the repository
git clone https://webpage-analyzer.git
cd webpage-analyzer

# Download dependencies
go mod download

# Run the application
go run main.go

# Open in browser
http://localhost:8080

# Build Docker image
docker build -t webpage-analyzer .

# Run container
docker run -p 8080:8080 webpage-analyzer

# Development with hot reload
docker-compose up

# Production
docker-compose -f docker-compose.prod.yml up -d

# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./test/integration/...