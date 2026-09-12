# URL Shortener

A production-oriented URL shortener built with **Go**, **DynamoDB**, **Redis**, and Docker.

## Architecture

```text
Client
  |
  v
Go HTTP Server
  |
  +---- Rate Limiter (Redis)
  |
  +---- URL Service
          |
          +---- Redis Cache
          |
          +---- DynamoDB
```

### Components

- **Go** — HTTP API and business logic
- **DynamoDB** — persistent URL storage
- **Redis** — URL cache and distributed rate limiting
- **Docker** — local runtime environment

## API

### 1. Ping

```http
GET /ping
```

Response:

```json
{
  "message": "pong"
}
```

### 2. Create Short URL

```http
POST /api/v1/urls
Content-Type: application/json
```

Request without expiry:

```json
{
  "url": "https://www.google.com"
}
```

Request with expiry:

```json
{
  "url": "https://www.linkedin.com/in/abhishekkumar-sde/",
  "expires_in": 60
}
```

`expires_in` is the URL lifetime in **seconds**.

Examples:

- `60` = 1 minute
- `3600` = 1 hour
- `86400` = 1 day
- `0` or omitted = no URL expiry

Response:

```json
{
  "code": "1",
  "short_url": "http://localhost:10002/1",
  "expires_at": 1780000000
}
```

For a URL without expiry, `expires_at` is `0`/omitted depending on JSON serialization.

The same long URL can be shortened multiple times. Each request generates a different short code.

### 3. Resolve Short URL

```http
GET /{code}
```

Example:

```http
GET /1
```

A valid short URL responds with an HTTP **302 redirect** to the original URL.

If the code does not exist or the URL has expired:

```http
HTTP 404
```

## Resolve Flow

Redis is the fast path for redirects.

```text
GET /{code}
      |
      v
    Redis
      |
   +--+--+
   |     |
 HIT   MISS
   |     |
   v     v
check  DynamoDB
expiry    |
   |      v
   |   check expiry
   |      |
   |      v
   |   populate Redis
   |      |
   +------+
      |
      v
   Redirect
```

Redis stores both the original URL and its expiry timestamp:

```json
{
  "long_url": "https://example.com",
  "expires_at": 1780000000
}
```

This allows a cache hit to validate expiry without reading DynamoDB.

DynamoDB remains the source of truth when Redis misses.

## Expiry

Expiry is represented by `expires_at`, a Unix timestamp in seconds.

On URL creation:

```text
expires_at = current_time + expires_in
```

The expiry timestamp is:

- stored in DynamoDB
- stored in the Redis cache
- returned in the create response

Redis uses the URL's remaining lifetime as its cache TTL.

DynamoDB TTL is enabled on the `expires_at` attribute. DynamoDB TTL cleanup is asynchronous, so the application still checks `expires_at` before redirecting.

## Storage

DynamoDB table:

```text
Partition key: code (String)
TTL attribute: expires_at
```

Example item:

```json
{
  "code": "1",
  "long_url": "https://example.com",
  "created_at": "2026-09-13T00:00:00Z",
  "expires_at": 1780000000
}
```

`expires_at` is not part of the key schema; it is only the TTL attribute.

## Rate Limiting

Redis is also used for API rate limiting.

Current limits:

| Endpoint | Limit |
|---|---:|
| Create URL | 100 requests/minute/IP |
| Resolve URL | 300 requests/minute/IP |

When the limit is exceeded:

```http
HTTP 429
```

```json
{
  "error": "rate limit exceeded"
}
```

## Error Responses

### Invalid URL

```http
HTTP 400
```

```json
{
  "error": "invalid URL"
}
```

### URL Not Found / Expired

```http
HTTP 404
```

```json
{
  "error": "url not found"
}
```

### Rate Limited

```http
HTTP 429
```

```json
{
  "error": "rate limit exceeded"
}
```

## Running Locally

### Prerequisites

- Docker
- Go
- Postman (optional)

Create a Docker network:

```bash
docker network create urlshortener-network
```

### Start Redis

```bash
docker run -d   --name urlshortener-redis   --network urlshortener-network   -p 10000:6379   redis:7-alpine
```

### Start DynamoDB Local

```bash
docker run -d   --name urlshortener-dynamodb   --network urlshortener-network   -p 10001:8000   amazon/dynamodb-local
```

### Build the Server

From the repository root:

```bash
docker build -f buildscripts/build/Dockerfile -t url-shortener:latest .
```

### Run the Server

```bash
docker run -d   --name urlshortener-server   --network urlshortener-network   -p 10002:8080   -e REDIS_HOST=urlshortener-redis   -e REDIS_PORT=6379   -e DYNAMODB_HOST=urlshortener-dynamodb   -e DYNAMODB_PORT=8000   url-shortener:latest
```

The API is available at:

```text
http://localhost:10002
```

### Important Docker Networking Note

From inside the Go container, use container ports:

```text
Redis      -> urlshortener-redis:6379
DynamoDB   -> urlshortener-dynamodb:8000
Server     -> 0.0.0.0:8080
```

The host ports `10000`, `10001`, and `10002` are for access from your machine.

## Testing

Run unit tests:

```bash
go test ./...
```

Run with the race detector:

```bash
go test -race ./...
```

Build locally:

```bash
go build ./...
```

## Project Structure

```text
urlshortner/
├── encoder/
│   ├── bl/                  # Business logic
│   ├── cmd/restserver/      # Server entry point
│   ├── dl/                  # DynamoDB data layer
│   ├── endpoint/            # Endpoint/business orchestration
│   ├── inithandler/         # DynamoDB initialization and TTL
│   ├── model/               # Request/response/domain models
│   ├── svcerror/            # Service errors
│   ├── svcparam/            # Service constants/config
│   └── transport/http/      # HTTP handlers/router
├── pkg/
│   ├── ratelimiter/         # Redis-backed rate limiter
│   └── redis/               # Redis cache
├── buildscripts/
│   └── build/               # Docker build configuration
├── URL-Shortener.postman_collection.json
├── go.mod
└── README.md
```

## Design Notes

### Short Code Generation

The current implementation uses a process-local atomic counter encoded using Base62.

Base62 uses:

```text
0-9
A-Z
a-z
```

This produces compact short codes.

DynamoDB uses:

```text
attribute_not_exists(code)
```

as a conditional write, preventing an existing code from being overwritten.

For a multi-instance production deployment, a process-local counter is not globally coordinated. A production design could use random Base62 IDs, a distributed ID generator, or a Redis-backed atomic counter depending on the requirements.

### Why Redis?

Redis serves two purposes:

1. URL cache for fast redirects
2. Rate limiting

The redirect path is optimized as:

```text
Redis HIT -> no DynamoDB read
Redis MISS -> DynamoDB -> Redis
```

### Why DynamoDB?

DynamoDB provides durable storage for the URL mapping:

```text
short code -> long URL + metadata
```

It also supports conditional writes and TTL configuration.

## Postman

Import:

```text
URL-Shortener.postman_collection.json
```

The collection contains:

- Ping
- Create Short URL
- Create Short URL with Expiry
- Resolve Short URL
- Invalid URL
- Not Found

The collection uses:

```text
http://localhost:10002
```

as the local API base URL.

## License

This project is for learning and system-design/interview practice.
