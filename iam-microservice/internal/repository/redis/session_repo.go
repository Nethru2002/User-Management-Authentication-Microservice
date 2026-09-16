package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

func (s *SessionRepository) StoreRefreshToken(ctx context.Context, userID string, tokenID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, "refresh:"+userID+":"+tokenID, "active", ttl).Err()
}

func (s *SessionRepository) ValidateRefreshToken(ctx context.Context, userID string, tokenID string) bool {
	val, err := s.rdb.Get(ctx, "refresh:"+userID+":"+tokenID).Result()
	return err == nil && val == "active"
}

func (s *SessionRepository) RevokeSession(ctx context.Context, userID string, tokenID string) error {
	return s.rdb.Del(ctx, "refresh:"+userID+":"+tokenID).Err()
}