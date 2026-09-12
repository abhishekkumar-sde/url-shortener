# URL Shortener — Production-style Local Setup

A layered Go URL shortener designed to run as three separate containers:

```text
Postman
   |
   v
url-shortener service
   |
   +----> Redis
   |
   +----> DynamoDB Local
```

## Architecture

```text
Transport
   |
Endpoint
   |
Business Logic
   |
Data Layer
   +---- DynamoDB
   +---- Redis
```

### Transport layer
Owns HTTP concerns:
- routing
- JSON decoding/encoding
- HTTP status codes
- request logging
- client IP extraction

### Endpoint layer
Owns application-facing operations and cross-cutting concerns such as rate limiting.

### Business layer
Owns URL-shortening rules:
- URL validation
- short-code generation
- cache-aside flow
- repository interaction

### Data layer
Owns infrastructure:
- DynamoDB persistence
- Redis cache
- Redis rate limiting

The business layer depends on interfaces, so the storage implementation can be replaced without changing business logic.

## Run

```bash
docker compose up --build
```

Check:

```bash
curl http://localhost:8080/health
```

Create:

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.example.com/some/very/long/url"}'
```

Example:

```json
{
  "code": "1",
  "short_url": "http://localhost:8080/1"
}
```

Resolve:

```bash
curl -i http://localhost:8080/1
```

Expected:

```text
HTTP/1.1 302 Found
Location: https://www.example.com/some/very/long/url
```

## Postman

Import:

```text
postman_collection.json
```

Run `Create URL`, copy the returned code into the collection `code` variable, then run `Resolve URL`.

## Rate limits

Create:
- 100 requests/minute/IP

Resolve:
- 300 requests/minute/IP

Redis uses fixed-window counters.

## Cache

Resolve follows cache-aside:

```text
GET code
  |
Redis?
  |-- hit --> return
  |
 miss
  |
DynamoDB
  |
Redis SET
  |
return
```

## Important local behavior

DynamoDB Local is configured with `-inMemory`, so its data disappears when containers are removed/restarted.

For persistent local development, replace the DynamoDB command with a shared database path and mount a volume.

## Stop

```bash
docker compose down
```

## Rebuild

```bash
docker compose up --build
```
