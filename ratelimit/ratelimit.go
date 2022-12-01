package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
)

type RateLimiter struct {
	redis       *redis.Client
	globalLimit int
	routes      map[string]int
}

func New(redisClient *redis.Client, globalRPI int, routes map[string]int) *RateLimiter {
	return &RateLimiter{
		redis:       redisClient,
		routes:      routes,
		globalLimit: globalRPI,
	}
}

// Function that checks the user's ratelimit.
// We are checking for two types of entries: the global rate limit and route rate limit
func (c *RateLimiter) CheckUserRateLimit(userId, path string) bool {
	ctx := context.TODO()

	globalKey := fmt.Sprintf("ratelimit:%v:global", userId)
	totalRequests, err := c.redis.Get(ctx, globalKey).Int()
	if err == redis.Nil {
		err = c.redis.Set(ctx, globalKey, 1, time.Second*60).Err()
		if err != nil {
			logrus.Error("redis error:", err)
			return false
		}
	}
	if totalRequests > c.globalLimit {
		return true
	}
	_, err = c.redis.Incr(ctx, globalKey).Result()
	if err != nil {
		logrus.Error("redis error:", err)
	}

	routeLimit, ok := c.routes[path]
	if !ok {
		return false
	}
	routeKey := fmt.Sprintf("ratelimit:%v:%v", userId, path)
	routeRequests, err := c.redis.Get(ctx, routeKey).Int()
	if err == redis.Nil {
		err = c.redis.Set(ctx, routeKey, 1, time.Second*60).Err()
		if err != nil {
			logrus.Error("redis error: ", err)
			return false
		}
		return false
	}
	if err != nil {
		logrus.Error("redis error: ", err)
		return false
	}
	if routeRequests > routeLimit {
		return true
	}
	_, err = c.redis.Incr(ctx, routeKey).Result()
	if err != nil {
		logrus.Error("redis error:", err)
	}
	return false
}

// TODO: implement this
func (r *RateLimiter) CheckEmailRateLimit(email string) bool {
	return true
}
