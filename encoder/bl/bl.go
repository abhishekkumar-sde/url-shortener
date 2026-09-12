package bl

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
	"url-shortener/encoder/svcparam"
)

type URLRepository interface {
	Create(context.Context, model.URL) error
	Get(context.Context, string) (model.URL, error)
	GetByLongURL(context.Context, string) (model.URL, error)
}

type URLCache interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string) error
}

type BL struct {
	repository URLRepository
	cache      URLCache
	baseURL    string
	counter    atomic.Uint64
}

func NewEncoderBL(repository URLRepository, cache URLCache, baseURL string) *BL {
	return &BL{
		repository: repository,
		cache:      cache,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (s *BL) Create(ctx context.Context, rawURL string) (model.CreateURLResponse, error) {
	if !isValidURL(rawURL) {
		return model.CreateURLResponse{}, svcerror.ErrInvalidURL
	}

	// Check whether this URL already exists.
	existing, err := s.repository.GetByLongURL(ctx, rawURL)
	if err == nil {
		return model.CreateURLResponse{
			Code:     existing.Code,
			ShortURL: s.baseURL + "/" + existing.Code,
		}, nil
	}

	// Create a new short URL
	for i := 0; i < 10; i++ {
		code := base62(s.counter.Add(1))

		u := model.URL{
			Code:      code,
			LongURL:   rawURL,
			CreatedAt: time.Now().UTC(),
		}

		err := s.repository.Create(ctx, u)
		if errors.Is(err, svcerror.ErrConflict) {
			continue
		}
		if err != nil {
			return model.CreateURLResponse{}, err
		}

		_ = s.cache.Set(ctx, code, rawURL)

		return model.CreateURLResponse{
			Code:     code,
			ShortURL: s.baseURL + "/" + code,
		}, nil
	}

	return model.CreateURLResponse{}, fmt.Errorf("unable to generate unique code")
}

func (s *BL) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || strings.ContainsAny(code, "/?# ") {
		return "", svcerror.ErrNotFound
	}

	if longURL, err := s.cache.Get(ctx, code); err == nil {
		return longURL, nil
	}

	u, err := s.repository.Get(ctx, code)
	if err != nil {
		return "", err
	}

	_ = s.cache.Set(ctx, code, u.LongURL)
	return u.LongURL, nil
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
