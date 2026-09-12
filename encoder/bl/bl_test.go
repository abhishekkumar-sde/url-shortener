package bl

import (
	"context"
	"testing"
	"time"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
)

type fakeRepo struct {
	urls map[string]model.URL
}

func (f *fakeRepo) Create(_ context.Context, u model.URL) error {
	if _, exists := f.urls[u.Code]; exists {
		return svcerror.ErrConflict
	}

	f.urls[u.Code] = u
	return nil
}

func (f *fakeRepo) Get(_ context.Context, code string) (model.URL, error) {
	u, ok := f.urls[code]
	if !ok {
		return model.URL{}, svcerror.ErrNotFound
	}

	return u, nil
}

type fakeCache struct {
	values map[string]string
}

func (f *fakeCache) Get(_ context.Context, code string) (string, error) {
	value, ok := f.values[code]
	if !ok {
		return "", svcerror.ErrNotFound
	}

	return value, nil
}

func (f *fakeCache) Set(_ context.Context, code string, longURL string, _ time.Duration) error {
	f.values[code] = longURL
	return nil
}

func TestCreateAndResolve(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	created, err := service.Create(context.Background(), "https://example.com", 0)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Code == "" {
		t.Fatal("expected generated code")
	}

	if created.ShortURL != "http://localhost:8080/"+created.Code {
		t.Fatalf("ShortURL = %q, want valid short URL", created.ShortURL)
	}

	got, err := service.Resolve(context.Background(), created.Code)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("Resolve() = %q, want %q", got, "https://example.com")
	}
}

func TestRejectInvalidURL(t *testing.T) {
	service := NewEncoderBL(
		&fakeRepo{urls: map[string]model.URL{}},
		&fakeCache{values: map[string]string{}},
		"http://localhost:8080",
	)

	_, err := service.Create(context.Background(), "not-a-url", 0)
	if err != svcerror.ErrInvalidURL {
		t.Fatalf("Create() error = %v, want ErrInvalidURL", err)
	}
}

func TestCreateSameURLGeneratesDifferentCodes(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	first, err := service.Create(context.Background(), "https://example.com", 0)
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second, err := service.Create(context.Background(), "https://example.com", 0)
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}

	if first.Code == second.Code {
		t.Fatalf("expected different codes, got %q for both", first.Code)
	}

	firstURL, err := service.Resolve(context.Background(), first.Code)
	if err != nil {
		t.Fatalf("Resolve(first) error = %v", err)
	}

	secondURL, err := service.Resolve(context.Background(), second.Code)
	if err != nil {
		t.Fatalf("Resolve(second) error = %v", err)
	}

	if firstURL != "https://example.com" {
		t.Fatalf("first Resolve() = %q, want %q", firstURL, "https://example.com")
	}

	if secondURL != "https://example.com" {
		t.Fatalf("second Resolve() = %q, want %q", secondURL, "https://example.com")
	}
}

func TestCreateWithExpiry(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	before := time.Now().Unix()

	created, err := service.Create(context.Background(), "https://example.com", 3600)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	after := time.Now().Unix()

	if created.ExpiresAt == 0 {
		t.Fatal("expected ExpiresAt to be set")
	}

	expectedMin := before + 3600
	expectedMax := after + 3600

	if created.ExpiresAt < expectedMin || created.ExpiresAt > expectedMax {
		t.Fatalf("ExpiresAt = %d, want between %d and %d", created.ExpiresAt, expectedMin, expectedMax)
	}

	got, err := service.Resolve(context.Background(), created.Code)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("Resolve() = %q, want %q", got, "https://example.com")
	}
}

func TestCreateWithoutExpiry(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	created, err := service.Create(context.Background(), "https://example.com", 0)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ExpiresAt != 0 {
		t.Fatalf("ExpiresAt = %d, want 0", created.ExpiresAt)
	}
}

func TestResolveNotFound(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	_, err := service.Resolve(context.Background(), "doesnotexist")
	if err != svcerror.ErrNotFound {
		t.Fatalf("Resolve() error = %v, want ErrNotFound", err)
	}
}
