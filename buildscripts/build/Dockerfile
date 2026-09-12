FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener ./encoder/cmd/restserver

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

USER app

COPY --from=builder /url-shortener /url-shortener

EXPOSE 8080

ENTRYPOINT ["/url-shortener"]