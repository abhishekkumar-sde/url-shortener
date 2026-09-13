package http

import (
	"net/http"

	"url-shortener/decoder/endpoint"
)

func NewRouter(endpoint *endpoint.URLEndpoint) http.Handler {
	mux := http.NewServeMux()

	h := NewHandler(endpoint)

	mux.HandleFunc("/ping", h.Ping)
	mux.HandleFunc("/", h.ResolveURL)

	return LoggingMiddleware(mux)
}
