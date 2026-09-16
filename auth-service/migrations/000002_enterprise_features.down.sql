DROP INDEX IF EXISTS idx_users_external_id;
DROP INDEX IF EXISTS idx_users_oauth;
DROP INDEX IF EXISTS idx_users_tenant_email;

ALTER TABLE users 
    DROP COLUMN IF EXISTS external_id,
    DROP COLUMN IF EXISTS oauth_subject,
    DROP COLUMN IF EXISTS oauth_provider,
    DROP COLUMN IF EXISTS email_verified,
    DROP COLUMN IF EXISTS mfa_secret,
    DROP COLUMN IF EXISTS mfa_enabled,
    DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants;