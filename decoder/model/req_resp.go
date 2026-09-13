package model

import "time"

type URL struct {
	Code      string    `dynamodbav:"code"`
	LongURL   string    `dynamodbav:"long_url"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	ExpiresAt int64     `dynamodbav:"expires_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
