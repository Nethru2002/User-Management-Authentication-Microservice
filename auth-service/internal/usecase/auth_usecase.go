package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/oauth"
	redisrepo "auth-service/internal/repository/redis"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, tenantID uuid.UUID, email, password, clientIP string) error
	Login(ctx context.Context, tenantID uuid.UUID, email, password, clientIP string) (*domain.AuthTokenResponse, error)
	VerifyMFAChallenge(ctx context.Context, mfaToken, code, clientIP string) (*domain.AuthTokenResponse, error)
	SetupMFA(ctx context.Context, userID uuid.UUID) (secret, uri string, err error)
	ConfirmMFASetup(ctx context.Context, userID uuid.UUID, code string) error

	RequestEmailVerification(ctx context.Context, tenantID uuid.UUID, email string) error
	ConfirmEmailVerification(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, tenantID uuid.UUID, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error

	OAuthLoginURL(ctx context.Context, provider string) (authURL string, state string, err error)
	HandleOAuthCallback(ctx context.Context, tenantID uuid.UUID, provider, code, state, clientIP string) (*domain.AuthTokenResponse, error)

	RefreshToken(ctx context.Context, rawRefreshToken, clientIP string) (*domain.AuthTokenResponse, error)
	Logout(ctx context.Context, rawRefreshToken, clientIP string) error
}

type authUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo redisrepo.SessionRepository
	keyManager  *utils.KeyManager
	totpManager *utils.TOTPManager
	emailSender utils.EmailSender
	auditLogger *logger.AuditLogger
	googleOAuth oauth.Provider
	githubOAuth oauth.Provider
	cfg         *config.Config
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo redisrepo.SessionRepository,
	keyManager *utils.KeyManager,
	emailSender utils.EmailSender,
	auditLogger *logger.AuditLogger,
	cfg *config.Config,
) AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		keyManager:  keyManager,
		totpManager: utils.NewTOTPManager("Enterprise Auth"),
		emailSender: emailSender,
		auditLogger: auditLogger,
		googleOAuth: oauth.NewGoogleProvider(),
		githubOAuth: oauth.NewGitHubProvider(),
		cfg:         cfg,
	}
}

func (u *authUseCase) Register(ctx context.Context, tenantID uuid.UUID, email, password, clientIP string) error {
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
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         cleanedEmail,
		Password:      string(hashedPassword),
		Role:          domain.RoleUser,
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return err
	}

	vToken, _ := utils.GenerateSecureToken(32)
	_ = u.sessionRepo.StoreEmailVerifyToken(ctx, vToken, user.ID.String(), 24*time.Hour)
	_ = u.emailSender.SendVerificationEmail(user.Email, vToken)

	u.auditLogger.Log(logger.EventRegisterSuccess, cleanedEmail, clientIP, "user registered successfully")
	return nil
}

func (u *authUseCase) Login(ctx context.Context, tenantID uuid.UUID, email, password, clientIP string) (*domain.AuthTokenResponse, error) {
	cleanedEmail, err := utils.ValidateAndCleanEmail(email)
	if err != nil {
		u.auditLogger.Log(logger.EventLoginFailed, email, clientIP, "invalid email format")
		return nil, domain.ErrUnauthorized
	}

	user, err := u.userRepo.GetByEmail(ctx, tenantID, cleanedEmail)
	if err != nil || !user.IsActive {
		u.auditLogger.Log(logger.EventLoginFailed, cleanedEmail, clientIP, "user not found or inactive")
		return nil, domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		u.auditLogger.Log(logger.EventLoginFailed, cleanedEmail, clientIP, "invalid credentials")
		return nil, domain.ErrUnauthorized
	}

	if user.MFAEnabled {
		mfaToken, _ := utils.GenerateSecureToken(32)
		_ = u.sessionRepo.StoreMFAToken(ctx, mfaToken, user.ID.String(), 5*time.Minute)
		return &domain.AuthTokenResponse{
			RequiresMFA: true,
			MFAToken:    mfaToken,
		}, nil
	}

	res, err := u.generateSession(ctx, user)
	if err != nil {
		return nil, err
	}

	u.auditLogger.Log(logger.EventLoginSuccess, cleanedEmail, clientIP, "successful authentication")
	return res, nil
}

