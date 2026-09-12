package bl

import (
	"context"
	"testing"

	"url-shortener/encoder/model"
	"url-shortener/encoder/svcerror"
)

type fakeRepo struct {
	urls map[string]model.URL
}

func (f *fakeRepo) Create(_ context.Context, u model.URL) error {
	f.urls[u.Code] = u
	return nil
}

func (f *fakeRepo) Get(_ context.Context, code string) (model.URL, error) {
	return f.urls[code], nil
}

type fakeCache struct {
	values map[string]string
}

func (f *fakeCache) Get(_ context.Context, code string) (string, error) {
	value, ok := f.values[code]
	if !ok {
		return "", errCacheMiss{}
	}
	return value, nil
}

func (f *fakeCache) Set(_ context.Context, code, value string) error {
	f.values[code] = value
	return nil
}

type errCacheMiss struct{}

func (errCacheMiss) Error() string { return "cache miss" }

func TestCreateAndResolve(t *testing.T) {
	repo := &fakeRepo{urls: map[string]model.URL{}}
	cache := &fakeCache{values: map[string]string{}}

	service := NewEncoderBL(repo, cache, "http://localhost:8080")

	created, err := service.Create(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Code == "" {
		t.Fatal("expected generated code")
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

	_, err := service.Create(context.Background(), "not-a-url")
	if err != svcerror.ErrInvalidURL {
		t.Fatalf("Create() error = %v, want ErrInvalidURL", err)
	}
}
