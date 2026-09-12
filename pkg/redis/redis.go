package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(client *redis.Client, ttl time.Duration) *Cache {
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, code string) (string, error) {
	return c.client.Get(ctx, key(code)).Result()
}

func (c *Cache) Set(ctx context.Context, code, longURL string) error {
	return c.client.Set(ctx, key(code), longURL, c.ttl).Err()
}

func key(code string) string {
	return "url:" + code
}