func (u *authUseCase) VerifyMFAChallenge(ctx context.Context, mfaToken, code, clientIP string) (*domain.AuthTokenResponse, error) {
	userIDStr, err := u.sessionRepo.GetUserIDByMFAToken(ctx, mfaToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	userID, _ := uuid.Parse(userIDStr)
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	if !u.totpManager.ValidateCode(user.MFASecret, code) {
		u.auditLogger.Log(logger.EventLoginFailed, user.Email, clientIP, "invalid MFA code attempt")
		return nil, domain.ErrInvalidMFACode
	}

	_ = u.sessionRepo.RevokeMFAToken(ctx, mfaToken)
	res, err := u.generateSession(ctx, user)
	if err != nil {
		return nil, err
	}

	u.auditLogger.Log(logger.EventLoginSuccess, user.Email, clientIP, "mfa challenge completed")
	return res, nil
}

func (u *authUseCase) SetupMFA(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", domain.ErrNotFound
	}

	secret, err := u.totpManager.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	user.MFASecret = secret
	user.UpdatedAt = time.Now().UTC()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return "", "", err
	}

	return secret, u.totpManager.ProvisioningURI(user.Email, secret), nil
}

func (u *authUseCase) ConfirmMFASetup(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrNotFound
	}

	if !u.totpManager.ValidateCode(user.MFASecret, code) {
		return domain.ErrInvalidMFACode
	}

	user.MFAEnabled = true
	user.UpdatedAt = time.Now().UTC()
	return u.userRepo.Update(ctx, user)
}

func (u *authUseCase) RequestEmailVerification(ctx context.Context, tenantID uuid.UUID, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil
	}
	token, _ := utils.GenerateSecureToken(32)
	_ = u.sessionRepo.StoreEmailVerifyToken(ctx, token, user.ID.String(), 24*time.Hour)
	return u.emailSender.SendVerificationEmail(user.Email, token)
}

func (u *authUseCase) ConfirmEmailVerification(ctx context.Context, token string) error {
	userIDStr, err := u.sessionRepo.GetUserIDByEmailVerifyToken(ctx, token)
	if err != nil {
		return domain.ErrInvalidToken
	}
	userID, _ := uuid.Parse(userIDStr)
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrNotFound
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now().UTC()
	_ = u.sessionRepo.RevokeEmailVerifyToken(ctx, token)
	return u.userRepo.Update(ctx, user)
}

func (u *authUseCase) RequestPasswordReset(ctx context.Context, tenantID uuid.UUID, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil
	}
	token, _ := utils.GenerateSecureToken(32)
	_ = u.sessionRepo.StorePasswordResetToken(ctx, token, user.ID.String(), 1*time.Hour)
	return u.emailSender.SendPasswordResetEmail(user.Email, token)
}

func (u *authUseCase) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}
	userIDStr, err := u.sessionRepo.GetUserIDByPasswordResetToken(ctx, token)
	if err != nil {
		return domain.ErrInvalidToken
	}
	userID, _ := uuid.Parse(userIDStr)
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrNotFound
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashed)
	user.UpdatedAt = time.Now().UTC()
	_ = u.sessionRepo.RevokePasswordResetToken(ctx, token)
	return u.userRepo.Update(ctx, user)
}

func (u *authUseCase) OAuthLoginURL(ctx context.Context, provider string) (string, string, error) {
	state, err := utils.GenerateSecureToken(24)
	if err != nil {
		return "", "", err
	}

	var authURL string
	switch provider {
	case "google":
		authURL = u.googleOAuth.GetAuthURL(state)
	case "github":
		authURL = u.githubOAuth.GetAuthURL(state)
	default:
		return "", "", errors.New("unsupported oauth provider: must be google or github")
	}

	if err := u.sessionRepo.StoreOAuthState(ctx, state, provider, 10*time.Minute); err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

func (u *authUseCase) HandleOAuthCallback(ctx context.Context, tenantID uuid.UUID, provider, code, state, clientIP string) (*domain.AuthTokenResponse, error) {
	if !u.sessionRepo.ValidateAndConsumeOAuthState(ctx, state, provider) {
		u.auditLogger.Log(logger.EventLoginFailed, "anonymous", clientIP, "oauth csrf state validation failed")
		return nil, errors.New("invalid or expired oauth state parameter")
	}

	var info *oauth.UserInfo
	var err error

	switch provider {
	case "google":
		info, err = u.googleOAuth.ExchangeCode(ctx, code)
	case "github":
		info, err = u.githubOAuth.ExchangeCode(ctx, code)
	default:
		return nil, errors.New("unsupported oauth provider")
	}

	if err != nil {
		return nil, domain.ErrOAuthExchange
	}

	user, err := u.userRepo.GetByOAuth(ctx, tenantID, provider, info.Subject)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			randomBytes := make([]byte, 16)
			_, _ = rand.Read(randomBytes)
			hashed, _ := bcrypt.GenerateFromPassword(randomBytes, bcrypt.DefaultCost)

			now := time.Now().UTC()
			user = &domain.User{
				ID:            uuid.New(),
				TenantID:      tenantID,
				Email:         info.Email,
				Password:      string(hashed),
				Role:          domain.RoleUser,
				IsActive:      true,
				EmailVerified: true,
				OAuthProvider: provider,
				OAuthSubject:  info.Subject,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := u.userRepo.Create(ctx, user); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return u.generateSession(ctx, user)
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