# Build, Deploy, And Assumptions

## Build

```bash
cd /home/ali/Projects/Go/webpage-analyzer
go build ./...
go test ./...
```

## Run Locally

```bash
cd /home/ali/Projects/Go/webpage-analyzer
go run .
```

Default address:

```text
http://localhost:8080
```

Custom port:

```bash
PORT=9090 go run .
```

## Deployment Steps

1. Install Go 1.18 or newer.
2. Copy the repository to the target host.
3. Build the binary with `go build`.
4. Ensure outbound HTTP/HTTPS traffic is allowed from the host.
5. Run the service with `PORT` configured for the environment.
6. Place the app behind a reverse proxy such as Nginx or Caddy for TLS termination in production.

## Assumptions

- Input URLs point to publicly reachable pages.
- The target pages return parseable HTML when analysis succeeds.
- Some websites may block bots, reject `HEAD` requests, or throttle repeated requests; the application handles part of this with `GET` fallback for link checks, but some false negatives remain possible.
- Login form detection is heuristic and may not cover JavaScript-rendered authentication flows.

## Possible Improvements

- Containerize the app with a small runtime image.
- Add health checks and readiness endpoints.
- Add request timeouts, concurrency controls, and circuit breaking around outbound requests.
- Persist historical analysis results.
