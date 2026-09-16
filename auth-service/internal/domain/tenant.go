package domain

import (
	"time"

	"github.com/google/uuid"
)

var DefaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}