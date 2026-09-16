package usecase

import (
	"context"
	"testing"
	"time"

	"enterprise-go-service/internal/config"
	"enterprise-go-service/internal/domain"
	"enterprise-go-service/internal/infrastructure/logger"
	"enterprise-go-service/pkg/utils"

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
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
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
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) List(ctx context.Context, pq domain.PaginationQuery) ([]*domain.User, int64, error) {
	args := m.Called(ctx, pq)
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

func setupTest(t *testing.T) (AuthUseCase, *MockUserRepository, *MockSessionRepository) {
	mockRepo := new(MockUserRepository)
	mockSession := new(MockSessionRepository)
	km, err := utils.NewKeyManager("")
	assert.NoError(t, err)

	audit := logger.NewAuditLogger(zap.NewNop())
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(mockRepo, mockSession, km, audit, cfg)
	return uc, mockRepo, mockSession
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	uc, mockRepo, _ := setupTest(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "test@example.com" && u.Role == domain.RoleUser
	})).Return(nil)

	err := uc.Register(context.Background(), "test@example.com", "validPass123", "127.0.0.1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	uc, mockRepo, mockSession := setupTest(t)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("validPass123"), bcrypt.DefaultCost)
	uid := uuid.New()
	mockUser := &domain.User{
		ID:        uid,
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      domain.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)
	mockSession.On("StoreRefreshToken", mock.Anything, uid.String(), mock.Anything, mock.Anything).Return(nil)

	res, err := uc.Login(context.Background(), "test@example.com", "validPass123", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	mockRepo.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}