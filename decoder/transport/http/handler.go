package http

import (
	"encoding/json"
	"errors"
	"net"
	nethttp "net/http"
	"strings"

	"url-shortener/decoder/endpoint"
	"url-shortener/decoder/model"
	"url-shortener/decoder/svcerror"
)

type Handler struct {
	endpoint *endpoint.URLEndpoint
}

func NewHandler(endpoint *endpoint.URLEndpoint) *Handler {
	return &Handler{endpoint: endpoint}
}

func (h *Handler) Ping(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]string{"message": "pong"})
}

func (h *Handler) ResolveURL(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/")
	longURL, err := h.endpoint.Resolve(r.Context(), clientIP(r), code)
	if err != nil {
		switch {
		case errors.Is(err, svcerror.ErrRateLimited):
			writeError(w, nethttp.StatusTooManyRequests, err.Error())
		case errors.Is(err, svcerror.ErrNotFound):
			writeError(w, nethttp.StatusNotFound, "short URL not found")
		default:
			writeError(w, nethttp.StatusInternalServerError, "could not resolve short URL")
		}
		return
	}

	nethttp.Redirect(w, r, longURL, nethttp.StatusFound)
}

func clientIP(r *nethttp.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func writeJSON(w nethttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w nethttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}
