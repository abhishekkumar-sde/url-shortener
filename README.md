# URL Shortener

A production-style URL shortener built with **Go**, **Redis**, **DynamoDB Local**, and **Docker**.

The project is organized into transport, endpoint, business-logic, data-access, and infrastructure layers.

## Architecture

```text
Client
  |
  v
HTTP Transport
  |
  v
Endpoint Layer
  |
  v
Business Logic
  |
  +--------------------+
  |                    |
  v                    v
DynamoDB Repository   Redis Cache
                       |
                       v
                  Redis Rate Limiter
```

### Project Structure

```text
.
├── buildscripts/
│   ├── build/
│   │   └── Dockerfile
│   ├── dynamodb/
│   │   └── run_dynamodb.sh
│   ├── redis/
│   │   └── run_redis.sh
│   ├── server/
│   │   └── run_server.sh
│   ├── default_env.sh
│   └── runall.sh
│
├── encoder/
│   ├── bl/                  # Business logic
│   ├── cmd/restserver/      # Application entry point
│   ├── dl/                  # Data layer
│   ├── endpoint/            # Endpoint/use-case layer
│   ├── inithandler/         # Dependency initialization
│   ├── model/               # Request/response/domain models
│   ├── svcerror/            # Service errors
│   ├── svcparam/            # Service constants/config keys
│   └── transport/http/      # HTTP handlers, middleware and router
│
├── pkg/
│   ├── ratelimiter/         # Redis-backed rate limiter
│   └── redis/               # Redis client and cache implementation
│
├── go.mod
├── go.sum
└── README.md
```

## Features

- Create a short URL from a long URL.
- Resolve a short code and redirect to the original URL.
- Base62 short-code generation.
- Redis caching for URL lookups.
- Redis-based request rate limiting.
- DynamoDB Local persistence.
- Automatic DynamoDB table initialization.
- HTTP request logging.
- Graceful HTTP server shutdown.
- Dockerized Go service, Redis, and DynamoDB Local.
- No Docker Compose required; the stack is started with shell scripts.

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.24 |
| HTTP | `net/http` |
| Router | `http.ServeMux` |
| Cache | Redis 7 |
| Rate limiter | Redis |
| Database | DynamoDB Local |
| AWS SDK | AWS SDK for Go v2 |
| Configuration | Viper |
| Containers | Docker |
| Short-code encoding | Base62 |

## Prerequisites

Install:

- Go 1.24+
- Docker
- curl or Postman

## Running the Project

From the project root:

```bash
./buildscripts/runall.sh
```

The script:

1. Creates the Docker network.
2. Starts Redis.
3. Starts DynamoDB Local.
4. Builds the Go Docker image.
5. Starts the URL shortener server.

If the scripts are not executable:

```bash
chmod +x buildscripts/runall.sh
chmod +x buildscripts/redis/run_redis.sh
chmod +x buildscripts/dynamodb/run_dynamodb.sh
chmod +x buildscripts/server/run_server.sh
```

### Container Ports

```text
Host                  Container

localhost:10000   ->   Redis:6379
localhost:10001   ->   DynamoDB:8000
localhost:10002   ->   Go server:8080
```

Container-to-container communication uses the Docker network and **container ports**, for example:

```text
urlshortener-redis:6379
urlshortener-dynamodb:8000
```

## API

Base URL:

```text
http://localhost:10002
```

### 1. Ping

Check that the HTTP server is running.

```http
GET /ping
```

Example:

```bash
curl http://localhost:10002/ping
```

Response:

```json
{
  "message": "pong"
}
```

Status:

```text
200 OK
```

---

### 2. Create Short URL

```http
POST /api/v1/urls
Content-Type: application/json
```

Request:

```json
{
  "url": "https://www.google.com"
}
```

Example:

```bash
curl -X POST http://localhost:10002/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.google.com"}'
```

Response:

```json
{
  "code": "1",
  "short_url": "http://localhost:10002/1"
}
```

The generated code is Base62 encoded.

Status:

```text
201 Created
```

#### Validation

