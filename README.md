# Auth Service — Production-Grade IAM Microservice

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-blue?style=flat)](#architecture-overview)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Production Grade](https://img.shields.io/badge/Production%20Grade-Gold%20Tier-gold)](#enterprise-security--production-standards)

A high-throughput, horizontally scalable, production-grade **Identity & Access Management (IAM)** microservice built in **Go (Golang)** following strict **Clean Architecture**, **Domain-Driven Design (DDD)** principles, and the official **Standard Go Project Layout**.

This service implements zero-trust asymmetric cryptography (`RS256` / RSA-4096), dynamic RFC 7517 public JWKS discovery, atomic distributed token-bucket rate limiting via Redis Lua, Prometheus observability, and SOC 2 / ISO 27001-ready structured security audit logging.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Enterprise Security & Production Standards](#enterprise-security--production-standards)
- [Technology Stack](#technology-stack)
- [Project Directory Structure](#project-directory-structure)
- [Prerequisites](#prerequisites)
- [Configuration & Environment Variables](#configuration--environment-variables)
- [Getting Started](#getting-started)
  - [1. Running with Docker Compose (Recommended)](#1-running-with-docker-compose-recommended)
  - [2. Running Standalone Locally](#2-running-standalone-locally)
- [Running Tests](#running-tests)
- [API Reference & Route Matrix](#api-reference--route-matrix)
- [End-to-End Walkthrough (cURL Examples)](#end-to-end-walkthrough-curl-examples)
- [Operational & Observability Endpoints](#operational--observability-endpoints)
- [License](#license)

---

## Architecture Overview

The system strictly isolates domain logic from infrastructure adapters using a multi-tiered Clean Architecture approach:

```text
               HTTP Requests (Clients / API Gateways)
                               │
                               ▼
 ┌───────────────────────────────────────────────────────────┐
 │               Presentation / Delivery Layer               │
 │  • Handlers: Auth, User, Health, JWKS                     │
 │  • Middlewares: RS256 Auth, RBAC, Rate-Limit, Metrics     │
 └─────────────────────────────┬─────────────────────────────┘
                               │ (DTOs)
                               ▼
 ┌───────────────────────────────────────────────────────────┐
 │                       Use Case Layer                      │
 │  • Core Application Business Logic                        │
 │  • Session Management & Token Rotation                    │
 └──────────────┬─────────────────────────────┬──────────────┘
                │                             │
                ▼                             ▼
 ┌───────────────────────────┐   ┌───────────────────────────┐
 │       Domain Layer        │   │  Infrastructure Adapters  │
 │ • Entities (User, Role)   │   │ • PostgreSQL (pgxpool)    │
 │ • Repository Interfaces   │   │ • Redis (Sessions & Rate) │
 │ • Error Definitions       │   │ • KeyManager (RSA-4096)   │
 │ • Permission Matrix       │   │ • Metrics & Zap Logger    │
 └───────────────────────────┘   └───────────────────────────┘
```

1. **Independent Domain:** Domain logic has no dependencies on databases, HTTP libraries, or third-party frameworks.
2. **Defensive Boundaries:** Request bodies are strictly restricted to 1 MB using `http.MaxBytesReader` to eliminate resource exhaustion and OOM vulnerabilities.
3. **Decoupled Key Management:** Downstream microservices verify authenticity purely in-memory using public keys retrieved from `/.well-known/jwks.json`.

---

## Enterprise Security & Production Standards

| Pillar | Implementation | Technical Advantage |
| :--- | :--- | :--- |
| **Cryptography** | **RS256 (RSA-4096)** Asymmetric Signing | Eliminates shared symmetric secrets. The IAM service signs tokens using a private key; all downstream services verify via the public JWKS endpoint. |
| **Session Control** | **15-Min Access + Redis Refresh Rotation** | Short-lived access tokens limit exposure. Refresh tokens are single-use, cryptographically random hashes with automatic rotation and instant revocation. |
| **Authorization** | **Role-Based Access Control (RBAC)** | Granular permission-to-role mappings (`read:profile`, `manage:system`) validated by middleware. |
| **Privilege Safety** | **Zero Public Role Escalation** | Public registration defaults strictly to `USER`. Admin accounts cannot be provisioned or escalated through public API endpoints. |
| **Concurrency** | **PostgreSQL Unique Constraints (`23505`)** | Prevents Time-of-Check to Time-of-Use (TOCTOU) race conditions during concurrent registrations. |
| **Traffic Shaping** | **Distributed Token-Bucket (Redis Lua)** | Atomic, proxy-aware (`X-Forwarded-For`) sliding rate limiter enforced across multi-replica deployments. |
| **Observability** | **Prometheus + Uber Zap JSON** | Four Golden Signals exposed via `/metrics` alongside structured logs tracking `request_id` and `trace_id`. |
| **Compliance** | **Immutable Security Audit Log** | Explicit logging trail (`SECURITY_AUDIT_TRAIL`) recording event type, actor email, IP address, and timestamps. |

---

## Technology Stack

- **Language:** Go 1.22+
- **Database:** PostgreSQL 16 (via `jackc/pgx/v5` connection pool)
- **Cache / Distributed State:** Redis 7 (via `redis/go-redis/v9`)
- **Asymmetric Tokens:** `golang-jwt/jwt/v5` & RSA-4096
- **Password Hashing:** `golang.org/x/crypto/bcrypt`
- **Telemetry & Metrics:** `prometheus/client_golang`
- **Structured Logging:** `go.uber.org/zap`
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
│   │       │   ├── auth_handler.go    # Authentication, refresh, logout, JWKS
│   │       │   ├── health_handler.go  # Liveness & readiness probes
│   │       │   └── user_handler.go    # Profiles and paginated administration
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go # RS256 token verification & claims injection
│   │       │   ├── cors_middleware.go # Cross-Origin Resource Sharing
│   │       │   ├── logger_middleware.go # HTTP request latency & status logging
│   │       │   ├── metrics_middleware.go # Prometheus HTTP latency tracking
│   │       │   ├── ratelimit_middleware.go # Redis Lua distributed token bucket
│   │       │   ├── rbac_middleware.go # Permission authorization barrier
│   │       │   ├── requestid_middleware.go # Correlation ID generation & propagation
│   │       │   └── tracing_middleware.go   # Distributed trace context propagation
│   │       └── router.go              # Master route tree assembly
│   ├── domain/
│   │   ├── claims.go                  # JWT custom claims definition
│   │   ├── errors.go                  # Unified domain errors
│   │   ├── pagination.go              # Pagination schemas & query models
│   │   ├── permission.go              # Permission constants & role mappings
│   │   ├── role.go                    # Role entity definition
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
│   │   └── metrics/
│   │       └── metrics.go             # Prometheus metric vector registrations
│   ├── repository/
│   │   ├── postgres/
│   │   │   └── user_repo.go           # PostgreSQL CRUD implementation
│   │   └── redis/
│   │       └── session_repo.go        # Refresh session store & token revocation
│   └── usecase/
│       ├── auth_usecase.go            # Authentication & session business logic
│       ├── auth_usecase_test.go       # Mock-based unit test suite
│       └── user_usecase.go            # User query and pagination business logic
├── migrations/
│   ├── 000001_init_schema.down.sql    # Schema teardown script
│   └── 000001_init_schema.up.sql      # Schema initialization & indexes
├── pkg/
│   └── utils/
│       ├── jwt.go                     # RSA keypair manager & JWKS builder
│       ├── response.go                # Standardized JSON response envelope
│       └── validator.go               # Email cleaning & password rule validation
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
- **Docker & Docker Compose:** *(Optional, recommended for immediate containerized setup)*

---

## Configuration & Environment Variables

Copy the template configuration file:

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
| `RSA_PRIVATE_KEY_PATH` | Path to RSA private PEM key *(auto-generates if empty)* | `""` |

---

## Getting Started

### 1. Running with Docker Compose (Recommended)

Starts PostgreSQL (`auth-postgres`), Redis (`auth-redis`), and the compiled application (`auth-service`) with automatic health checks:

```bash
docker compose up -d --build
```

To tail application logs:
```bash
docker compose logs -f iam-service
```

To shut down:
```bash
docker compose down
```

---

### 2. Running Standalone Locally

#### Step 1: Start PostgreSQL and Redis Containers
```powershell
docker run --name auth-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=nethru2002 -e POSTGRES_DB=auth_db -p 5432:5432 -d postgres:16-alpine
docker run --name auth-redis -p 6379:6379 -d redis:7-alpine
```

#### Step 2: Set Environment Variables & Start the Service

##### On Windows (PowerShell):
```powershell
$env:APP_PORT="8080"
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="nethru2002"
$env:DB_NAME="auth_db"
$env:DB_SSLMODE="disable"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="6379"
$env:REDIS_PASSWORD=""

go run cmd/server/main.go
```

##### On Linux / macOS / Git Bash:
```bash
export APP_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=nethru2002
export DB_NAME=auth_db
export DB_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=""

go run cmd/server/main.go
```

---

## Running Tests

Execute the unit test suite with coverage calculation:

```bash
go test -v -cover ./...
```

To run with Go's race detector (requires `CGO_ENABLED=1` and a C compiler such as GCC/MinGW):
```bash
go test -race -v -cover ./...
```

---

## API Reference & Route Matrix

### Response Envelope Format
All API endpoints return a standardized JSON structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {},
  "error": ""
}
```

### Endpoints

| Method | Endpoint | Access | Required Permission | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Public | None | Kubernetes liveness probe |
| `GET` | `/readyz` | Public | None | Kubernetes readiness probe (pings DB & Redis) |
| `GET` | `/metrics` | Public | None | Prometheus telemetry scraper |
| `GET` | `/.well-known/jwks.json` | Public | None | RFC 7517 Public Key Discovery Set |
| `POST`| `/api/v1/auth/register` | Public | None | User registration (default: `USER` role) |
| `POST`| `/api/v1/auth/login` | Public | None | Authenticate and issue RS256 token pair |
| `POST`| `/api/v1/auth/refresh` | Public | None | Refresh access token & rotate refresh token |
| `POST`| `/api/v1/auth/logout` | Public | None | Invalidate active refresh session |
| `GET` | `/api/v1/users/profile`| Bearer | `read:profile` | Fetch active user profile |
| `GET` | `/api/v1/admin/users` | Bearer | `manage:system` | Paginated list of users (`ADMIN` only) |

---

## End-to-End Walkthrough (cURL Examples)

> **Note for Windows PowerShell users:** Use `curl.exe` instead of `curl` to avoid PowerShell's built-in `Invoke-WebRequest` alias.

### 1. Register a Standard User
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"StrongPassword123"}'
```

### 2. Authenticate & Obtain Tokens
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"StrongPassword123"}'
```

**Response (`200 OK`):**
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6...",
    "refresh_token": "f47ac10b58cc4372a5670e02b2c3d479...",
    "expires_in": 900
  }
}
```

### 3. Access Protected Profile
```powershell
curl.exe -i -X GET http://localhost:8080/api/v1/users/profile `
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 4. Refresh Access Token (Token Rotation)
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/refresh `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

### 5. Logout & Terminate Session
```powershell
curl.exe -i -X POST http://localhost:8080/api/v1/auth/logout `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

### 6. Admin RBAC Verification
1. Register an account intended for administration:
   ```powershell
   curl.exe -i -X POST http://localhost:8080/api/v1/auth/register `
     -H "Content-Type: application/json" `
     -d '{"email":"admin@enterprise.com","password":"AdminSecurePassword123"}'
   ```
2. Promote the user directly in PostgreSQL (preventing public role escalation):
   ```powershell
   docker exec -it auth-postgres psql -U postgres -d auth_db -c "UPDATE users SET role = 'ADMIN' WHERE email = 'admin@enterprise.com';"
   ```
3. Log in as `admin@enterprise.com` to receive an `ADMIN`-scoped token.
4. Access the paginated administration endpoint:
   ```powershell
   curl.exe -i -X GET "http://localhost:8080/api/v1/admin/users?page=1&limit=10" `
     -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
   ```

---

## Operational & Observability Endpoints

### 1. Prometheus Telemetry (`/metrics`)
Exposes live runtime statistics and request latency histograms:
```powershell
curl.exe http://localhost:8080/metrics
```
Monitored metrics:
- `http_requests_total{method, path, status}`
- `http_request_duration_seconds{method, path}`
- `http_active_requests`
- Go runtime memory, GC pauses, and goroutine allocations.

### 2. Downstream Key Discovery (`/.well-known/jwks.json`)
Exposes public keys conforming to RFC 7517 for verification by downstream services and API Gateways:
```powershell
curl.exe http://localhost:8080/.well-known/jwks.json
```

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).