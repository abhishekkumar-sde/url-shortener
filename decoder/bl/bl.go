package bl

import (
	"context"
	"strings"
	"time"

	"url-shortener/decoder/model"
	"url-shortener/decoder/svcerror"
)

type URLRepository interface {
	Get(context.Context, string) (model.URL, error)
}

type URLCache interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string, time.Duration) error
}

type BL struct {
	repository URLRepository
	cache      URLCache
	baseURL    string
}

func NewDecoderBL(repository URLRepository, cache URLCache, baseURL string) *BL {
	return &BL{
		repository: repository,
		cache:      cache,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (s *BL) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)

	if code == "" || strings.ContainsAny(code, "/?# ") {
		return "", svcerror.ErrNotFound
	}

	// Redis is the fast path.
	if longURL, err := s.cache.Get(ctx, code); err == nil {
		return longURL, nil
	}

	// Cache miss -> DynamoDB.
	u, err := s.repository.Get(ctx, code)
	if err != nil {
		return "", err
	}

	// URL has expired.
	if u.ExpiresAt > 0 && time.Now().Unix() >= u.ExpiresAt {
		return "", svcerror.ErrNotFound
	}

	// Populate Redis with remaining lifetime.
	if u.ExpiresAt > 0 {
		ttl := time.Until(time.Unix(u.ExpiresAt, 0))

		if ttl > 0 {
			_ = s.cache.Set(ctx, code, u.LongURL, ttl)
		}
	} else {
		_ = s.cache.Set(ctx, code, u.LongURL, 0)
	}

	return u.LongURL, nil
}
