-- 1. Tenants Table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Seed Default Tenant for backwards compatibility
INSERT INTO tenants (id, name, slug, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'Default Organization', 'default', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 2. Alter Users Table to add Tenant, MFA, Social Login, and SCIM columns
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001' REFERENCES tenants(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS mfa_secret VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS oauth_subject VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255) NOT NULL DEFAULT '';

-- Drop old global unique index and create tenant-scoped unique index
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_subject);
CREATE INDEX IF NOT EXISTS idx_users_external_id ON users(tenant_id, external_id);