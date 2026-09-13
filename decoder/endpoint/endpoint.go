package endpoint

import (
	"context"
	"time"

	"url-shortener/decoder/bl"
	"url-shortener/decoder/svcerror"
)

type RateLimiter interface {
	Allow(context.Context, string, string, int64, time.Duration) (bool, error)
}

type URLEndpoint struct {
	decoderBL *bl.BL
	limiter   RateLimiter
}

func NewURLEndpoint(decoderBL *bl.BL, limiter RateLimiter) *URLEndpoint {
	return &URLEndpoint{decoderBL: decoderBL, limiter: limiter}
}

func (e *URLEndpoint) Resolve(ctx context.Context, clientIP, code string) (string, error) {
	allowed, err := e.limiter.Allow(ctx, clientIP, "resolve", 300, time.Minute)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", svcerror.ErrRateLimited
	}

	return e.decoderBL.Resolve(ctx, code)
}
