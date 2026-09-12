package svcerror

import "errors"

var ErrRateLimited = errors.New("rate limit exceeded")
var ErrInvalidURL = errors.New("invalid URL")
var ErrNotFound = errors.New("url not found")
var ErrConflict = errors.New("url code already exists")
