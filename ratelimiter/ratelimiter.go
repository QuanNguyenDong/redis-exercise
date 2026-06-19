package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
    // Allow returns true if the request is within the rate limit.
    Allow(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error)

    // Remaining returns how many requests the client can still make in the current window.
    Remaining(ctx context.Context, clientID string, limit int) (int, error)
}

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter() *RateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &RateLimiter{client: client}
}

func (limiter *RateLimiter) Allow(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error) {
	key := "ratelimit:" + clientID
	count, err := limiter.client.Incr(ctx, key).Result()

	if count == 1 {
		limiter.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), err
}

func (limiter *RateLimiter) Remaining(ctx context.Context, clientID string, limit int) (int, error) {
	key := "ratelimit:" + clientID
	count, err := limiter.client.Get(ctx, key).Int()

	if err == redis.Nil {
		return limit, nil
	}
	if err != nil {
		return 0, err
	}

	return max(limit - count, 0), nil
}
