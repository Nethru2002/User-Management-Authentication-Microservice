package postgres

import (
	"context"
	"errors"

	"auth-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.TenantID, u.Email, u.Password, u.Role, u.IsActive,
		u.MFAEnabled, u.MFASecret, u.EmailVerified, u.OAuthProvider, u.OAuthSubject, u.ExternalID,
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at 
	          FROM users WHERE tenant_id = $1 AND email = $2`
	var u domain.User
	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Password, &u.Role, &u.IsActive,
		&u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.OAuthProvider, &u.OAuthSubject, &u.ExternalID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at 
	          FROM users WHERE id = $1`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Password, &u.Role, &u.IsActive,
		&u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.OAuthProvider, &u.OAuthSubject, &u.ExternalID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByOAuth(ctx context.Context, tenantID uuid.UUID, provider, subject string) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at 
	          FROM users WHERE tenant_id = $1 AND oauth_provider = $2 AND oauth_subject = $3`
	var u domain.User
	err := r.db.QueryRow(ctx, query, tenantID, provider, subject).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Password, &u.Role, &u.IsActive,
		&u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.OAuthProvider, &u.OAuthSubject, &u.ExternalID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at 
	          FROM users WHERE tenant_id = $1 AND external_id = $2`
	var u domain.User
	err := r.db.QueryRow(ctx, query, tenantID, externalID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Password, &u.Role, &u.IsActive,
		&u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.OAuthProvider, &u.OAuthSubject, &u.ExternalID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET email = $1, role = $2, is_active = $3, mfa_enabled = $4, mfa_secret = $5, 
	          email_verified = $6, external_id = $7, updated_at = $8, password = $9 WHERE id = $10`
	_, err := r.db.Exec(ctx, query,
		u.Email, u.Role, u.IsActive, u.MFAEnabled, u.MFASecret,
		u.EmailVerified, u.ExternalID, u.UpdatedAt, u.Password, u.ID,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, pq domain.PaginationQuery) ([]*domain.User, int64, error) {
	if pq.Page < 1 {
		pq.Page = 1
	}
	if pq.Limit < 1 {
		pq.Limit = 10
	}
	offset := (pq.Page - 1) * pq.Limit

	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, tenant_id, email, password, role, is_active, mfa_enabled, mfa_secret, email_verified, oauth_provider, oauth_subject, external_id, created_at, updated_at 
	          FROM users WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, tenantID, pq.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Password, &u.Role, &u.IsActive,
			&u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.OAuthProvider, &u.OAuthSubject, &u.ExternalID,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}

	return users, total, nil
}