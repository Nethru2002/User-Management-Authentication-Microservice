Here is a complete, enterprise-grade `README.md` for your project. You can replace the contents of your `iam-microservice/README.md` (or root `README.md`) with this.

---

```markdown
# Enterprise Identity & Access Management (IAM) Microservice

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-brightgreen.svg)]()
[![Security](https://img.shields.io/badge/Security-RS256%20%7C%20JWKS%20%7C%20RBAC-red.svg)]()
[![Observability](https://img.shields.io/badge/Observability-Prometheus%20%7C%20Zap%20Audit-orange.svg)]()
[![Database](https://img.shields.io/badge/Storage-PostgreSQL%20(pgxpool)%20%7C%20Redis-blueviolet.svg)]()
[![License](https://img.shields.io/badge/License-MIT-lightgrey.svg)](LICENSE)

An enterprise-grade, high-throughput **Identity and Access Management (IAM) Microservice** built with **Go** following strict **Clean Architecture** principles and the official **Standard Go Project Layout**. 

Engineered for zero-trust architectures, Kubernetes multi-pod deployments, high-concurrency workloads, and SOC 2 / ISO 27001 audit compliance.

---

## 🏛️ Architecture & System Design

The service implements a decoupled, four-tier **Clean Architecture** model where domain entities, use-case business rules, and interface boundaries remain completely agnostic of external drivers, frameworks, and storage engines.

```text
enterprise-go-service/
├── cmd/
│   ├── migrate/                     # Standalone migration execution tool
│   └── server/                      # Production service runtime entrypoint
├── internal/
│   ├── config/                      # Strongly typed environment configuration & validation
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/             # Presentation layer HTTP route handlers
│   │       ├── middleware/          # Security, rate limiting, logging, tracing, metrics
│   │       └── router.go            # ServeMux route registration & pipeline assembly
│   ├── domain/                      # Core entities, interfaces, roles, permissions, errors
│   ├── infrastructure/
│   │   ├── cache/                   # Pooled Redis client connection driver
│   │   ├── database/                # Tuned PostgreSQL pgxpool driver & migrator
│   │   ├── logger/                  # Uber Zap structured logger & dedicated audit logger
│   │   └── metrics/                 # Prometheus telemetry collector & metric registrations
│   ├── repository/
│   │   ├── postgres/                # PostgreSQL implementation of UserRepository
│   │   └── redis/                   # Redis implementation of SessionRepository (tokens)
│   └── usecase/                     # Core business logic workflows & unit tests
├── migrations/                      # Version-controlled idempotent SQL schema migrations
├── pkg/
│   └── utils/                       # Shared utilities (RS256/JWKS, envelopes, validators)
├── .env.example                     # Production environment variable template
├── go.mod                           # Go module definition
└── README.md                        # Documentation
```

---

## 🚀 Enterprise Feature Highlights

### 1. Asymmetric Cryptography & Public Discovery (RS256 + JWKS)
* Tokens are signed using **RSA-4096 (RS256)** private keys.
* Exposes a public standard **RFC 7517 JSON Web Key Set (JWKS)** endpoint at `GET /.well-known/jwks.json`.
* Downstream microservices and API gateways can dynamically fetch and cache public keys to verify authenticity without knowing the private signing key or reading the central IAM database.

### 2. Ephemeral Sessions & Token Rotation
* **Access Tokens:** Short-lived (15 minutes), containing custom claims, user roles, and Key ID (`kid`).
* **Refresh Tokens:** Opaque, high-entropy random byte tokens stored in Redis with a 7-day TTL.
* **Token Rotation:** Every token refresh automatically invalidates the consumed refresh token and issues a new pair, preventing replay attacks.
* **Instant Revocation:** Logout explicitly purges session state from Redis.

### 3. Distributed Token-Bucket Rate Limiting
* Uses an atomic **Redis Lua Script** to enforce strict request budgets (e.g., 100 requests/minute per client).
* Supports multi-replica cluster environments without relying on node-local in-memory maps.
* Proxy-aware client identification using `X-Forwarded-For` and `X-Real-IP`.

### 4. Zero-Trust Hardening & Data Integrity
* **No Unbounded Memory (Slowloris / OOM Protection):** Incoming payloads are limited to 1 MB using `http.MaxBytesReader`.
* **Privilege Escalation Prevention:** Public registration strictly forces the `USER` role; administrative accounts must be created or promoted internally.
* **TOCTOU Elimination:** Solves Time-of-Check to Time-of-Use race conditions via PostgreSQL `23505` (`unique_violation`) constraint mappings.
* **Fail-Fast Configuration:** The service aborts boot if required secrets or connection configurations are absent.

### 5. Production Observability & SRE Primitives
* **Prometheus Metrics (`/metrics`):** Exposes HTTP request volume, status codes, latency histograms ($p50, p95, p99$), and active concurrent requests.
* **Structured Security Audit Logs:** Emits dedicated audit logs (`SECURITY_AUDIT_TRAIL`) with actor emails, timestamps, client IPs, and event types for compliance.
* **Distributed Tracing:** Auto-injects and propagates `X-Request-ID` and `X-Trace-ID` through HTTP headers, contexts, and log fields.
* **True Readiness Checks:** `/readyz` explicitly executes non-blocking health pings across PostgreSQL connection pools and Redis instances.

---

## 🛠️ Technology Stack

| Layer | Component / Library |
| :--- | :--- |
| **Language** | Go 1.22+ |
| **Database Pool** | `github.com/jackc/pgx/v5/pgxpool` |
| **Caching & Sessions**| `github.com/redis/go-redis/v9` |
| **Token Cryptography**| `github.com/golang-jwt/jwt/v5` (RS256) & `crypto/rsa` |
| **Password Hashing** | `golang.org/x/crypto/bcrypt` (Cost: 10) |
| **Telemetry / SRE** | `github.com/prometheus/client_golang` (Prometheus) |
| **Logging** | `go.uber.org/zap` |
| **Migrations** | `github.com/golang-migrate/migrate/v4` |
| **Testing** | `github.com/stretchr/testify` |

---

## ⚙️ Environment Configuration

Create a `.env` file in the project root:

```env
# Application Server
APP_PORT=8080

