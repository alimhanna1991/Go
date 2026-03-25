# Webpage Analyzer

`webpage-analyzer` is a Go web application that analyzes a user-provided webpage URL and renders the results in a server-side HTML interface.

The application reports:

- HTML version
- Page title
- Heading counts by level
- Internal and external link counts
- Inaccessible link count
- Login form detection
- Reachability errors with HTTP status codes when available

## Run

```bash
cd /home/ali/Projects/Go/webpage-analyzer
go run .
```

Open `http://localhost:8080`.

You can override the port with `PORT`, for example:

```bash
PORT=9090 go run .
```

## Architecture

- `main.go`: application bootstrap and HTTP server wiring
- `internal/handlers`: HTTP handlers and template view model
- `internal/services`: orchestration layer for the analyzer use case
- `internal/analyzer`: webpage analysis logic split by responsibility
- `internal/http`: outbound HTTP client abstraction and implementation
- `internal/models`: shared result and request models
- `web/templates`: server-rendered HTML templates
- `web/static`: static assets

## Main Decisions And Assumptions

- The app targets publicly reachable URLs over HTTP or HTTPS.
- URLs without a scheme are normalized to `http://...` when they still look like valid hosts.
- Non-`200 OK` responses from the target page are shown to the user as reachability errors with status code and description.
- Link accessibility is checked with `HEAD` first and falls back to `GET` when a server does not support `HEAD`.
- Login form detection is heuristic-based and focuses on actual `<form>` elements containing password fields plus identity or sign-in hints.

## Quality

- Dependency boundaries are interface-based to keep analyzer and handler code testable.
- The handler uses an explicit page view model instead of passing raw domain models into the template directly.
- Tests cover analyzer behavior, link analysis, login form detection, HTTP client behavior, and handler responses.

## Suggested Improvements

- Add asynchronous link checking with worker pools for faster analysis on pages with many links.
- Add structured logging and request tracing.
- Add rate limiting and outbound allow-list controls for safer production deployment.
- Add caching for repeated analyses of the same URL.
- Extend login detection heuristics and expose richer link diagnostics in the UI.
