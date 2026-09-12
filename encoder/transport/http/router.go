package http

import (
	"net/http"

	"url-shortener/encoder/endpoint"
)

func NewRouter(endpoint *endpoint.URLEndpoint) http.Handler {
	mux := http.NewServeMux()

	h := NewHandler(endpoint)

	mux.HandleFunc("/ping", h.Ping)
	mux.HandleFunc("/api/v1/urls", h.CreateURL)
	mux.HandleFunc("/", h.ResolveURL)

	return LoggingMiddleware(mux)
}
