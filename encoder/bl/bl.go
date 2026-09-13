package bl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
	"url-shortener/encoder/svcparam"
)

type URLRepository interface {
	Create(context.Context, model.URL) error
}

type URLCache interface {
	Set(context.Context, string, string, time.Duration) error
}

type IDGenerator interface {
	NextID(context.Context) (uint64, error)
}

type BL struct {
	repository  URLRepository
	cache       URLCache
	idGenerator IDGenerator
	baseURL     string
}

func NewEncoderBL(repository URLRepository, cache URLCache, idGenerator IDGenerator, baseURL string) *BL {
	return &BL{
		repository:  repository,
		cache:       cache,
		idGenerator: idGenerator,
		baseURL:     strings.TrimRight(baseURL, "/"),
	}
}

func (s *BL) Create(ctx context.Context, rawURL string, expiresIn int64) (model.CreateURLResponse, error) {
	if !isValidURL(rawURL) {
		return model.CreateURLResponse{}, svcerror.ErrInvalidURL
	}

	var expiresAt int64

	if expiresIn > 0 {
		expiresAt = time.Now().Add(
			time.Duration(expiresIn) * time.Second,
		).Unix()
	}

	for i := 0; i < 10; i++ {
		id, err := s.idGenerator.NextID(ctx)
		if err != nil {
			return model.CreateURLResponse{}, fmt.Errorf("generate ID: %w", err)
		}

		code := base62(id)

		u := model.URL{
			Code:      code,
			LongURL:   rawURL,
			CreatedAt: time.Now().UTC(),
			ExpiresAt: expiresAt,
		}

		err = s.repository.Create(ctx, u)

		if errors.Is(err, svcerror.ErrConflict) {
			continue
		}

		if err != nil {
			return model.CreateURLResponse{}, err
		}

		if expiresAt > 0 {
			ttl := time.Until(time.Unix(expiresAt, 0))

			if ttl > 0 {
				_ = s.cache.Set(ctx, code, rawURL, ttl)
			}
		} else {
			_ = s.cache.Set(ctx, code, rawURL, 0)
		}

		return model.CreateURLResponse{
			Code:      code,
			ShortURL:  s.baseURL + "/" + code,
			ExpiresAt: expiresAt,
		}, nil
	}

	return model.CreateURLResponse{}, fmt.Errorf("unable to generate unique code")
}

func isValidURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil &&
		parsed.Host != "" &&
		(parsed.Scheme == "http" || parsed.Scheme == "https")
}

func base62(n uint64) string {
	alphabet := svcparam.Alphabet

	if n == 0 {
		return "0"
	}

	var buf [11]byte
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}

	return string(buf[i:])
}
