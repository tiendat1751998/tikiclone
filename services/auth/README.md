# Auth Service

Centralized authentication and authorization service for the Tiki Clone platform.

## Architecture

### Layer Structure (DDD)

```
cmd/server/main.go          — Entry point, dependency wiring, graceful shutdown
internal/
  config/                   — Environment-based configuration
  domain/                   — Core domain types, errors, audit events
  application/              — AuthService: business logic orchestration
  infrastructure/
    mysql/                  — MySQL repositories (user, session, audit)
    redis/                  — Redis store (sessions, blacklists, reset/verify tokens)
    jwt/                    — JWT token generation and validation
    hash/                   — Password hashing (Argon2id / bcrypt)
  security/                 — Rate limiter, suspicious activity detector
  transport/
    http/                   — Gin HTTP handlers, router, middleware
    grpc/                   — gRPC server for inter-service auth
  metrics/                  — Prometheus metrics
  tracing/                  — OpenTelemetry setup
```

### Data Flow

```
Client → [HTTP/gRPC] → Handler → AuthService → Repositories/Redis/JWT → MySQL/Redis
```

### Authentication Flow

1. **Register** — User submits email/password → validated → hashed (Argon2id) → stored in MySQL → JWT tokens issued → session created
2. **Login** — Email/password verified → rate limited → account lockout check → suspicious location detection → JWT tokens issued
3. **Token Refresh** — Refresh token validated → rotation check (reuse detection) → new token pair issued
4. **Logout** — Session revoked → tokens blacklisted in Redis → refresh token revoked
5. **Password Reset** — Request with email → reset token stored in Redis (SHA-256 hashed, 15min TTL) → reset validates token → password changed → all sessions revoked
6. **Email Verification** — Verify token validated → user marked verified → token consumed

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/auth/register | Register new user | No |
| POST | /api/v1/auth/login | Login | No |
| POST | /api/v1/auth/refresh | Refresh token pair | No |
| POST | /api/v1/auth/logout | Logout single session | Yes |
| POST | /api/v1/auth/logout/all | Logout all sessions | Yes |
| GET | /api/v1/auth/sessions | List active sessions | Yes |
| DELETE | /api/v1/auth/sessions/:session_id | Revoke specific session | Yes |
| GET | /api/v1/auth/profile | Get user profile | Yes |
| POST | /api/v1/auth/validate | Validate access token | Yes |
| POST | /api/v1/auth/password-reset/request | Request password reset | No |
| POST | /api/v1/auth/password-reset/reset | Execute password reset | No |
| POST | /api/v1/auth/verify-email | Verify email with token | No |
| POST | /api/v1/auth/verify-email/send | Send verification email | Yes |

### Health Endpoints

| Path | Purpose |
|------|---------|
| GET /health | Liveness probe (process alive) |
| GET /ready | Readiness probe (dependencies healthy) |
| GET /startup | Startup probe (initialization complete) |
| GET /metrics | Prometheus metrics |

## Security

- **Password Hashing**: Argon2id (configurable memory/time/threads) or bcrypt fallback
- **JWT**: RS256-signed access (15min) and refresh (7d) tokens with rotation and blacklisting
- **Rate Limiting**: Per-IP login (5/5min), register (3/hour), password reset (3/hour)
- **Account Lockout**: 10 failed attempts → 30min lockout
- **Session Management**: Max 10 concurrent sessions, idle timeout, refresh rotation
- **Token Blacklisting**: Redis-backed blacklist with TTL
- **Suspicious Detection**: New-location login detection with configurable threshold
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, HSTS, XSS-Protection, Referrer-Policy, Permissions-Policy
- **Request Sanitization**: Content-Type validation on write methods
- **Audit Logging**: All auth events logged to MySQL (login, register, logout, password changes, suspicious activity)

## Configuration

All configuration via environment variables. See `internal/config/config.go` for full list.

Key variables:
- `AUTH_HTTP_PORT` — HTTP server port (default: 8080)
- `AUTH_GRPC_PORT` — gRPC server port (default: 9090)
- `MYSQL_*` — MySQL connection settings
- `REDIS_*` — Redis connection settings
- `JWT_*` — JWT secrets and TTLs
- `PASSWORD_ALGORITHM` — Hashing algorithm (argon2/bcrypt)

## Probes

- **Liveness** (`/health`): Basic process health
- **Readiness** (`/ready`): Checks MySQL and Redis connectivity
- **Startup** (`/startup`): Initialization complete check

## Observability

- **Tracing**: OpenTelemetry with configurable exporter endpoint and sample ratio
- **Metrics**: Prometheus (request duration, error rates, business metrics)
- **Logging**: Structured zap logger with trace context injection

## Dependencies

- MySQL (primary data store)
- Redis (sessions, rate limiting, token blacklisting, password reset tokens)
- Go 1.22+

## Related Services

- **API Gateway**: Routes external requests to auth service
- **gRPC**: Internal services validate tokens via gRPC `ValidateToken` RPC
