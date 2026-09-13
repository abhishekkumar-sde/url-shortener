package bl

import (
	"context"
	"errors"
	"testing"
	"time"

	"url-shortener/decoder/model"
	"url-shortener/decoder/svcerror"
)

type fakeRepo struct {
	urls  map[string]model.URL
	calls int
}

func (f *fakeRepo) Get(
	_ context.Context,
	code string,
) (model.URL, error) {
	f.calls++

	u, ok := f.urls[code]
	if !ok {
		return model.URL{}, svcerror.ErrNotFound
	}

	return u, nil
}

type fakeCache struct {
	values   map[string]string
	getCalls int
	setCalls int
	lastTTL  time.Duration
}

func (f *fakeCache) Get(
	_ context.Context,
	code string,
) (string, error) {
	f.getCalls++

	value, ok := f.values[code]
	if !ok {
		return "", errors.New("cache miss")
	}

	return value, nil
}

func (f *fakeCache) Set(
	_ context.Context,
	code string,
	longURL string,
	ttl time.Duration,
) error {
	f.setCalls++
	f.lastTTL = ttl

	f.values[code] = longURL
	return nil
}

func newTestBL() (*BL, *fakeRepo, *fakeCache) {
	repo := &fakeRepo{
		urls: make(map[string]model.URL),
	}

	cache := &fakeCache{
		values: make(map[string]string),
	}

	service := NewDecoderBL(
		repo,
		cache,
		"http://localhost:10002",
	)

	return service, repo, cache
}

func TestResolveCacheHit(t *testing.T) {
	service, repo, cache := newTestBL()

	cache.values["abc"] = "https://example.com"

	got, err := service.Resolve(
		context.Background(),
		"abc",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("unexpected URL: %s", got)
	}

	if repo.calls != 0 {
		t.Fatalf("repository should not be called on cache hit")
	}
}

func TestResolveCacheMiss(t *testing.T) {
	service, repo, cache := newTestBL()

	repo.urls["abc"] = model.URL{
		Code:      "abc",
		LongURL:   "https://example.com",
		CreatedAt: time.Now().UTC(),
	}

	got, err := service.Resolve(
		context.Background(),
		"abc",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("unexpected URL: %s", got)
	}

	if repo.calls != 1 {
		t.Fatalf("expected repository to be called once")
	}

	if cache.setCalls != 1 {
		t.Fatalf("expected URL to be cached")
	}
}

func TestResolveExpiredURL(t *testing.T) {
	service, repo, _ := newTestBL()

	repo.urls["expired"] = model.URL{
		Code:      "expired",
		LongURL:   "https://example.com",
		CreatedAt: time.Now().Add(-10 * time.Minute).UTC(),
		ExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
	}

	_, err := service.Resolve(
		context.Background(),
		"expired",
	)

	if !errors.Is(err, svcerror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveNotFound(t *testing.T) {
	service, repo, cache := newTestBL()

	_, err := service.Resolve(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, svcerror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if repo.calls != 1 {
		t.Fatalf("expected repository to be called once")
	}

	if cache.getCalls != 1 {
		t.Fatalf("expected cache to be checked once")
	}
}

func TestResolveInvalidCode(t *testing.T) {
	service, repo, cache := newTestBL()

	_, err := service.Resolve(
		context.Background(),
		"abc/def",
	)

	if !errors.Is(err, svcerror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if repo.calls != 0 {
		t.Fatalf("repository should not be called")
	}

	if cache.getCalls != 0 {
		t.Fatalf("cache should not be called")
	}
}
