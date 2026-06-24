# phragmaos

A lightweight, embeddable HTTP rate limiting middleware for Go, built on the **token bucket** algorithm. It supports rate limiting by client IP address or API key out of the box, with zero external dependencies.

---

## How it works

Each unique client (identified by IP or API key) gets its own token bucket. The bucket starts full and loses one token per request. Tokens refill continuously at a fixed rate. When the bucket is empty, requests are rejected with a `429 Too Many Requests` response until enough tokens have refilled.

**Default configuration:**

| Parameter   | Value | Description                         |
|-------------|-------|-------------------------------------|
| Capacity    | 15    | Maximum tokens per bucket           |
| Refill rate | 1/s   | Tokens added per second             |
| Identifier  | IP    | How clients are distinguished       |

---

## Getting started

```bash
go get github.com/akshit-git24/phragmaos
```

### Run the server

```go
package main

import "github.com/akshit-git24/phragmaos"

func main() {
    phragmaos.Run()
}
```

---

## Configuration

All configuration is done via environment variables.

| Variable      | Default | Description                                                          |
|---------------|---------|----------------------------------------------------------------------|
| `LISTEN_PORT` | `:8080` | Address and port the server listens on (e.g. `:9090`, `0.0.0.0:8080`) |
| `EXTRACTOR`   | `IP`    | Identifier strategy: `IP` for client IP, `APIKEY` for `X-API-KEY` header |

### Example — rate limit by API key on port 9090

```bash
EXTRACTOR=APIKEY LISTEN_PORT=:9090 go run .
```

---

## API

### `GET /welcome`

A sample protected endpoint that demonstrates rate limiting in action.

**Response headers:**

| Header                  | Description                                      |
|-------------------------|--------------------------------------------------|
| `X-RateLimit-Limit`     | Maximum requests allowed per window              |
| `X-RateLimit-Remaining` | Tokens remaining in the current bucket           |
| `Retry-After`           | Seconds to wait before retrying (only on `429`)  |

**Responses:**

- `200 OK` — request allowed
- `429 Too Many Requests` — rate limit exceeded

---

## Identifier strategies

### IP (default)

Extracts the client IP from `r.RemoteAddr`. Best for public-facing APIs where you want to throttle by network address.

### API Key

Reads the `X-API-KEY` request header. Use this when clients authenticate with keys and you want per-key quotas.

To activate: set `EXTRACTOR=APIKEY` (also accepts `apikey`, `API_KEY`, `API-KEY`).

---

## Project structure

```
phragmaos/
├── main.go          # HTTP server, routing, and identifier extractors
├── token_bucket.go  # Token bucket implementation
├── store.go         # Per-client bucket store (sync.Map)
└── go.mod
```

---

## License

MIT
