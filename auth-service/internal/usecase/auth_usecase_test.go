package usecase

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/logger"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByOAuth(ctx context.Context, tenantID uuid.UUID, provider, subject string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, provider, subject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepository) List(ctx context.Context, tenantID uuid.UUID, pq domain.PaginationQuery) ([]*domain.User, int64, error) {
	args := m.Called(ctx, tenantID, pq)
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return m.Called(ctx, userID, token, ttl).Error(0)
}
func (m *MockSessionRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *MockSessionRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockSessionRepository) StoreMFAToken(ctx context.Context, mfaToken, userID string, ttl time.Duration) error {
	return m.Called(ctx, mfaToken, userID, ttl).Error(0)
}
func (m *MockSessionRepository) GetUserIDByMFAToken(ctx context.Context, mfaToken string) (string, error) {
	args := m.Called(ctx, mfaToken)
	return args.String(0), args.Error(1)
}
func (m *MockSessionRepository) RevokeMFAToken(ctx context.Context, mfaToken string) error {
	return m.Called(ctx, mfaToken).Error(0)
}
func (m *MockSessionRepository) StoreEmailVerifyToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return m.Called(ctx, token, userID, ttl).Error(0)
}
func (m *MockSessionRepository) GetUserIDByEmailVerifyToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *MockSessionRepository) RevokeEmailVerifyToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockSessionRepository) StorePasswordResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return m.Called(ctx, token, userID, ttl).Error(0)
}
func (m *MockSessionRepository) GetUserIDByPasswordResetToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *MockSessionRepository) RevokePasswordResetToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockSessionRepository) StoreOAuthState(ctx context.Context, state, provider string, ttl time.Duration) error {
	return m.Called(ctx, state, provider, ttl).Error(0)
}
func (m *MockSessionRepository) ValidateAndConsumeOAuthState(ctx context.Context, state, provider string) bool {
	return m.Called(ctx, state, provider).Bool(0)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendVerificationEmail(to, token string) error {
	return m.Called(to, token).Error(0)
}
func (m *MockEmailSender) SendPasswordResetEmail(to, token string) error {
	return m.Called(to, token).Error(0)
}

func setupTest(t *testing.T) (AuthUseCase, *MockUserRepository, *MockSessionRepository, *MockEmailSender) {
	mockRepo := new(MockUserRepository)
	mockSession := new(MockSessionRepository)
	mockEmail := new(MockEmailSender)

	km, err := utils.NewKeyManager("")
	assert.NoError(t, err)

	audit := logger.NewAuditLogger(zap.NewNop())
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(mockRepo, mockSession, km, mockEmail, audit, cfg)
	return uc, mockRepo, mockSession, mockEmail
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	uc, mockRepo, mockSession, mockEmail := setupTest(t)
	tenantID := domain.DefaultTenantID

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "test@example.com" && u.Role == domain.RoleUser && u.TenantID == tenantID
	})).Return(nil)
	mockSession.On("StoreEmailVerifyToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil)
	mockEmail.On("SendVerificationEmail", "test@example.com", mock.AnythingOfType("string")).Return(nil)

	err := uc.Register(context.Background(), tenantID, "test@example.com", "validPass123", "127.0.0.1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	uc, mockRepo, mockSession, _ := setupTest(t)
	tenantID := domain.DefaultTenantID

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("validPass123"), bcrypt.DefaultCost)
	uid := uuid.New()
	mockUser := &domain.User{
		ID:         uid,
		TenantID:   tenantID,
		Email:      "test@example.com",
		Password:   string(hashedPassword),
		Role:       domain.RoleUser,
		IsActive:   true,
		MFAEnabled: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockRepo.On("GetByEmail", mock.Anything, tenantID, "test@example.com").Return(mockUser, nil)
	mockSession.On("StoreRefreshToken", mock.Anything, uid.String(), mock.Anything, mock.Anything).Return(nil)

	res, err := uc.Login(context.Background(), tenantID, "test@example.com", "validPass123", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.False(t, res.RequiresMFA)
	mockRepo.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}