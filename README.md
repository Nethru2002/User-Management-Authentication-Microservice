# Auth Service — Enterprise Identity & Access Management (IAM) Platform

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-blue?style=flat)](#architecture-overview)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Production Grade](https://img.shields.io/badge/Production%20Grade-100%25%20Gold%20Tier-gold)](#enterprise-security--production-standards)
[![SCIM 2.0](https://img.shields.io/badge/SCIM-2.0%20Compliant-orange)](#enterprise-scim-20-directory-synchronization)

A high-throughput, horizontally scalable, production-grade **Identity and Access Management (IAM)** platform built in **Go (Golang)** following strict **Clean Architecture**, **Domain-Driven Design (DDD)** principles, and the official **Standard Go Project Layout**.

Engineered as a commercial-grade, white-label alternative to platforms like Keycloak and Auth0, this service implements **zero-trust asymmetric cryptography (`RS256` / RSA-4096)** with **RFC 7517 JWKS discovery**, **multi-tenant isolation**, **RFC 6238 TOTP Multi-Factor Authentication (MFA)**, **federated social logins (Google & GitHub)** with anti-CSRF protections, **SCIM 2.0 directory synchronization**, **atomic Redis Lua distributed rate limiting**, **Prometheus telemetry**, and **SOC 2 / ISO 27001-ready structured security audit trails**.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Enterprise Security & Production Standards](#enterprise-security--production-standards)
- [Feature Matrix](#feature-matrix)
- [Technology Stack](#technology-stack)
- [Project Directory Structure](#project-directory-structure)
- [Prerequisites](#prerequisites)
- [Configuration & Environment Variables](#configuration--environment-variables)
- [Getting Started](#getting-started)
  - [1. Generating Persistent RSA Keys](#1-generating-persistent-rsa-keys-recommended)
  - [2. Running with Docker Compose](#2-running-with-docker-compose-full-stack)
  - [3. Running Standalone Locally](#3-running-standalone-locally)
- [Running Tests](#running-tests)
- [API Reference & Route Matrix](#api-reference--route-matrix)
- [End-to-End Walkthrough (cURL Examples)](#end-to-end-walkthrough-curl-examples)
  - [Authentication & Token Lifecycle](#1-standard-authentication--token-lifecycle)
  - [Multi-Factor Authentication (TOTP 2FA)](#2-multi-factor-authentication-totp-2fa)
  - [Email Verification & Password Reset](#3-email-verification--password-reset)
  - [Federated Social Login (OAuth2)](#4-federated-social-login-oauth2)
  - [SCIM 2.0 Identity Provider Synchronization](#5-scim-20-identity-provider-synchronization)
- [Operational & Observability Endpoints](#operational--observability-endpoints)
- [License](#license)

---

## Architecture Overview

The system strictly isolates core business rules from transport protocols and infrastructure drivers using Clean Architecture:

```text
               HTTP Clients / API Gateways / IdP Sync (Okta/Azure AD)
                                      │
                                      ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │                      Presentation & Delivery Layer                     │
 │  • Master Router & Method-Based Route Dispatcher (Go 1.22 Wildcards)   │
 │  • Handlers: Auth, User, Health, SCIM 2.0, JWKS                        │
 │  • Middlewares: Tenant Context, Metrics, Zap Logger, Distributed Rate  │
 │                 Limiter (Lua), CORS, Tracing & Request ID Propagation   │
 └───────────────────────────────────┬────────────────────────────────────┘
                                     │ (DTOs)
                                     ▼
 ┌────────────────────────────────────────────────────────────────────────┐
 │                            Use Case Layer                              │
 │  • AuthUseCase: Login, Register, Rotate, Revoke, MFA Setup & Challenge │
 │  • SCIMUseCase: RFC 7643 / RFC 7644 Provisioning & In-Memory Envelopes  │
 │  • UserUseCase: Paginated Multi-Tenant Profile & System Queries        │
 └─────────────────┬────────────────────────────────────┬─────────────────┘
                   │                                    │
                   ▼                                    ▼
 ┌───────────────────────────────────┐   ┌────────────────────────────────┐
 │           Domain Layer            │   │    Infrastructure Adapters     │
 │ • User, Tenant, Role, SCIM Models │   │ • PostgreSQL Pool (pgxpool v5) │
 │ • Repository Interfaces           │   │ • Redis Cluster Client (v9)    │
 │ • Custom Errors & Claims          │   │ • KeyManager (RSA-4096 / RS256)│
 │ • RBAC Permission Matrix          │   │ • Prometheus Collector & Zap   │
 └───────────────────────────────────┘   └────────────────────────────────┘
```

1. **Independent Core:** The domain layer has zero third-party dependencies, maintaining complete decoupling from web frameworks and databases.
2. **Defensive Ingress:** Request bodies are strictly restricted to 1 MB using `http.MaxBytesReader` to eliminate resource exhaustion and OOM vulnerabilities.
3. **Decoupled Verification:** Downstream microservices verify authenticity purely in-memory using public keys retrieved from `/.well-known/jwks.json`.

---

## Enterprise Security & Production Standards

| Pillar | Implementation | Technical Advantage |
| :--- | :--- | :--- |
| **Cryptography** | **RS256 (RSA-4096)** with Key ID (`kid`) Fingerprinting | Asymmetric keypair signing. Private key signs tokens; downstream microservices verify via the public JWKS endpoint without secret sharing. |
| **Session Control** | **15-Min Access + Redis Refresh Rotation** | Short-lived access tokens limit exposure. Refresh tokens are single-use, cryptographically random hashes with automatic rotation and instant revocation. |
| **Multi-Tenancy** | **Partitioned Data Isolation (`tenant_id`)** | Schema-level data segregation across users, roles, and SCIM resources via `X-Tenant-ID` middleware with zero-downtime default fallback. |
| **Two-Factor Auth** | **RFC 6238 TOTP Engine** | Zero-dependency standard-library implementation with clock-drift compensation ($-1, 0, +1$ intervals) and standard `otpauth://` provisioning. |
| **Anti-CSRF Protection** | **OAuth2 State Verification (RFC 6749)** | Ephemeral, cryptographically random state tokens stored in Redis with 10-minute TTLs prevent OAuth login CSRF attacks. |
| **Directory Sync** | **SCIM 2.0 (RFC 7643 / RFC 7644)** | Standardized endpoints for automated user provisioning and lifecycle synchronization from enterprise identity providers. |
| **Concurrency Safety** | **PostgreSQL Unique Constraints (`23505`)** | Prevents Time-of-Check to Time-of-Use (TOCTOU) race conditions during concurrent registrations. |
| **Traffic Shaping** | **Distributed Token-Bucket (Redis Lua)** | Atomic, proxy-aware (`X-Forwarded-For`) sliding rate limiter enforced across multi-replica deployments. |
| **Observability** | **Prometheus + Uber Zap JSON** | Four Golden Signals exposed via `/metrics` alongside structured logs tracking `request_id` and `trace_id`. |
| **Compliance** | **Immutable Security Audit Log** | Explicit logging trail (`SECURITY_AUDIT_TRAIL`) recording event type, actor email, IP address, and timestamps for SOC 2 / ISO 27001 readiness. |

---

## Feature Matrix

- [x] **Zero-Trust Asymmetric Authentication:** RS256 token signing with persistent RSA-4096 keys.
- [x] **RFC 7517 Public JWKS Discovery:** `GET /.well-known/jwks.json` endpoint for microservice ecosystems.
- [x] **Dual-Token Session Lifecycle:** 15-minute access tokens + Redis-backed rotating refresh tokens.
- [x] **Multi-Tenancy & Tenant Isolation:** Complete partitioning via `tenant_id` and `X-Tenant-ID` headers.
- [x] **Multi-Factor Authentication (MFA / 2FA):** Standard RFC 6238 TOTP setup, QR provisioning URI, and two-step verification challenges.
- [x] **Federated Social Identity (OAuth2 / OIDC):** Integrated Google and GitHub login flows with anti-CSRF state validation.
- [x] **SCIM 2.0 Identity Management:** Standard `/scim/v2/Users` and `/scim/v2/ServiceProviderConfig` endpoints for Okta and Microsoft Entra ID (Azure AD).
- [x] **Email Verification & Self-Service Password Reset:** Redis-backed one-time tokens with expiration.
- [x] **Role-Based Access Control (RBAC):** Middleware-level permission enforcement (`read:profile`, `manage:system`).
- [x] **Distributed Traffic Control:** Atomic Redis Lua token bucket rate limiter (100 req/min per IP).
- [x] **SRE Observability:** Prometheus metrics scraping (`/metrics`) and Uber Zap JSON logging.
- [x] **Kubernetes Probes:** Liveness (`/healthz`) and active dependency readiness checks (`/readyz`).
- [x] **Clean Go 1.22 Routing:** Method-based route registration and path wildcards (`{id}`).

---

## Technology Stack

- **Language:** Go 1.22+
- **Database:** PostgreSQL 16 (via `jackc/pgx/v5` connection pool)
- **Cache / Distributed State:** Redis 7 (via `redis/go-redis/v9`)
- **Asymmetric Tokens:** `golang-jwt/jwt/v5` & RSA-4096
- **Password Hashing:** `golang.org/x/crypto/bcrypt`
- **Telemetry & Metrics:** `prometheus/client_golang`
- **Structured Logging:** `go.uber.org/zap`
- **Environment Management:** `joho/godotenv`
- **Schema Migrations:** `golang-migrate/migrate/v4`
- **Unit Testing & Mocking:** `stretchr/testify`

---

## Project Directory Structure

```text
auth-service/
├── cmd/
│   ├── migrate/
│   │   └── main.go                    # Standalone database migration runner
│   └── server/
│       └── main.go                    # Application bootstrap & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go                  # Strongly typed environment configuration
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth_handler.go    # Authentication, MFA, OAuth, password reset, JWKS
│   │       │   ├── health_handler.go  # Liveness & readiness probes
│   │       │   ├── scim_handler.go    # SCIM 2.0 enterprise directory sync
│   │       │   └── user_handler.go    # Profiles and paginated administration
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go # RS256 token verification & claims injection
│   │       │   ├── cors_middleware.go # Cross-Origin Resource Sharing
│   │       │   ├── logger_middleware.go # HTTP request latency & status logging
│   │       │   ├── metrics_middleware.go # Prometheus HTTP latency tracking
│   │       │   ├── ratelimit_middleware.go # Redis Lua distributed token bucket
│   │       │   ├── rbac_middleware.go # Permission authorization barrier
│   │       │   ├── requestid_middleware.go # Correlation ID generation & propagation
│   │       │   ├── tenant_middleware.go  # Multi-tenant context extraction
│   │       │   └── tracing_middleware.go   # Distributed trace context propagation
│   │       └── router.go              # Master route tree assembly
│   ├── domain/
│   │   ├── claims.go                  # JWT custom claims definition
│   │   ├── errors.go                  # Unified domain errors
│   │   ├── pagination.go              # Pagination schemas & query models
│   │   ├── permission.go              # Permission constants & role mappings
│   │   ├── role.go                    # Role entity definition
│   │   ├── scim.go                    # SCIM 2.0 resource & message schemas
│   │   ├── tenant.go                  # Tenant entity & default tenant definitions
│   │   └── user.go                    # User model & repository contracts
│   ├── infrastructure/
│   │   ├── cache/
│   │   │   └── redis.go               # Redis client connection factory
│   │   ├── database/
│   │   │   ├── migrator.go            # Programmatic schema migration runner
│   │   │   └── postgres.go            # Tuned pgxpool connection manager
│   │   ├── logger/
│   │   │   ├── audit.go               # Security audit logging engine
│   │   │   └── zap.go                 # Uber Zap JSON logger factory
│   │   ├── metrics/
│   │   │   └── metrics.go             # Prometheus metric vector registrations
│   │   └── oauth/
│   │       └── provider.go            # Google and GitHub OAuth2 provider clients
│   ├── repository/
│   │   ├── postgres/
│   │   │   └── user_repo.go           # Multi-tenant PostgreSQL CRUD implementation
│   │   └── redis/
│   │       └── session_repo.go        # Refresh, MFA, reset, and OAuth state store
│   └── usecase/
│       ├── auth_usecase.go            # Authentication, MFA, and OAuth business logic
│       ├── auth_usecase_test.go       # Mock-based unit test suite
│       ├── scim_usecase.go            # SCIM 2.0 user lifecycle logic
│       └── user_usecase.go            # User query and pagination business logic
├── migrations/
│   ├── 000001_init_schema.down.sql    # Schema teardown script
│   ├── 000001_init_schema.up.sql      # Initial schema & indexes
│   ├── 000002_enterprise_features.down.sql # Revert multi-tenant, MFA, SCIM schema
│   └── 000002_enterprise_features.up.sql   # Multi-tenant, MFA, OAuth, SCIM schema
├── pkg/
│   └── utils/
│       ├── email.go                   # Extensible email dispatcher (Console / SMTP)
│       ├── jwt.go                     # RSA keypair manager & JWKS builder
│       ├── response.go                # Standardized JSON response envelope
│       ├── totp.go                    # RFC 6238 TOTP generator & validator
│       └── validator.go               # Email cleaning & password rule validation
├── .env                               # Local runtime configuration (git-ignored)
├── .env.example                       # Environment configuration template
├── .gitignore                         # Clean build, secret, and OS exclusions
├── Dockerfile                         # Production multi-stage minimal container
├── docker-compose.yml                 # Full stack container orchestrator
├── go.mod                             # Go module definitions
├── go.sum                             # Cryptographic dependency checksums
└── README.md                          # Project documentation
```

---

## Prerequisites

- **Go:** `1.22` or higher
- **PostgreSQL:** `15` or higher
- **Redis:** `7.x` or higher
- **Docker & Docker Compose:** *(Optional, recommended for containerized setup)*

---

## Configuration & Environment Variables

Create your local `.env` file from the provided template:

```bash
cp .env.example .env
```

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_PORT` | HTTP server listening port | `8080` |
| `DB_HOST` | PostgreSQL host address | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL user password | *Required* |
| `DB_NAME` | Database name | `auth_db` |
| `DB_SSLMODE` | PostgreSQL SSL mode (`disable`, `require`) | `disable` |
| `DB_MAX_CONNS` | Maximum database connection pool capacity | `25` |
| `DB_MIN_CONNS` | Minimum idle database connections | `5` |
| `DB_MAX_CONN_IDLE`| Maximum idle connection duration | `15m` |
| `DB_MAX_CONN_LIFE`| Maximum total connection lifetime | `1h` |
| `REDIS_HOST` | Redis host address | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis authentication password | `""` |
| `RSA_PRIVATE_KEY_PATH` | Path to persistent RSA private PEM key | `private.pem` |
| `GOOGLE_CLIENT_ID` | Google OAuth2 client ID | `""` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret | `""` |
| `GOOGLE_REDIRECT_URI` | Google OAuth2 redirect callback URL | `http://localhost:8080/api/v1/auth/oauth/callback?provider=google` |
| `GITHUB_CLIENT_ID` | GitHub OAuth2 client ID | `""` |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth2 client secret | `""` |
| `GITHUB_REDIRECT_URI` | GitHub OAuth2 redirect callback URL | `http://localhost:8080/api/v1/auth/oauth/callback?provider=github` |

---

## Getting Started

### 1. Generating Persistent RSA Keys (Recommended)

To avoid in-memory key generation drift across multiple pod replicas, generate a persistent 4096-bit RSA private key:

#### Using Go:
```powershell
Set-Content -Path genkey.go -Value @'
package main
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)
func main() {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	b := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	_ = os.WriteFile("private.pem", b, 0600)
}
'@
go run genkey.go
Remove-Item genkey.go
```

#### Using OpenSSL (Linux / macOS):
```bash
openssl genrsa -out private.pem 4096
```

Ensure `RSA_PRIVATE_KEY_PATH=private.pem` is set in your `.env`.

---

### 2. Running with Docker Compose (Full Stack)

Starts PostgreSQL (`auth-postgres`), Redis (`auth-redis`), and the compiled microservice container (`auth-service`) with automatic migrations:

```bash
docker compose up -d --build
```

Tail application logs:
```bash
docker compose logs -f iam-service
```

Shut down:
```bash
docker compose down
```

---

### 3. Running Standalone Locally

#### Step 1: Start PostgreSQL and Redis Containers
```powershell
docker run --name auth-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=nethru2002 -e POSTGRES_DB=auth_db -p 5432:5432 -d postgres:16-alpine
docker run --name auth-redis -p 6379:6379 -d redis:7-alpine
```

#### Step 2: Launch the Microservice
Because `main.go` automatically loads your `.env` via `godotenv`, simply run:

```powershell
go run cmd/server/main.go
```

---

## Running Tests

Execute the comprehensive unit test suite with coverage calculation:

```powershell
go test -v -cover ./...
```

Run with Go's race detector (requires `CGO_ENABLED=1` and a C compiler such as GCC/MinGW):
```powershell
go test -race -v -cover ./...
```

---

## API Reference & Route Matrix

### Response Envelope Format
All standard application endpoints return a uniform JSON envelope:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {},
  "error": ""
}
```

### Route Table

| Method | Endpoint | Access | Required Permission | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Public | None | Kubernetes liveness probe |
| `GET` | `/readyz` | Public | None | Kubernetes readiness probe (pings DB & Redis) |
| `GET` | `/metrics` | Public | None | Prometheus telemetry scraper |
| `GET` | `/.well-known/jwks.json` | Public | None | RFC 7517 Public Key Discovery Set |
| `POST`| `/api/v1/auth/register` | Public | None | User registration (default: `USER` role) |
| `POST`| `/api/v1/auth/login` | Public | None | Password login; returns tokens or MFA challenge |
| `POST`| `/api/v1/auth/mfa/challenge`| Public | None | Complete 2FA verification with TOTP code |
| `POST`| `/api/v1/auth/refresh` | Public | None | Refresh access token & rotate refresh token |
| `POST`| `/api/v1/auth/logout` | Public | None | Invalidate active refresh session in Redis |
| `GET` | `/api/v1/auth/verify-email`| Public | None | Confirm email verification token |
| `POST`| `/api/v1/auth/password/reset-request` | Public | None | Request password reset token via email |
| `POST`| `/api/v1/auth/password/reset-confirm` | Public | None | Set new password with one-time reset token |
| `GET` | `/api/v1/auth/oauth/url` | Public | None | Generate OAuth2 authorization URL with anti-CSRF state |
| `GET` | `/api/v1/auth/oauth/callback`| Public | None | Handle provider callback, exchange code, link account |
| `GET` | `/api/v1/users/profile` | Bearer | `read:profile` | Fetch authenticated user profile |
| `POST`| `/api/v1/users/mfa/setup`| Bearer | `read:profile` | Generate TOTP secret and `otpauth://` provisioning URI |
| `POST`| `/api/v1/users/mfa/confirm`| Bearer | `read:profile` | Validate TOTP code and activate MFA for account |
| `GET` | `/api/v1/admin/users` | Bearer | `manage:system` | Paginated user management (`ADMIN` only) |
| `GET` | `/scim/v2/ServiceProviderConfig` | Bearer | `manage:system` | SCIM 2.0 capability discovery specification |
| `GET` | `/scim/v2/Users` | Bearer | `manage:system` | List enterprise users via SCIM 2.0 |
| `POST`| `/scim/v2/Users` | Bearer | `manage:system` | Provision enterprise user via SCIM 2.0 |
| `GET` | `/scim/v2/Users/{id}` | Bearer | `manage:system` | Retrieve SCIM 2.0 user by ID |
| `PUT` | `/scim/v2/Users/{id}` | Bearer | `manage:system` | Update SCIM 2.0 user by ID |
| `DELETE`| `/scim/v2/Users/{id}`| Bearer | `manage:system` | Deprovision SCIM 2.0 user by ID |

---

## End-to-End Walkthrough (cURL Examples)

> **Note for Windows PowerShell users:** Use `curl.exe` instead of `curl` to bypass PowerShell's built-in `Invoke-WebRequest` alias.

### 1. Standard Authentication & Token Lifecycle

#### Registration
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"StrongPassword123"}'
```

#### Login
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"StrongPassword123"}'
```

**Response (`200 OK`):**
```json
{
  "success": true,
  "message": "authentication response",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjY4ZGYy...",
    "refresh_token": "f47ac10b58cc4372a5670e02b2c3d479...",
    "expires_in": 900
  }
}
```

#### Access User Profile
```powershell
curl.exe -i -X GET http://localhost:8080/api/v1/users/profile `
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### Refresh Token (Single-Use Rotation)
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/refresh `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

#### Logout (Revoke Session)
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/logout `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

---

### 2. Multi-Factor Authentication (TOTP 2FA)

#### Step 1: Enroll in MFA
Call the setup endpoint with your active access token:
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/users/mfa/setup `
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
*Returns a Base32 `secret` and an `otpauth://` URI ready for QR code generation.*

#### Step 2: Confirm and Activate MFA
Send the 6-digit code displayed in Google Authenticator:
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/users/mfa/confirm `
  -H "Authorization: Bearer <ACCESS_TOKEN>" `
  -H "Content-Type: application/json" `
  -d '{"code":"123456"}'
```

#### Step 3: Authenticate with MFA Challenge
Subsequent logins will return an MFA challenge:
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"StrongPassword123"}'
```

**Response (`200 OK`):**
```json
{
  "success": true,
  "message": "authentication response",
  "data": {
    "requires_mfa": true,
    "mfa_token": "92f7c01b2a4e4d..."
  }
}
```

Complete the challenge using the `mfa_token` and the current 6-digit TOTP code:
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/mfa/challenge `
  -H "Content-Type: application/json" `
  -d '{"mfa_token":"<MFA_TOKEN>","code":"123456"}'
```
*Returns the finalized RS256 access and refresh tokens.*

---

### 3. Email Verification & Password Reset

#### Confirm Email Verification
```powershell
curl.exe -i -X GET "http://localhost:8080/api/v1/auth/verify-email?token=<VERIFICATION_TOKEN>"
```

#### Request Password Reset
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/password/reset-request `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com"}'
```

#### Confirm Password Reset
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/password/reset-confirm `
  -H "Content-Type: application/json" `
  -d '{"token":"<RESET_TOKEN>","new_password":"NewStrongPassword456"}'
```

---

### 4. Federated Social Login (OAuth2)

#### Step 1: Generate Authorization URL (with Anti-CSRF State)
```powershell
curl.exe -i "http://localhost:8080/api/v1/auth/oauth/url?provider=google"
```
*Returns the Google OAuth URL with a secure `state` parameter cached in Redis for 10 minutes.*

#### Step 2: Handle Provider Callback
```powershell
curl.exe -i "http://localhost:8080/api/v1/auth/oauth/callback?provider=google&code=<AUTH_CODE>&state=<STATE_TOKEN>"
```
*Validates the state token against Redis, exchanges the code for user info, automatically provisions the user in the active tenant, and issues authentication tokens.*

---

### 5. SCIM 2.0 Identity Provider Synchronization

The SCIM 2.0 endpoints allow external Identity Providers (Okta, Microsoft Entra ID / Azure AD, OneLogin) to automatically provision, update, and deprovision users.

#### Query Service Provider Capabilities
```powershell
curl.exe -i http://localhost:8080/scim/v2/ServiceProviderConfig `
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

#### Provision User via SCIM 2.0
```powershell
curl.exe -i -X POST http://localhost:8080/scim/v2/Users `
  -H "Authorization: Bearer <ADMIN_TOKEN>" `
  -H "Content-Type: application/scim+json" `
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "bob@enterprise.com",
    "externalId": "okta-usr-98765",
    "active": true,
    "emails": [{"value": "bob@enterprise.com", "primary": true}]
  }'
```

#### List Users via SCIM 2.0
```powershell
curl.exe -i "http://localhost:8080/scim/v2/Users?startIndex=1&count=10" `
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

---

## Operational & Observability Endpoints

### 1. Prometheus Telemetry (`/metrics`)
Real-time metrics tracking the Four Golden Signals:
```powershell
curl.exe http://localhost:8080/metrics
```
Exposed vectors:
- `http_requests_total{method, path, status}`
- `http_request_duration_seconds{method, path}` (latency histograms)
- `http_active_requests`
- Go runtime memory allocation, GC pauses, and goroutine counts.

### 2. Public Key Discovery (`/.well-known/jwks.json`)
Public JSON Web Key Set conforming to RFC 7517:
```powershell
curl.exe http://localhost:8080/.well-known/jwks.json
```
Downstream microservices and API gateways (Kong, Traefik, AWS API Gateway) cache this endpoint to verify JWT signatures locally without querying the database.

### 3. Kubernetes Health Checks
- **Liveness Probe:** `GET /healthz` (service execution verification)
- **Readiness Probe:** `GET /readyz` (active ping against PostgreSQL and Redis pools with 2s timeout)

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).