Only `http` and `https` URLs are accepted.

Example invalid request:

```json
{
  "url": "not-a-url"
}
```

Response:

```json
{
  "error": "invalid URL"
}
```

Status:

```text
400 Bad Request
```

---

### 3. Resolve Short URL

```http
GET /{code}
```

Example:

```bash
curl -i http://localhost:10002/1
```

The service returns an HTTP redirect to the original URL.

```text
302 Found
Location: https://www.google.com
```

The URL is first looked up in Redis. On a cache miss, the service reads from DynamoDB and then populates Redis.

### Not Found

```text
GET /does-not-exist
```

Response:

```json
{
  "error": "short URL not found"
}
```

Status:

```text
404 Not Found
```

## Rate Limiting

The endpoint layer applies Redis-based rate limiting by client IP and operation.

| Operation | Limit | Window |
|---|---:|---|
| Create URL | 100 requests | 1 minute |
| Resolve URL | 300 requests | 1 minute |

When the limit is exceeded:

```json
{
  "error": "rate limit exceeded"
}
```

Status:

```text
429 Too Many Requests
```

## Redis Cache

Redis stores URL mappings using keys in the form:

```text
url:<code>
```

For example:

```text
url:1 -> https://www.google.com
```

The cache TTL is currently:

```text
24 hours
```

## DynamoDB

DynamoDB Local stores URL mappings in:

```text
URLMappings
```

The partition key is:

```text
code (String)
```

Stored attributes:

```text
code
long_url
created_at
```

The application automatically checks for the table during startup and creates it if it does not exist.

DynamoDB Local runs with:

```text
-sharedDb
-inMemory
```

so data is intentionally lost when the DynamoDB Local container is restarted.

## Configuration

Configuration is supplied through environment variables and read using Viper.

Current variables:

```text
HTTP_PORT
REDIS_HOST
REDIS_PORT
DYNAMODB_HOST
DYNAMODB_PORT
DYNAMODB_REGION
DYNAMODB_TABLE
BASE_URL
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
```

The Docker scripts configure these automatically.

## Useful Docker Commands

View running containers:

```bash
docker ps
```

View server logs:

```bash
docker logs -f urlshortener-server
```

View Redis logs:

```bash
docker logs -f urlshortener-redis
```

View DynamoDB logs:

```bash
docker logs -f urlshortener-dynamodb
```

Open a shell in the Alpine-based server container:

```bash
docker exec -it urlshortener-server sh
```

Stop the stack:

```bash
docker rm -f urlshortener-server urlshortener-redis urlshortener-dynamodb
```

Remove the Docker network:

```bash
docker network rm urlshortener-network
```

## Testing

Run Go tests:

```bash
go test ./...
```

Build locally:

```bash
go build ./encoder/cmd/restserver
```

## Postman

A ready-to-import Postman collection is included:

```text
postman/URL-Shortener.postman_collection.json
```

It contains:

- Ping
- Create Short URL
- Resolve Short URL
- Invalid URL example
- Not Found example

## Design Notes

### Base62

Short codes use:

```text
0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
```

This gives 62 possible characters per position.

### Cache-Aside Pattern

URL resolution follows:

```text
Request
  |
  v
Redis
  |
  +-- hit --> return URL
  |
  +-- miss
        |
        v
    DynamoDB
        |
        v
    Update Redis
        |
        v
    return URL
```

### Layered Design

The application keeps responsibilities separated:

```text
Transport
    ↓
Endpoint
    ↓
Business Logic
    ↓
Data Interfaces
    ↓
Redis / DynamoDB
```

This makes the business logic testable without requiring real Redis or DynamoDB dependencies.

## Future Improvements

Potential production improvements include:

- Distributed/atomic rate limiting with a Lua script.
- Collision-free ID generation across multiple server instances.
- Persistent DynamoDB instead of DynamoDB Local.
- Authentication/authorization for URL creation.
- Custom aliases.
- URL expiration.
- Metrics and tracing.
- Health/readiness checks.
- Horizontal scaling behind a load balancer.
- More comprehensive integration tests.
