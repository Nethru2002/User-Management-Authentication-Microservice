package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	Role          Role      `json:"role"`
	IsActive      bool      `json:"is_active"`
	MFAEnabled    bool      `json:"mfa_enabled"`
	MFASecret     string    `json:"-"`
	EmailVerified bool      `json:"email_verified"`
	OAuthProvider string    `json:"oauth_provider,omitempty"`
	OAuthSubject  string    `json:"oauth_subject,omitempty"`
	ExternalID    string    `json:"external_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	MFAToken     string `json:"mfa_token,omitempty"`
	RequiresMFA  bool   `json:"requires_mfa,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByOAuth(ctx context.Context, tenantID uuid.UUID, provider, subject string) (*User, error)
	GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, pq PaginationQuery) ([]*User, int64, error)
}