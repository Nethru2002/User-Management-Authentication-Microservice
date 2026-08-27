package usecase

import (
	"context"
	"math"

	"enterprise-go-service/internal/domain"

	"github.com/google/uuid"
)

type UserUseCase interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ListUsers(ctx context.Context, pq domain.PaginationQuery) (*domain.PaginatedResult, error)
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

func (u *userUseCase) ListUsers(ctx context.Context, pq domain.PaginationQuery) (*domain.PaginatedResult, error) {
	users, total, err := u.userRepo.List(ctx, pq)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pq.Limit)))

	return &domain.PaginatedResult{
		Data:       users,
		Page:       pq.Page,
		Limit:      pq.Limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}, nil
}