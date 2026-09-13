package bl

import (
	"context"
	"testing"
	"time"

	"url-shortener/decoder/model"
	"url-shortener/decoder/svcerror"
)

type fakeRepo struct {
	urls map[string]model.URL
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

func newTestService() *BL {
	return NewDecoderBL(
		&fakeRepo{urls: map[string]model.URL{}},
		&fakeCache{values: map[string]string{}},
		"http://localhost:8080",
	)
}

func TestResolveNotFound(t *testing.T) {
	service := newTestService()

	_, err := service.Resolve(context.Background(), "doesnotexist")
	if err != svcerror.ErrNotFound {
		t.Fatalf("Resolve() error = %v, want ErrNotFound", err)
	}
}
