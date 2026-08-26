package usecase

import (
	"context"

	"enterprise-go-service/internal/domain"

	"github.com/google/uuid"
)

type UserUseCase interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type userUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(userRepo domain.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (u *userUseCase) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}