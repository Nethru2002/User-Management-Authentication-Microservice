package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) *CacheRepository {
	return &CacheRepository{rdb: rdb}
}

func (c *CacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}