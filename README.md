# Webpage Analyzer

This repository contains a Go web application for analyzing webpages and the deployment files used to run it locally or in containerized environments.

The application source lives in [`/webpage-analyzer`](/webpage-analyzer).

For build, Docker usage, deployment notes, assumptions, and application improvement suggestions, see [`/DEPLOYMENT.md`](/DEPLOYMENT.md).

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
- [`/webpage-analyzer/main.go`](/webpage-analyzer/main.go): application entrypoint
- [`/webpage-analyzer/internal`](/webpage-analyzer/internal): core application packages
- [`/webpage-analyzer/web`](/webpage-analyzer/web): templates and static assets
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

Docker build, `DockerUp` / `DockerDown`, Compose usage, and deployment details are documented in [`/DEPLOYMENT.md`](/DEPLOYMENT.md).

## Testing

Run tests from the application directory:

```bash
cd /home/ali/Projects/Go/webpage-analyzer
go test ./...
```

## Architecture

- `main.go`: config loading and dependency wiring
- `internal/analyzer`: HTML, title, heading, link, and login/auth analysis
- `internal/browser`: rendered-page fallback support through headless Chrome/Chromium
- `internal/cache`: Redis-backed analysis result caching
- `internal/config`: YAML-based runtime configuration
- `internal/handlers`: HTTP handlers and template view model
- `internal/http`: outbound HTTP client abstraction and implementation
- `internal/logging`: file, SQLite, and Elasticsearch error logging backends
- `internal/services`: orchestration layer for analysis, cache, and logging

## CI/CD

The Jenkins pipeline:

- downloads Go modules
- runs linting
- runs unit tests with coverage
- builds the binary
- builds the Docker image
- runs deployment steps for staging and production

The pipeline is defined in [`/Jenkinsfile`](/Jenkinsfile).

## Git Flow

This repository should follow a simple git-flow branching model:

- `master`: production-ready branch
- `develop`: integration branch for upcoming work
- `feature/<name>`: feature branches created from `develop`
- `release/<version>`: release preparation branches created from `develop`
- `hotfix/<name>`: urgent production fixes created from `master`

Recommended flow:

1. Create feature work from `develop`
2. Merge finished features back into `develop`
3. Create a `release/*` branch when preparing a release
4. Merge releases into both `master` and `develop`
5. Create `hotfix/*` branches from `master` for urgent fixes, then merge them into both `master` and `develop`

## Notes

- The headless browser fallback used for login detection depends on Chromium/Chrome being available in the runtime environment.
- SQLite logging requires CGO, which is already accounted for in the Docker build.
- The production compose file can run without Elasticsearch/Kibana unless the `observability` profile is enabled.
