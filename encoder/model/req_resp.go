package model

import "time"

type URL struct {
	Code      string    `dynamodbav:"code"`
	LongURL   string    `dynamodbav:"long_url"`
	CreatedAt time.Time `dynamodbav:"created_at"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
