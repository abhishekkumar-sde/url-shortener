FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

ARG SERVICE

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /url-shortener \
    ./${SERVICE}/cmd/restserver

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

USER app

COPY --from=builder /url-shortener /url-shortener

EXPOSE 8080

ENTRYPOINT ["/url-shortener"]