# Webpage Analyzer

This repository contains a Go application for analyzing webpages, split into a small web frontend service and a separate analysis API service.

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
cd /home/ali/Projects/Go/webpage-analyzer
go mod download
```

The default web service reads runtime settings from [`/webpage-analyzer/config/app.yaml`](/webpage-analyzer/config/app.yaml).
The analysis service sample config is [`/webpage-analyzer/config/app.analysis.yaml`](/webpage-analyzer/config/app.analysis.yaml).

Run the web service:

```bash
cd /home/ali/Projects/Go/webpage-analyzer
go run .
```

Run the analysis service:

```bash
cd /home/ali/Projects/Go/webpage-analyzer
APP_CONFIG_PATH=config/app.analysis.yaml go run .
```

## Configuration

Main runtime configuration is in [`/webpage-analyzer/config/app.yaml`](/webpage-analyzer/config/app.yaml).

Supported configuration areas:

- `service.role`
- `server.port`
- `analysis_api.base_url`
- `analysis_api.timeout_seconds`
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

- `main.go`: config loading and service startup
- `config/app.yaml`: web frontend configuration
- `config/app.analysis.yaml`: analysis API configuration
- `internal/app`: service-specific bootstrap and runtime assembly
- `internal/api`: HTTP contract between the web service and the analysis service
- `internal/analyzer`: webpage analysis logic and URL policy
- `internal/browser`: rendered-page fallback through headless Chrome/Chromium
- `internal/cache`: Redis-backed analysis result caching
- `internal/config`: YAML-based runtime configuration
- `internal/handlers`: web frontend handlers and template view models
- `internal/http`: outbound HTTP client abstraction and implementation
- `internal/logging`: file, SQLite, and Elasticsearch logging backends
- `internal/services`: analysis use-case orchestration, cache, and logging

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
