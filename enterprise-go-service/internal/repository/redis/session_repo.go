package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	StoreRefreshToken(ctx context.Context, userID string, token string, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (string, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return &sessionRepository{rdb: rdb}
}

func (s *sessionRepository) StoreRefreshToken(ctx context.Context, userID string, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.rdb.Set(ctx, key, userID, ttl).Err()
}

func (s *sessionRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.rdb.Get(ctx, key).Result()
}

func (s *sessionRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.rdb.Del(ctx, key).Err()
}