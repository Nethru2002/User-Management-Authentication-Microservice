# Enterprise User Management & Authentication Microservice

An enterprise-grade, production-ready user management and authentication microservice built with Go (Golang). It strictly implements Clean Architecture, the Standard Go Project Layout, and enterprise infrastructure best practices.

---

## Architecture & Design Patterns

- **Clean Architecture:** Strict separation of concerns divided into Domain, Use Case, Repository, Infrastructure, and Delivery layers.
- **Enterprise Layout:** Standard modular directory structure (`cmd/`, `internal/`, `pkg/`, `migrations/`).
- **Resilience & Observability:** Graceful shutdowns, structured logging via Uber Zap, and distributed tracing headers (Request-ID / Trace-ID).
- **Security:** JWT-based authentication, bcrypt password hashing, sliding window rate limiting, CORS configuration, and Role-Based Access Control (RBAC).
- **Database & Caching:** PostgreSQL persistence managed via `pgxpool` with automated versioned migrations (`golang-migrate`), and Redis caching integration.

---

## Project Structure

```text
enterprise-go-service/
├── cmd/
│   ├── migrate/
│   │   └── main.go
│   └── server/
│       └── main.go
├── configs/
│   └── config.yaml
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth_handler.go
│   │       │   └── user_handler.go
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go
│   │       │   ├── cors_middleware.go
│   │       │   ├── logger_middleware.go
│   │       │   ├── ratelimit_middleware.go
│   │       │   ├── requestid_middleware.go
│   │       │   └── tracing_middleware.go
│   │       └── router.go
│   ├── domain/
│   │   ├── claims.go
│   │   ├── errors.go
│   │   ├── role.go
│   │   └── user.go
│   ├── infrastructure/
│   │   ├── cache/
│   │       │   └── redis.go
│   │   ├── database/
│   │       │   ├── migrator.go
│   │       │   └── postgres.go
│   │       └── logger/
│   │           └── zap.go
│   ├── repository/
│   │   ├── postgres/
│   │       │   └── user_repo.go
│   │       └── redis/
│   │           └── cache_repo.go
│   └── usecase/
│       ├── auth_usecase.go
│       ├── auth_usecase_test.go
│       └── user_usecase.go
├── migrations/
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── pkg/
│   └── utils/
│       └── jwt.go
├── .env
├── .gitignore
├── go.mod
└── go.sum
```

---

## Prerequisites

- Go (v1.22 or higher)
- PostgreSQL
- Redis
- Docker (Optional, for running containers)

---

## Environment Configuration

Create a `.env` file in the root directory:

```env
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=enterprise_db
DB_SSLMODE=disable
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=super-secret-enterprise-key-change-me
```

---

## Getting Started & Execution

1. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run Database Migrations:**
   ```bash
   go run cmd/migrate/main.go
   ```

3. **Start the Production Server:**
   - **PowerShell (Windows):**
     ```powershell
     $env:APP_PORT="8080"
     $env:DB_HOST="localhost"
     $env:DB_PORT="5432"
     $env:DB_USER="postgres"
     $env:DB_PASSWORD="your_password"
     $env:DB_NAME="enterprise_db"
     $env:DB_SSLMODE="disable"
     $env:REDIS_HOST="localhost"
     $env:REDIS_PORT="6379"
     $env:REDIS_PASSWORD=""
     $env:JWT_SECRET="super-secret-enterprise-key-change-me"

     go run cmd/server/main.go
     ```
   - **Linux / macOS / Git Bash:**
     ```bash
     export $(cat .env | xargs)
     go run cmd/server/main.go
     ```

---

## Testing

Execute all unit test suites with code coverage metrics:
```bash
go test -v -cover ./...
```

---

## API Endpoints Reference

### Public Routes
- **Health Check:** `GET /healthz`
- **Readiness Check:** `GET /readyz`
- **Register User:** `POST /api/v1/auth/register`
  - *Payload:* `{"email": "user@example.com", "password": "securepassword", "role": "USER"}`
- **Login:** `POST /api/v1/auth/login`
  - *Payload:* `{"email": "user@example.com", "password": "securepassword"}`

### Protected Routes (Requires Bearer Token)
- **Get Profile:** `GET /api/v1/users/profile`
  - *Headers:* `Authorization: Bearer <JWT_TOKEN>`