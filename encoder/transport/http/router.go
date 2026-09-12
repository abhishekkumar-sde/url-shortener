package http

import (
	"net/http"

	"url-shortener/encoder/endpoint"
)

func NewRouter(endpoint *endpoint.URLEndpoint) http.Handler {
	mux := http.NewServeMux()

	h := NewHandler(endpoint)

	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/urls", h.CreateURL)
	mux.HandleFunc("/", h.ResolveURL)

	return LoggingMiddleware(mux)
}