# PostgreSQL Connection & Pool Settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=enterprise_db
DB_SSLMODE=disable
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_IDLE=15m
DB_MAX_CONN_LIFE=1h

# Redis Session & Cache Settings
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Asymmetric Cryptography (Optional: path to PEM file; generates in-memory if empty)
RSA_PRIVATE_KEY_PATH=
```

---

## 📦 Getting Started

### 1. Start Infrastructure Dependencies (Docker)

If you don't have PostgreSQL and Redis installed locally, spin them up with Docker:

```bash
# Start PostgreSQL (Database: enterprise_db)
docker run --name enterprise-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_secure_password \
  -e POSTGRES_DB=enterprise_db \
  -p 5432:5432 -d postgres:16-alpine

# Start Redis
docker run --name enterprise-redis \
  -p 6379:6379 -d redis:7-alpine
```

### 2. Install & Verify Go Modules

```bash
go mod tidy
go mod verify
```

### 3. Run Unit Tests

Execute the test suite with coverage:

```bash
go test -v -cover ./...
```

### 4. Start the Application Server

The service validates and executes all schema migrations located in `migrations/` automatically on startup:

```bash
go run cmd/server/main.go
```

---

## 📡 API Reference & Endpoints

### Public Observability & Probes

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/healthz` | Kubernetes Liveness Probe (Returns `200 OK`) |
| `GET` | `/readyz` | Kubernetes Readiness Probe (Validates DB & Redis connectivity) |
| `GET` | `/metrics` | Prometheus telemetry scrape target |
| `GET` | `/.well-known/jwks.json` | RFC 7517 Public Key Set for external token validation |

### Authentication & Lifecycle Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Register a new user (Enforces `RoleUser`) |
| `POST` | `/api/v1/auth/login` | Authenticate with credentials (Returns RS256 + Refresh tokens) |
| `POST` | `/api/v1/auth/refresh` | Rotate expired Access Token using active Refresh Token |
| `POST` | `/api/v1/auth/logout` | Revoke active Refresh Token from Redis |

### Protected Resource Endpoints

| Method | Endpoint | Required Permission | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/users/profile` | `read:profile` | Fetch authenticated user profile |
| `GET` | `/api/v1/admin/users` | `manage:system` | Paginated listing of all registered users (Admin only) |

---

## 💻 Manual Verification & cURL Guide

*(Note: On Windows PowerShell, invoke commands using `curl.exe` to avoid conflicts with native cmdlets).*

### 1. Register a User Account
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@enterprise.com",
    "password": "SuperSecretPassword123"
  }'
```

### 2. Log In & Receive Token Pair
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@enterprise.com",
    "password": "SuperSecretPassword123"
  }'
```
*Response Body:*
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "8dfa2b109e432c...",
    "expires_in": 900
  }
}
```

### 3. Fetch Protected Profile
```bash
curl -i -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

### 4. Rotate Access Token
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<YOUR_REFRESH_TOKEN>"
  }'
```

### 5. Revoke Session (Logout)
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<YOUR_REFRESH_TOKEN>"
  }'
```

### 6. Create and Test an Admin Account
To access administrative routes, create or promote a user in PostgreSQL:

1. Register `admin@enterprise.com`:
   ```bash
   curl -i -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "admin@enterprise.com", "password": "AdminPassword123"}'
   ```
2. Promote the user directly in the database:
   ```bash
   docker exec -it enterprise-postgres psql -U postgres -d enterprise_db -c "UPDATE users SET role = 'ADMIN' WHERE email = 'admin@enterprise.com';"
   ```
3. Authenticate to retrieve an admin token, then query the admin endpoint:
   ```bash
   curl -i -X GET "http://localhost:8080/api/v1/admin/users?page=1&limit=5" \
     -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
   ```

---

## 🔒 Security Best Practices for Production

When transitioning from local development to production Kubernetes/Cloud environments:

1. **Persistent RSA Key Pairs:** Mount a persistent private key file (`RSA_PRIVATE_KEY_PATH=/etc/secrets/private_key.pem`) from a secret manager (HashiCorp Vault, AWS Secrets Manager, Kubernetes Secrets) so multiple pods share the identical signing identity across rolling restarts.
2. **Reverse Proxy Configuration:** When deploying behind an AWS ALB, Cloudflare, or NGINX Ingress, ensure the proxy sends valid `X-Forwarded-For` and `X-Real-IP` headers so rate-limiting reflects real client IPs.
3. **Database Connection Tuning:** Fine-tune `DB_MAX_CONNS` and `DB_MIN_CONNS` according to available database instance cores and connection limits.

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.
```