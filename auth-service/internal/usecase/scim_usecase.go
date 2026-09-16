package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"auth-service/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SCIMUseCase interface {
	CreateUser(ctx context.Context, tenantID uuid.UUID, req *domain.SCIMUser) (*domain.SCIMUser, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.SCIMUser, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *domain.SCIMUser) (*domain.SCIMUser, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, tenantID uuid.UUID, pq domain.PaginationQuery) (*domain.SCIMListResponse, error)
}

type scimUseCase struct {
	userRepo domain.UserRepository
}

func NewSCIMUseCase(userRepo domain.UserRepository) SCIMUseCase {
	return &scimUseCase{userRepo: userRepo}
}

func (s *scimUseCase) CreateUser(ctx context.Context, tenantID uuid.UUID, req *domain.SCIMUser) (*domain.SCIMUser, error) {
	email := req.UserName
	if len(req.Emails) > 0 && req.Emails[0].Value != "" {
		email = req.Emails[0].Value
	}

	randomPass := make([]byte, 16)
	_, _ = rand.Read(randomPass)
	hashed, _ := bcrypt.GenerateFromPassword(randomPass, bcrypt.DefaultCost)

	now := time.Now().UTC()
	user := &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         email,
		Password:      string(hashed),
		Role:          domain.RoleUser,
		IsActive:      req.Active,
		ExternalID:    req.ExternalID,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.toSCIM(user), nil
}

func (s *scimUseCase) GetUser(ctx context.Context, id uuid.UUID) (*domain.SCIMUser, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toSCIM(u), nil
}

func (s *scimUseCase) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.SCIMUser) (*domain.SCIMUser, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(req.Emails) > 0 && req.Emails[0].Value != "" {
		u.Email = req.Emails[0].Value
	}
	u.IsActive = req.Active
	u.ExternalID = req.ExternalID
	u.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return s.toSCIM(u), nil
}

func (s *scimUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *scimUseCase) ListUsers(ctx context.Context, tenantID uuid.UUID, pq domain.PaginationQuery) (*domain.SCIMListResponse, error) {
	users, total, err := s.userRepo.List(ctx, tenantID, pq)
	if err != nil {
		return nil, err
	}

	scimResources := make([]*domain.SCIMUser, 0, len(users))
	for _, u := range users {
		scimResources = append(scimResources, s.toSCIM(u))
	}

	return &domain.SCIMListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: total,
		ItemsPerPage: pq.Limit,
		StartIndex:   (pq.Page-1)*pq.Limit + 1,
		Resources:    scimResources,
	}, nil
}

func (s *scimUseCase) toSCIM(u *domain.User) *domain.SCIMUser {
	return &domain.SCIMUser{
		Schemas:    []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:         u.ID.String(),
		ExternalID: u.ExternalID,
		UserName:   u.Email,
		Active:     u.IsActive,
		Emails: []domain.SCIMEmail{
			{Value: u.Email, Primary: true, Type: "work"},
		},
		Meta: domain.SCIMMeta{
			ResourceType: "User",
			Created:      u.CreatedAt.Format(time.RFC3339),
			LastModified: u.UpdatedAt.Format(time.RFC3339),
			Location:     fmt.Sprintf("/scim/v2/Users/%s", u.ID),
		},
	}
}