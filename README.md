# Webpage Analyzer

This repository contains a Go web application for analyzing webpages and the deployment files used to run it locally or in containerized environments.

The application source lives in [`/Go/webpage-analyzer`](/Go/webpage-analyzer).

## What It Does

- Detects the HTML version
- Extracts the page title
- Counts headings by level
- Counts internal and external links
- Checks inaccessible links
- Detects login/auth flows, including JS-rendered pages via headless browser fallback
- Returns useful error messages for unreachable or non-OK pages

## Repository Layout

- [`/Go/webpage-analyzer`](/Go/webpage-analyzer): Go application source
- [`/Go/Dockerfile`](/Go/Dockerfile): production image
- [`/Go/Dockerfile.dev`](/Go/Dockerfile.dev): development image
- [`/Go/docker-compose.yml`](/Go/docker-compose.yml): local development stack
- [`/Go/docker-compose.prod.yml`](/Go/docker-compose.prod.yml): production-oriented stack with optional observability services
- [`/Go/Jenkinsfile`](/Go/Jenkinsfile): CI/CD pipeline
- [`/Go/nginx.conf`](/Go/nginx.conf): reverse proxy config

## Local Run

```bash
cd /Go/webpage-analyzer
go mod download
go run .
```

The app reads runtime settings from [`/Go/webpage-analyzer/config/app.yaml`](/Go/webpage-analyzer/config/app.yaml).

Default URL:

```text
http://localhost:8080
```

## Configuration

Main runtime configuration is in [`/Go/webpage-analyzer/config/app.yaml`](/Go/webpage-analyzer/config/app.yaml).

Supported configuration areas:

- `server.port`
- `http_client.timeout_seconds`
- `browser.enabled`
- `browser.command`
- `logging.backends`
- `logging.file`
- `logging.database`
- `logging.elasticsearch`
- `cache.enabled`
- `cache.ttl_seconds`
- `cache.redis`

## Logging

The app supports configurable error logging backends:

- File logging
- SQLite database logging
- Elasticsearch logging

Backends are selected in config with `logging.backends`.

Examples:

- `["file"]`
- `["db"]`
- `["elasticsearch"]`
- `["file", "db"]`

## Caching

The app supports Redis-backed result caching for repeated analyses.

To enable it, set:

```yaml
cache:
  enabled: true
```

and provide valid Redis connection settings in the same config file.

## Docker

Build the production image:

```bash
docker build -t webpage-analyzer .
```

Run the development stack:

```bash
docker compose up --build
```

Run the production-oriented stack:

```bash
docker compose -f docker-compose.prod.yml up --build
```

Run the production stack with Elasticsearch and Kibana:

```bash
docker compose -f docker-compose.prod.yml --profile observability up --build
```

## Testing

Run tests from the application directory:

```bash
cd /Go/webpage-analyzer
go test ./...
```

## CI/CD

The Jenkins pipeline:

- downloads Go modules
- runs linting
- runs unit tests with coverage
- builds the binary
- builds the Docker image
- runs deployment steps for staging and production

The pipeline is defined in [`/Go/Jenkinsfile`](/Go/Jenkinsfile).

## Notes

- The headless browser fallback used for login detection depends on Chromium/Chrome being available in the runtime environment.
- SQLite logging requires CGO, which is already accounted for in the Docker build.
- The production compose file can run without Elasticsearch/Kibana unless the `observability` profile is enabled.
