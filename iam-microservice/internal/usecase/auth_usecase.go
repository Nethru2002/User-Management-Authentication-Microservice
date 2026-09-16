package usecase

import (
	"context"
	"time"

	"enterprise-go-service/internal/domain"
	"enterprise-go-service/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, email, password string, role domain.Role) error
	Login(ctx context.Context, email, password string) (string, error)
}

type authUseCase struct {
	userRepo domain.UserRepository
}

func NewAuthUseCase(userRepo domain.UserRepository) AuthUseCase {
	return &authUseCase{userRepo: userRepo}
}

func (u *authUseCase) Register(ctx context.Context, email, password string, role domain.Role) error {
	existing, _ := u.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return domain.ErrConflict
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return u.userRepo.Create(ctx, user)
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}