package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, identifier, operation string, limit int64, window time.Duration) (bool, error) {
	bucket := time.Now().UTC().Unix() / int64(window.Seconds())
	key := fmt.Sprintf("ratelimit:%s:%s:%d", operation, identifier, bucket)

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := r.client.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}
