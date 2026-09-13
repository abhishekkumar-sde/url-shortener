package bl

import (
	"context"
	"errors"
	"testing"
	"time"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
)

type fakeRepo struct {
	urls  map[string]model.URL
	calls int
	err   error
}

func (f *fakeRepo) Create(_ context.Context, u model.URL) error {
	f.calls++

	if f.err != nil {
		return f.err
	}

	if _, exists := f.urls[u.Code]; exists {
		return svcerror.ErrConflict
	}

	f.urls[u.Code] = u
	return nil
}

type fakeCache struct {
	setCalls int
	lastTTL  time.Duration
}

func (f *fakeCache) Set(
	_ context.Context,
	_ string,
	_ string,
	ttl time.Duration,
) error {
	f.setCalls++
	f.lastTTL = ttl
	return nil
}

type fakeIDGenerator struct {
	ids   []uint64
	index int
}

func (f *fakeIDGenerator) NextID(_ context.Context) (uint64, error) {
	if f.index >= len(f.ids) {
		return 0, errors.New("no more IDs")
	}

	id := f.ids[f.index]
	f.index++
	return id, nil
}

func newTestBL(
	repo URLRepository,
	cache URLCache,
	idGenerator IDGenerator,
) *BL {
	return NewEncoderBL(
		repo,
		cache,
		idGenerator,
		"http://localhost:10002",
	)
}

func TestCreateInvalidURL(t *testing.T) {
	repo := &fakeRepo{urls: make(map[string]model.URL)}
	cache := &fakeCache{}
	ids := &fakeIDGenerator{ids: []uint64{1}}

	service := newTestBL(repo, cache, ids)

	_, err := service.Create(
		context.Background(),
		"invalid-url",
		0,
	)

	if !errors.Is(err, svcerror.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}

	if repo.calls != 0 {
		t.Fatalf("repository should not be called")
	}
}

func TestCreateWithoutExpiry(t *testing.T) {
	repo := &fakeRepo{urls: make(map[string]model.URL)}
	cache := &fakeCache{}
	ids := &fakeIDGenerator{ids: []uint64{1}}

	service := newTestBL(repo, cache, ids)

	resp, err := service.Create(
		context.Background(),
		"https://example.com",
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Code == "" {
		t.Fatal("expected code")
	}

	if resp.ShortURL != "http://localhost:10002/"+resp.Code {
		t.Fatalf("unexpected short URL: %s", resp.ShortURL)
	}

	if resp.ExpiresAt != 0 {
		t.Fatalf("expected no expiry, got %d", resp.ExpiresAt)
	}

	if cache.setCalls != 1 {
		t.Fatalf("expected cache Set to be called once")
	}
}

func TestCreateWithExpiry(t *testing.T) {
	repo := &fakeRepo{urls: make(map[string]model.URL)}
	cache := &fakeCache{}
	ids := &fakeIDGenerator{ids: []uint64{1}}

	service := newTestBL(repo, cache, ids)

	before := time.Now().Unix()

	resp, err := service.Create(
		context.Background(),
		"https://example.com",
		60,
	)

	after := time.Now().Unix()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ExpiresAt < before+60 || resp.ExpiresAt > after+60 {
		t.Fatalf("unexpected ExpiresAt: %d", resp.ExpiresAt)
	}

	if cache.lastTTL <= 0 {
		t.Fatalf("expected positive cache TTL")
	}
}

func TestCreateNegativeExpiry(t *testing.T) {
	repo := &fakeRepo{urls: make(map[string]model.URL)}
	cache := &fakeCache{}
	ids := &fakeIDGenerator{ids: []uint64{1}}

	service := newTestBL(repo, cache, ids)

	_, err := service.Create(
		context.Background(),
		"https://example.com",
		-1,
	)

	if !errors.Is(err, svcerror.ErrInvalidExpiry) {
		t.Fatalf("expected ErrInvalidExpiry, got %v", err)
	}

	if repo.calls != 0 {
		t.Fatalf("repository should not be called")
	}
}

type conflictRepo struct {
	calls int
}

func (f *conflictRepo) Create(
	_ context.Context,
	u model.URL,
) error {
	f.calls++

	if f.calls == 1 {
		return svcerror.ErrConflict
	}

	return nil
}

func TestCreateRetriesOnConflict(t *testing.T) {
	repo := &conflictRepo{}
	cache := &fakeCache{}
	ids := &fakeIDGenerator{
		ids: []uint64{1, 2},
	}

	service := newTestBL(repo, cache, ids)

	resp, err := service.Create(
		context.Background(),
		"https://example.com",
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Code == "" {
		t.Fatal("expected code")
	}

	if repo.calls != 2 {
		t.Fatalf(
			"expected 2 repository calls, got %d",
			repo.calls,
		)
	}
}
