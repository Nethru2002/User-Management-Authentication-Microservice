package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (string, error)
	RevokeRefreshToken(ctx context.Context, token string) error

	StoreMFAToken(ctx context.Context, mfaToken, userID string, ttl time.Duration) error
	GetUserIDByMFAToken(ctx context.Context, mfaToken string) (string, error)
	RevokeMFAToken(ctx context.Context, mfaToken string) error

	StoreEmailVerifyToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetUserIDByEmailVerifyToken(ctx context.Context, token string) (string, error)
	RevokeEmailVerifyToken(ctx context.Context, token string) error

	StorePasswordResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetUserIDByPasswordResetToken(ctx context.Context, token string) (string, error)
	RevokePasswordResetToken(ctx context.Context, token string) error

	// Anti-CSRF OAuth State Validation
	StoreOAuthState(ctx context.Context, state, provider string, ttl time.Duration) error
	ValidateAndConsumeOAuthState(ctx context.Context, state, provider string) bool
}

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return &sessionRepository{rdb: rdb}
}

func (s *sessionRepository) StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("refresh_token:%s", token), userID, ttl).Err()
}

func (s *sessionRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("refresh_token:%s", token)).Result()
}

func (s *sessionRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("refresh_token:%s", token)).Err()
}

func (s *sessionRepository) StoreMFAToken(ctx context.Context, mfaToken, userID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("mfa_challenge:%s", mfaToken), userID, ttl).Err()
}

func (s *sessionRepository) GetUserIDByMFAToken(ctx context.Context, mfaToken string) (string, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("mfa_challenge:%s", mfaToken)).Result()
}

func (s *sessionRepository) RevokeMFAToken(ctx context.Context, mfaToken string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("mfa_challenge:%s", mfaToken)).Err()
}

func (s *sessionRepository) StoreEmailVerifyToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("email_verify:%s", token), userID, ttl).Err()
}

func (s *sessionRepository) GetUserIDByEmailVerifyToken(ctx context.Context, token string) (string, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("email_verify:%s", token)).Result()
}

func (s *sessionRepository) RevokeEmailVerifyToken(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("email_verify:%s", token)).Err()
}

func (s *sessionRepository) StorePasswordResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("pwd_reset:%s", token), userID, ttl).Err()
}

func (s *sessionRepository) GetUserIDByPasswordResetToken(ctx context.Context, token string) (string, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("pwd_reset:%s", token)).Result()
}

func (s *sessionRepository) RevokePasswordResetToken(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("pwd_reset:%s", token)).Err()
}

func (s *sessionRepository) StoreOAuthState(ctx context.Context, state, provider string, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("oauth_state:%s", state), provider, ttl).Err()
}

func (s *sessionRepository) ValidateAndConsumeOAuthState(ctx context.Context, state, provider string) bool {
	key := fmt.Sprintf("oauth_state:%s", state)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil || val != provider {
		return false
	}
	_ = s.rdb.Del(ctx, key)
	return true
}