# Build, Deployment, Assumptions, And Improvements

This document is the short handoff note bundled with the repository.

## Main Build Steps

```bash
cd /Go/webpage-analyzer
go mod download
go test ./...
go build ./...
```

## Main Run Steps

```bash
cd /Go/webpage-analyzer
go run .
```

The application reads runtime configuration from:

[`config/app.yaml`](/Go/webpage-analyzer/config/app.yaml)

The listening port is controlled by:

- `server.port`

## Docker / Deployment Steps

1. Build the image from the repository root with `docker build -t webpage-analyzer .`.
2. Run the development stack with `docker compose up --build`.
3. Run the production-oriented stack with `docker compose -f docker-compose.prod.yml up --build`.
4. If Redis caching is required, enable `cache.enabled: true` in [`config/app.yaml`](/Go/webpage-analyzer/config/app.yaml) and ensure Redis is reachable.
5. If Elasticsearch logging is required, enable the backend in config and start the observability profile from Compose.
6. Place the service behind Nginx or another reverse proxy for external access.

## Decisions And Assumptions

- The application is implemented as a modular monolith, not microservices, because the scope does not justify distributed complexity.
- The app accepts public HTTP/HTTPS URLs and normalizes missing schemes when the input still looks like a host.
- HTML analysis uses raw HTTP responses first, then headless browser rendering when login/auth detection likely needs client-side rendering.
- Login detection is heuristic-based because websites implement authentication flows differently.
- Link accessibility uses `HEAD` first and falls back to `GET` when required.
- Error logging is pluggable and can write to file, SQLite, or Elasticsearch depending on configuration.
- Analysis results can be cached in Redis to reduce repeat network calls for the same URL.
- SQLite logging was chosen as the lightweight database-backed option for local or single-node deployments.
- Chromium/Chrome must be available in environments where rendered-page detection is needed.

## Suggestions For Improvement

- Add background workers for link accessibility checks to reduce response time on large pages.
- Add request-scoped structured logs and correlation IDs.
- Add an admin page or metrics endpoint for cache hit rate, analysis timing, and error rates.
- Add richer validation and SSRF protection around outbound URL fetching.
- Add persistence for historical analysis results and re-analysis comparisons.
- Add stronger integration tests for Docker, Compose, Redis, and Elasticsearch-backed deployments.
- Add configuration profiles for local, staging, and production instead of a single shared YAML file.
- Add URL allow/deny rules, private-network blocking, and stricter outbound request validation to harden SSRF protection.
- Run link accessibility checks concurrently with bounded worker pools to reduce latency on pages with many links.
- Improve login detection for multi-step authentication flows, SSO redirects, iframes, and other client-rendered auth patterns.
- Add retry policies and circuit breaking for outbound HTTP, Redis, and Elasticsearch operations.
- Expose a JSON API alongside the server-rendered HTML interface.
- Add cache invalidation controls and freshness metadata so users can understand when cached results were produced.
- Can add Resolvers ( GraphQl or REST API ) for Mobile app .
