# Enterprise User Management & Authentication Microservice

An advanced, production-grade, highly scalable user management and authentication microservice built with **Go (Golang)** following strict **Clean Architecture** principles and the official **Standard Go Project Layout**.

---

## Architecture Overview

This project implements a decoupled, multi-tier clean architecture pattern to ensure separation of concerns, high testability, and enterprise-grade maintainability:

```text
enterprise-go-service/
├── cmd/
│   ├── server/           # Application server bootstrap entrypoint
│   └── migrate/          # Standalone database migration runner
├── configs/              # Configuration templates
├── internal/
│   ├── delivery/         # Presentation layer (HTTP routers, handlers, middleware)
│   ├── domain/           # Core enterprise business models, entities, and interfaces
│   ├── infrastructure/   # External drivers (PostgreSQL pools, Redis clients, Zap loggers, Migrators)
│   ├── repository/       # Data-access layer (Postgres queries, Redis caching/session tracking)
│   └── usecase/          # Core business logic and use-case implementations
├── migrations/           # SQL migration version files (up/down)
├── pkg/                  # Public shared utilities (JWT token management, JSON response envelopes)
├── .env                  # Environment variables
├── go.mod                # Go module definitions
└── README.md             # Project documentation
```

---

## Core Features & Enterprise Capabilities

1. **Clean Architecture & Separation of Concerns:** Domain logic remains entirely independent of databases, caching layers, and HTTP frameworks.
2. **Advanced JWT Authentication & RBAC:** Secure password-hashing via `bcrypt`, custom JWT claims, token generation/validation, and granular permission-based resource protection (`read:profile`, `manage:system`).
3. **Session Tracking & Caching:** Integrated high-performance Redis client for caching and session management.
4. **Automated Database Migrations:** Version-controlled database schema management powered by `golang-migrate`.
5. **Production Observability & Resilience:**
   - **Structured JSON Logging:** Powered by Uber `zap`.
   - **Distributed Tracing:** Automated `X-Request-ID` and `X-Trace-ID` generation and context injection.
   - **Health Probes:** Kubernetes-ready `/healthz` (liveness) and `/readyz` (readiness) endpoints.
   - **Graceful Shutdown:** Intercepts `SIGINT` / `SIGTERM` signals to drain connections and prevent data corruption.
6. **Security & Traffic Control:** Built-in sliding-window token-bucket IP rate-limiter and strict CORS middleware.
7. **Robust Pagination & Standardized Envelopes:** Uniform JSON response formatting (`success`, `data`, `error`, `message`) paired with offset-limit pagination logic.
8. **Unit Testing & Mocking:** Comprehensive unit test suite utilizing `testify` and interface mocks.

---

## Technology Stack

- **Language:** Go 1.22+
- **Database:** PostgreSQL (via `pgx/v5` connection pool)
- **Cache / Session Store:** Redis (via `go-redis/v9`)
- **Authentication:** `golang-jwt/v5` & `golang.org/x/crypto/bcrypt`
- **Logging:** Uber `zap`
- **Migrations:** `golang-migrate`
- **Testing:** `stretchr/testify`

---

## Getting Started & Prerequisites

Ensure you have the following installed on your system:
- [Go (1.22 or higher)](https://golang.org/)
- [PostgreSQL](https://www.postgresql.org/)
- [Docker](https://www.docker.com/) (Optional, for running Redis)

---

## Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/enterprise-go-service.git
   cd enterprise-go-service
   ```

2. **Configure Environment Variables:**
   Create a `.env` file in the project root directory:
   ```env
   APP_PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=nethru2002
   DB_NAME=enterprise_db
   DB_SSLMODE=disable
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   JWT_SECRET=super-secret-enterprise-key-change-me
   ```

3. **Start Redis (via Docker):**
   ```bash
   docker run --name enterprise-redis -p 6379:6379 -d redis:alpine
   ```

4. **Tidy Dependencies:**
   ```bash
   go mod tidy
   ```

---

## Running Migrations & The Application

1. **Run Database Migrations:**
   ```bash
   go run cmd/migrate/main.go
   ```

2. **Run Unit Tests with Coverage:**
   ```bash
   go test -v -cover ./...
   ```

3. **Start the Production Server:**
   ```powershell
   # PowerShell Environment Setup & Execution
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
   $env:JWT_SECRET="super-secret-enterprise-key-change-me"

   go run cmd/server/main.go
   ```

---

## API Endpoints Reference

### Public Routes
- `GET /healthz` — Service liveness check
- `GET /readyz` — Service readiness check
- `POST /api/v1/auth/register` — Register a new user account
- `POST /api/v1/auth/login` — Authenticate and receive a JWT token

### Protected Routes (Requires Bearer Token)
- `GET /api/v1/users/profile` — Fetch logged-in user profile (Requires `read:profile` permission)
- `GET /api/v1/admin/users` — Fetch paginated system user list (Requires `manage:system` permission / ADMIN role)

---

## Example API Requests

### 1. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@enterprise.com", "password": "securepassword123", "role": "ADMIN"}'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@enterprise.com", "password": "securepassword123"}'
```

### 3. Access Profile (Protected)
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>"
```

---

## License

This project is open-source and available under the [MIT License](LICENSE).