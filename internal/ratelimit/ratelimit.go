package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
)

type RateLimiter struct {
	redis *redis.Client
}

func New(redisClient *redis.Client, globalRPI int, routes map[string]int) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
	}
}

func (r *RateLimiter) CheckRateLimit(ctx context.Context, identifier, path string, limit int, cooldown time.Duration) bool {

	key := fmt.Sprintf("ratelimit:%v:%v", path, identifier)
	totalRequests, err := r.redis.Get(ctx, key).Int()

	if err == redis.Nil {
		err = r.redis.Set(ctx, key, 1, cooldown).Err()
		if err != nil {
			logrus.Error("redis error:", err)
			return false
		}
	}
	if totalRequests > limit {
		return true
	}
	_, err = r.redis.Incr(ctx, key).Result()
	if err != nil {
		logrus.Error("redis error:", err)
	}
	return false
}
