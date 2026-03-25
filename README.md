# Webpage Analyzer

This repository contains a Go web application for analyzing webpages and the deployment files used to run it locally or in containerized environments.

The application source lives in [`/webpage-analyzer`](/webpage-analyzer).

For build and deployment notes, assumptions, and application improvement suggestions, see [`/DEPLOYMENT.md`](/DEPLOYMENT.md).

## What It Does

- Detects the HTML version
- Extracts the page title
- Counts headings by level
- Counts internal and external links
- Checks inaccessible links
- Detects login/auth flows, including JS-rendered pages via headless browser fallback
- Returns useful error messages for unreachable or non-OK pages

## Repository Layout

- [`/webpage-analyzer`](/webpage-analyzer): Go application source
- [`/Dockerfile`](/Dockerfile): production image
- [`/Dockerfile.dev`](/Dockerfile.dev): development image
- [`/docker-compose.yml`](/docker-compose.yml): local development stack
- [`/docker-compose.prod.yml`](/docker-compose.prod.yml): production-oriented stack with optional observability services
- [`/Jenkinsfile`](/Jenkinsfile): CI/CD pipeline
- [`/nginx.conf`](/nginx.conf): reverse proxy config

## Local Run

```bash
cd /webpage-analyzer
go mod download
go run .
```

The app reads runtime settings from [`/webpage-analyzer/config/app.yaml`](/webpage-analyzer/config/app.yaml).

Default URL:

```text
http://localhost:8080
```

## Configuration

Main runtime configuration is in [`/webpage-analyzer/config/app.yaml`](/webpage-analyzer/config/app.yaml).

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

Start the built container with the helper script:

```bash
cd /Go
./DockerUp
```

Stop and remove it with:

```bash
cd /Go
./DockerDown
```

Default helper-script values:

- Container name: `webpage-analyzer-app`
- Image: `localhost/webpage-analyzer:latest`
- Host port: `8080`

You can override them when starting the container:

```bash
PORT=9090 IMAGE=localhost/webpage-analyzer:latest ./DockerUp
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
cd /webpage-analyzer
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

The pipeline is defined in [`/Jenkinsfile`](/Jenkinsfile).

## Notes

- The headless browser fallback used for login detection depends on Chromium/Chrome being available in the runtime environment.
- SQLite logging requires CGO, which is already accounted for in the Docker build.
- The production compose file can run without Elasticsearch/Kibana unless the `observability` profile is enabled.
