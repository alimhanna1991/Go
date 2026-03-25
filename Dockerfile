# Production Dockerfile for the config-driven app in /webpage-analyzer
FROM docker.io/library/golang:1.21-bookworm AS builder

WORKDIR /src/webpage-analyzer

COPY webpage-analyzer/go.mod webpage-analyzer/go.sum ./
RUN go mod download

COPY webpage-analyzer/ ./

RUN go test ./...
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/webpage-analyzer ./main.go

FROM docker.io/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    chromium \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN ln -s /usr/bin/chromium /usr/local/bin/google-chrome

RUN groupadd --gid 1001 appgroup && \
    useradd --uid 1001 --gid appgroup --create-home --shell /usr/sbin/nologin appuser

WORKDIR /app/webpage-analyzer

COPY --from=builder /out/webpage-analyzer ./webpage-analyzer
COPY --from=builder /src/webpage-analyzer/web ./web
COPY --from=builder /src/webpage-analyzer/config ./config

RUN mkdir -p logs && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/ || exit 1

CMD ["./webpage-analyzer"]
