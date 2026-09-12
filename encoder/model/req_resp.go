package model

import "time"

type URL struct {
	Code      string    `dynamodbav:"code"`
	LongURL   string    `dynamodbav:"long_url"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	ExpiresAt int64     `dynamodbav:"expires_at"`
}

type CreateURLRequest struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in,omitempty"`
}

type CreateURLResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"short_url"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
