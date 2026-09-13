package endpoint

import (
	"context"
	"time"

	"url-shortener/encoder/bl"
	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
)

type RateLimiter interface {
	Allow(context.Context, string, string, int64, time.Duration) (bool, error)
}

type URLEndpoint struct {
	encoderBL *bl.BL
	limiter   RateLimiter
}

func NewURLEndpoint(encoderBL *bl.BL, limiter RateLimiter) *URLEndpoint {
	return &URLEndpoint{encoderBL: encoderBL, limiter: limiter}
}

func (e *URLEndpoint) Create(ctx context.Context, clientIP string, req model.CreateURLRequest) (model.CreateURLResponse, error) {
	allowed, err := e.limiter.Allow(ctx, clientIP, "create", 100, time.Minute)

	if err != nil {
		return model.CreateURLResponse{}, err
	}

	if !allowed {
		return model.CreateURLResponse{}, svcerror.ErrRateLimited
	}

	return e.encoderBL.Create(ctx, req.URL, req.ExpiresIn)
}
