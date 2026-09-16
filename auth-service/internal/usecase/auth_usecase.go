package usecase

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/logger"
	redisrepo "auth-service/internal/repository/redis"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, email, password, clientIP string) error
	Login(ctx context.Context, email, password, clientIP string) (*domain.AuthTokenResponse, error)
	RefreshToken(ctx context.Context, rawRefreshToken, clientIP string) (*domain.AuthTokenResponse, error)
	Logout(ctx context.Context, rawRefreshToken, clientIP string) error
}

type authUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo redisrepo.SessionRepository
	keyManager  *utils.KeyManager
	auditLogger *logger.AuditLogger
	cfg         *config.Config
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo redisrepo.SessionRepository,
	keyManager *utils.KeyManager,
	auditLogger *logger.AuditLogger,
	cfg *config.Config,
) AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		keyManager:  keyManager,
		auditLogger: auditLogger,
		cfg:         cfg,
	}
}

func (u *authUseCase) Register(ctx context.Context, email, password, clientIP string) error {
	cleanedEmail, err := utils.ValidateAndCleanEmail(email)
	if err != nil {
		return err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     cleanedEmail,
		Password:  string(hashedPassword),
		Role:      domain.RoleUser, // Prevent privilege escalation
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return err
	}

	u.auditLogger.Log(logger.EventRegisterSuccess, cleanedEmail, clientIP, "user registered successfully")
	return nil
}

func (u *authUseCase) Login(ctx context.Context, email, password, clientIP string) (*domain.AuthTokenResponse, error) {
	cleanedEmail, err := utils.ValidateAndCleanEmail(email)
	if err != nil {
		u.auditLogger.Log(logger.EventLoginFailed, email, clientIP, "invalid email format")
		return nil, domain.ErrUnauthorized
	}

	user, err := u.userRepo.GetByEmail(ctx, cleanedEmail)
	if err != nil || !user.IsActive {
		u.auditLogger.Log(logger.EventLoginFailed, cleanedEmail, clientIP, "user not found or inactive")
		return nil, domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		u.auditLogger.Log(logger.EventLoginFailed, cleanedEmail, clientIP, "invalid credentials")
		return nil, domain.ErrUnauthorized
	}

	res, err := u.generateSession(ctx, user)
	if err != nil {
		return nil, err
	}

	u.auditLogger.Log(logger.EventLoginSuccess, cleanedEmail, clientIP, "successful authentication")
	return res, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, rawRefreshToken, clientIP string) (*domain.AuthTokenResponse, error) {
	userIDStr, err := u.sessionRepo.GetUserIDByRefreshToken(ctx, rawRefreshToken)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	_ = u.sessionRepo.RevokeRefreshToken(ctx, rawRefreshToken)

	res, err := u.generateSession(ctx, user)
	if err != nil {
		return nil, err
	}

	u.auditLogger.Log(logger.EventTokenRefreshed, user.Email, clientIP, "token refreshed successfully")
	return res, nil
}

func (u *authUseCase) Logout(ctx context.Context, rawRefreshToken, clientIP string) error {
	u.auditLogger.Log(logger.EventSessionRevoked, "anonymous", clientIP, "session revoked")
	return u.sessionRepo.RevokeRefreshToken(ctx, rawRefreshToken)
}

func (u *authUseCase) generateSession(ctx context.Context, user *domain.User) (*domain.AuthTokenResponse, error) {
	accessToken, err := u.keyManager.GenerateAccessToken(user.ID, user.Email, user.Role, u.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	if err := u.sessionRepo.StoreRefreshToken(ctx, user.ID.String(), refreshToken, u.cfg.RefreshTokenTTL); err != nil {
		return nil, err
	}

	return &domain.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.cfg.AccessTokenTTL.Seconds()),
	}, nil
}