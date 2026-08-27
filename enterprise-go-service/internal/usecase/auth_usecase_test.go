package usecase

import (
	"context"
	"testing"
	"time"

	"enterprise-go-service/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
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
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, pq domain.PaginationQuery) ([]*domain.User, int64, error) {
	args := m.Called(ctx, pq)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, domain.ErrNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := NewAuthUseCase(mockRepo)
	err := uc.Register(context.Background(), "test@example.com", "password123", domain.RoleUser)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mockUser := &domain.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      domain.RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

	uc := NewAuthUseCase(mockRepo)
	token, err := uc.Login(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}