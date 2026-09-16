# Enterprise User Management & IAM Microservice

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-blue?style=flat)](#architecture-overview)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Security Standard](https://img.shields.io/badge/Security-Gold%20Tier%20%7C%20OWASP%20Hardened-brightgreen)](#security--hardening-matrix)

A production-grade, highly scalable **Identity & Access Management (IAM)** microservice built in **Go (Golang)** following strict **Clean Architecture** principles and the official **Standard Go Project Layout**. 

Engineered for zero-trust microservice ecosystems with asymmetric cryptographic token signing (RS256/RSA-4096), RFC 7517 public JWKS key discovery, Redis-backed sliding-window token-bucket rate limiting, Prometheus telemetry, SOC 2/ISO 27001-ready structured audit trails, and automated database schema migrations.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Project Directory Structure](#project-directory-structure)
- [Enterprise Production Highlights](#enterprise-production-highlights)
- [Security & Hardening Matrix](#security--hardening-matrix)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
- [Configuration Reference](#configuration-reference)
- [Installation & Quickstart](#installation--quickstart)
- [Running Automated Migrations](#running-automated-migrations)
- [Testing & Quality Assurance](#testing--quality-assurance)
- [API Endpoints Reference](#api-endpoints-reference)
- [API Walkthrough & Testing Guide](#api-walkthrough--testing-guide)
- [Observability & Monitoring](#observability--monitoring)
- [Docker & Containerized Deployment](#docker--containerized-deployment)
- [Production Best Practices & Operations](#production-best-practices--operations)
- [License](#license)

---

## Architecture Overview

The microservice follows a decoupled, multi-tier **Clean Architecture** (Ports and Adapters / Hexagonal) model to guarantee that enterprise domain entities remain entirely isolated from database drivers, HTTP transports, and third-party frameworks:

```text
 ┌─────────────────────────────────────────────────────────────┐
 │                HTTP Delivery / Transport Layer              │
 │  (Handlers, Middlewares: RateLimit, Auth, RBAC, Metrics)    │
 └──────────────────────────────┬──────────────────────────────┘
                                │
 ┌──────────────────────────────▼──────────────────────────────┐
 │                      Use Case Layer                         │
 │        (Business Logic, Auth Flow, Session Rules)           │
 └──────────────────────────────┬──────────────────────────────┘
                                │
 ┌──────────────────────────────▼──────────────────────────────┐
 │                       Domain Layer                          │
 │      (Pure Entities, Errors, Claims, Role Definitions)      │
 └──────────────────────────────▲──────────────────────────────┘
                                │
 ┌──────────────────────────────┴──────────────────────────────┐
 │            Infrastructure & Repository Data Layer           │
 │(PostgreSQL pgxpool, Redis Store, Prometheus, Zap, Migrator) │
 └─────────────────────────────────────────────────────────────┘
```

---

## Project Directory Structure

```text
enterprise-go-service/
├── cmd/
│   ├── migrate/
│   │   └── main.go                 # Standalone migration CLI runner
│   └── server/
│       └── main.go                 # Application bootstrap entrypoint & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Strongly typed, fail-fast environment config
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth_handler.go # Registration, login, token refresh, logout, JWKS
│   │       │   ├── health_handler.go# Kubernetes livez & readyz dependency probes
│   │       │   └── user_handler.go # User profile retrieval & paginated admin queries
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go    # RS256 Bearer token validation
│   │       │   ├── cors_middleware.go    # Strict CORS controls & pre-flight checks
│   │       │   ├── logger_middleware.go  # Structured Zap request/latency logging
│   │       │   ├── metrics_middleware.go # Prometheus HTTP latency & active gauge
│   │       │   ├── ratelimit_middleware.go# Atomic Redis Lua sliding-window limiter
│   │       │   ├── rbac_middleware.go     # Granular permission-level access control
│   │       │   ├── requestid_middleware.go# X-Request-ID context propagation
│   │       │   └── tracing_middleware.go  # X-Trace-ID distributed trace propagation
│   │       └── router.go           # Central mux, route definitions, and middleware chaining
│   ├── domain/
│   │   ├── claims.go               # JWT claims schema definition
│   │   ├── errors.go               # Central domain errors (ErrNotFound, ErrConflict, etc.)
│   │   ├── pagination.go           # Standard pagination models (Page, Limit, TotalRows)
│   │   ├── permission.go           # RBAC permissions & role-to-permission mapping
│   │   ├── role.go                 # Role enum types (ADMIN, USER)
│   │   └── user.go                 # Core domain user entity & interface contracts
│   ├── infrastructure/
│   │   ├── cache/
│   │   │   └── redis.go            # Production connection-pooled Redis client
│   │   ├── database/
│   │   │   ├── migrator.go         # golang-migrate integration with clean teardown
│   │   │   └── postgres.go         # High-concurrency tuned pgxpool.Pool
│   │   ├── logger/
│   │   │   ├── audit.go            # SOC 2-compliant security audit trail logger
│   │   │   └── zap.go              # Production JSON logger (Uber Zap)
│   │   └── metrics/
│   │       └── metrics.go          # Prometheus metrics registry (Four Golden Signals)
│   ├── repository/
│   │   ├── postgres/
│   │   │   └── user_repo.go        # PostgreSQL implementation with constraint safety
│   │   └── redis/
│   │       └── session_repo.go     # Refresh token lifecycle & session revocation
│   └── usecase/
│       ├── auth_usecase.go         # Auth orchestration, bcrypt hashing, rotation logic
│       ├── auth_usecase_test.go    # Unit tests with testify & interface mocks
│       └── user_usecase.go         # User query use case logic
├── migrations/
│   ├── 000001_init_schema.down.sql # Reversible rollback schema
│   └── 000001_init_schema.up.sql   # Schema DDL with email index
├── pkg/
│   └── utils/
│       ├── jwt.go                  # RSA-4096 Key Manager, token issuance, JWKS builder
│       ├── response.go             # Unified JSON response envelopes
│       └── validator.go            # Input sanitization, email & password constraints
├── .env.example                    # Environment variable template
├── .gitignore                      # Exhaustive security and build artifact ignore list
├── docker-compose.yml              # Multi-container orchestration stack
├── Dockerfile                      # Production multi-stage, non-root containerfile
├── go.mod                          # Go module dependencies
├── go.sum                          # Cryptographic checksums
└── README.md                       # Service documentation
```

---

## Enterprise Production Highlights

1. **Asymmetric RS256 Authentication & JWKS (`/.well-known/jwks.json`):**
   Tokens are signed using an RSA-4096 private key. Any downstream microservice (e.g., Billing, Orders) or API Gateway can verify JWTs in memory by fetching public keys from the standard RFC 7517 JWKS endpoint without ever touching your database or private keys.
2. **Short-Lived Access Tokens & Refresh Rotation:**
   Access tokens expire after 15 minutes. Opaque 256-bit cryptographic refresh tokens are tracked in Redis. When refreshing, the old refresh token is immediately deleted and rotated to prevent replay attacks.
3. **Defense Against TOCTOU Race Conditions:**
   Eliminates check-then-insert antipatterns. Concurrency safety is enforced at the database level by mapping PostgreSQL error code `23505` (`unique_violation`) directly to `domain.ErrConflict`.
4. **Distributed Redis Rate Limiter:**
   Executes an atomic token-bucket script in Redis. Works reliably across multiple horizontal Kubernetes pods, supporting proxy-aware client IP extraction (`X-Forwarded-For`, `X-Real-IP`).
5. **Denial-of-Service (DoS) / OOM Boundaries:**
   All incoming request bodies are bound by `http.MaxBytesReader(w, r.Body, 1<<20)` (1 MB ceiling), neutralizing memory exhaustion and Slowloris attacks.
6. **No Mass-Assignment / Privilege Escalation:**
   Public registration strictly enforces the `USER` role. Admin roles can only be granted via database seeding or internal administrative channels.
7. **Real Kubernetes Dependency Healthchecks:**
   `/healthz` serves as a liveness probe; `/readyz` acts as a true readiness probe that pings both PostgreSQL connection pools and Redis with a 2-second timeout before accepting traffic.
8. **Observability & Audit Compliance:**
   - Real-time Prometheus `/metrics` exposing request count, status code breakdowns, active concurrent requests, and latency histograms ($p50, p95, p99$).
   - Immutable security audit logger (`SECURITY_AUDIT_TRAIL`) recording logins, logouts, refresh events, and registration attempts for SOC 2 and ISO 27001 readiness.
   - Structured JSON logs via Uber Zap with automated `X-Request-ID` and `X-Trace-ID` context propagation.

---

## Security & Hardening Matrix

| Threat / Requirement | Architectural Mitigation | Status |
| :--- | :--- | :---: |
| **Credential Leakage** | Hardcoded passwords removed; fail-fast configuration boot check. | Verified |
| **Privilege Escalation** | Public registration DTO excludes role selection; defaults strictly to `USER`. | Verified |
| **Database Race Conditions** | Direct insert mapping PostgreSQL error code `23505` (unique constraint). | Verified |
| **Payload Flooding / OOM** | `http.MaxBytesReader` restricts payloads to $\le 1\text{ MB}$. | Verified |
| **Cross-Service Compromise** | Asymmetric RSA-4096 tokens; downstreams only hold the public key via JWKS. | Verified |
| **Multi-Pod Rate Limit Bypass** | Distributed Redis Lua script execution across horizontal replicas. | Verified |
| **Token Replay / Hijacking** | Short 15m expiration + one-time refresh token rotation with immediate Redis invalidation. | Verified |
| **Audit Compliance** | Structured security logs with client IP, action type, timestamp, and actor ID. | Verified |

---

## Technology Stack

- **Language:** Go `1.22+`
- **Primary Database:** PostgreSQL 16 (via `github.com/jackc/pgx/v5` connection pool)
- **Cache & Session Store:** Redis 7 (via `github.com/redis/go-redis/v9`)
- **Cryptography & Tokens:** `golang-jwt/jwt/v5` (RS256 / RSA-4096) & `golang.org/x/crypto/bcrypt`
- **Database Migrations:** `github.com/golang-migrate/migrate/v4`
- **Logging:** Uber Zap (`go.uber.org/zap`)
- **Metrics & Telemetry:** Prometheus Client (`github.com/prometheus/client_golang`)
- **Testing & Assertions:** `github.com/stretchr/testify`

---

## Prerequisites

Ensure you have the following installed locally:
- **Go 1.22** or higher ([golang.org](https://golang.org/))
- **Docker & Docker Compose** ([docker.com](https://www.docker.com/))
- **cURL** (or an API testing tool like Postman, Bruno, or Thunder Client)

---

## Configuration Reference

All application parameters are injected via environment variables. Create a `.env` file in the project root based on the following reference:

```env
# Application Server
APP_PORT=8080

# PostgreSQL Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=nethru2002
DB_NAME=enterprise_db
DB_SSLMODE=disable

# PostgreSQL Connection Pool Tuning
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_IDLE=15m
DB_MAX_CONN_LIFE=1h

# Redis Session & Rate Limiting Cache
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Optional: Path to custom RSA private PEM key.
# If omitted, the service auto-generates a secure 4096-bit RSA key pair at startup.
RSA_PRIVATE_KEY_PATH=
```

---

## Installation & Quickstart

### 1. Clone the Repository
```bash
git clone https://github.com/Nethru2002/Enterprise-User-Management-Authentication-Microservice.git
cd Enterprise-User-Management-Authentication-Microservice/enterprise-go-service
```

### 2. Configure Environment
```bash
cp .env.example .env
```

### 3. Spin Up Infrastructure (Postgres & Redis)
Run PostgreSQL and Redis via Docker:

```powershell
# In PowerShell:
docker run --name enterprise-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=nethru2002 -e POSTGRES_DB=enterprise_db -p 5432:5432 -d postgres:16-alpine
docker run --name enterprise-redis -p 6379:6379 -d redis:7-alpine
```

### 4. Tidy and Download Dependencies
```bash
go mod tidy
go mod verify
```

### 5. Start the Server
```powershell
# In PowerShell:
$env:APP_PORT="8080"
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="nethru2002"
$env:DB_NAME="enterprise_db"
$env:DB_SSLMODE="disable"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="6379"
$env:REDIS_PASSWORD=""

go run cmd/server/main.go
```

The server automatically applies pending schema migrations at boot and binds to port `8080`.

---

## Running Automated Migrations

The microservice runs migrations automatically on boot. However, you can also execute migrations independently using the standalone CLI runner:

```bash
go run cmd/migrate/main.go
```

---

## Testing & Quality Assurance

Run the unit test suite with mock repository validations and code coverage:

```bash
go test -v -cover ./...
```

To run test coverage profiling:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## API Endpoints Reference

### Infrastructure & SRE Probes
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Kubernetes liveness probe (checks if process is alive) | No |
| `GET` | `/readyz` | Kubernetes readiness probe (actively pings Postgres & Redis) | No |
| `GET` | `/metrics` | Prometheus metrics scrape target (latency, RPS, gauges) | No |
| `GET` | `/.well-known/jwks.json` | RFC 7517 JWKS public keys for downstream services | No |

### Public Authentication
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Register a new user account (defaults to `USER` role) | No |
| `POST` | `/api/v1/auth/login` | Authenticate and issue RS256 access & refresh tokens | No |
| `POST` | `/api/v1/auth/refresh` | Rotate and issue a new access token via refresh token | No |
| `POST` | `/api/v1/auth/logout` | Revoke active refresh session in Redis | No |

### Protected Endpoints
| Method | Endpoint | Description | Permissions Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/users/profile` | Retrieve the authenticated user's profile | `Bearer Token` (`read:profile`) |
| `GET` | `/api/v1/admin/users` | List users with pagination (`page`, `limit`) | `Bearer Token` (`manage:system` / `ADMIN`) |

---

## API Walkthrough & Testing Guide

Use `curl.exe` (Windows PowerShell) or `curl` (Linux/macOS) to interact with the service.

### 1. Verify Probes
```bash
curl.exe -i http://localhost:8080/healthz
curl.exe -i http://localhost:8080/readyz
```

### 2. Discover Public JWKS
```bash
curl.exe -i http://localhost:8080/.well-known/jwks.json
```

### 3. Register a Regular User
```bash
curl.exe -i -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"SuperSecurePassword123"}'
```

### 4. Authenticate & Retrieve Tokens
```bash
curl.exe -i -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"alice@enterprise.com","password":"SuperSecurePassword123"}'
```
*Response Envelope:*
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "63f89836940...",
    "expires_in": 900
  }
}
```

### 5. Fetch Protected Profile
```bash
$TOKEN = "<PASTE_ACCESS_TOKEN_HERE>"

curl.exe -i -X GET http://localhost:8080/api/v1/users/profile `
  -H "Authorization: Bearer $TOKEN"
```

### 6. Create and Test an Admin Account
To establish an admin account securely without opening public registration loopholes:

1. Register an account:
   ```bash
   curl.exe -i -X POST http://localhost:8080/api/v1/auth/register `
     -H "Content-Type: application/json" `
     -d '{"email":"admin@enterprise.com","password":"AdminSecurePassword123"}'
   ```
2. Promote the user directly in PostgreSQL:
   ```bash
   docker exec -it enterprise-postgres psql -U postgres -d enterprise_db -c "UPDATE users SET role = 'ADMIN' WHERE email = 'admin@enterprise.com';"
   ```
3. Authenticate to receive an Admin token:
   ```bash
   curl.exe -i -X POST http://localhost:8080/api/v1/auth/login `
     -H "Content-Type: application/json" `
     -d '{"email":"admin@enterprise.com","password":"AdminSecurePassword123"}'
   ```
4. Query the paginated admin endpoint:
   ```bash
   $ADMIN_TOKEN = "<PASTE_ADMIN_ACCESS_TOKEN_HERE>"

   curl.exe -i -X GET "http://localhost:8080/api/v1/admin/users?page=1&limit=10" `
     -H "Authorization: Bearer $ADMIN_TOKEN"
   ```

### 7. Rotate Tokens
```bash
curl.exe -i -X POST http://localhost:8080/api/v1/auth/refresh `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<PASTE_REFRESH_TOKEN_HERE>"}'
```

### 8. Revoke Session (Logout)
```bash
curl.exe -i -X POST http://localhost:8080/api/v1/auth/logout `
  -H "Content-Type: application/json" `
  -d '{"refresh_token":"<PASTE_REFRESH_TOKEN_HERE>"}'
```

---

## Observability & Monitoring

### 1. Prometheus Telemetry (`/metrics`)
The service exposes native Prometheus metrics instrumenting the **Four Golden Signals**:
- `http_requests_total{method, path, status}`: Request volume and HTTP status breakdown.
- `http_request_duration_seconds{method, path}`: Latency histogram for measuring $p50, p95, p99$ response times.
- `http_active_requests`: Real-time gauge of in-flight requests.
- Standard Go runtime metrics (goroutine count, GC pause durations, heap allocations).

### 2. Structured JSON Logging & Tracing
All logs are structured JSON emitted via Uber Zap:
- Automated request tracing via `X-Request-ID` and `X-Trace-ID`.
- Incoming HTTP requests, latency, status codes, and error traces.

### 3. Security Audit Trails
Security-critical events are captured under `SECURITY_AUDIT_TRAIL`:
```json
{
  "level": "INFO",
  "logger": "audit",
  "msg": "SECURITY_AUDIT_TRAIL",
  "event_type": "AUTH_LOGIN_SUCCESS",
  "actor_email": "admin@enterprise.com",
  "actor_ip": "127.0.0.1:58432",
  "detail": "successful authentication",
  "timestamp": "2026-09-16T05:58:00Z"
}
```

---

## Docker & Containerized Deployment

### 1. Multi-Stage Dockerfile
The included `Dockerfile` utilizes a two-stage build:
1. **Builder stage:** Compiles binaries statically (`CGO_ENABLED=0`) with symbol stripping (`-ldflags="-w -s"`).
2. **Runtime stage:** Runs inside a hardened, non-root `alpine:3.19` container for a minimal attack surface.

### 2. Running via Docker Compose
To spin up the microservice, PostgreSQL, and Redis together with automatic parameter injection from `.env`:

```bash
docker compose up -d --build
```

Verify service status:
```bash
docker compose ps
```

Tear down the stack:
```bash
docker compose down
```

---

## Production Best Practices & Operations

1. **Persistent RSA Key Provisioning:**  
   In multi-pod Kubernetes deployments, supply an explicit `RSA_PRIVATE_KEY_PATH` mounted from a Kubernetes Secret or Cloud Vault (AWS Secrets Manager, HashiCorp Vault) so all replicas share the same signing key across restarts.
2. **Reverse Proxy & Ingress Settings:**  
   When deploying behind an Ingress controller (AWS ALB, NGINX, Cloudflare), ensure proxy headers (`X-Forwarded-For`, `X-Real-IP`) are forwarded so the distributed Redis rate limiter correctly tracks client IPs.
3. **Database Read Replicas (High Scale):**  
   For workloads exceeding tens of thousands of requests per second, the `UserRepository` can be extended with separate primary (write) and replica (read) connection pools.

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
