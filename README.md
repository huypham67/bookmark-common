# bookmark-common

Shared Go library for the bookmark microservices. It centralizes the cross-cutting
concerns — JWT auth, rate limiting, logging, database/Redis clients, password
hashing, request/response helpers and code utilities — so each service implements
them once, consistently. Extracted from `bookmark-service-monolithic` and imported
by `user-service` and `bookmark-service`.

This module ships **no `main`, no binary and no migrations of its own** — it is a
library consumed via `go get`.

## Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26 |
| Web framework | Gin | v1.12.0 |
| ORM / DB | GORM (postgres driver) | v1.31.1 / v1.6.0 |
| Cache | go-redis | v9.19.0 |
| Logger | Zerolog | v1.35.1 |
| Auth | JWT (RSA, golang-jwt) | v5.3.1 |
| Migrations | golang-migrate | v4.19.1 |
| Config | envconfig | v1.4.0 |
| Validation | go-playground/validator | v10.30.2 |
| Password hashing | bcrypt (golang.org/x/crypto) | v0.52.0 |
| Testing | Testify + miniredis | v1.11.1 / v2.38.0 |
| Mocks | Mockery (dev tool) | v2 |

## Installation

```bash
go get github.com/huypham67/bookmark-common
go mod tidy
```

## What's Inside

### Middleware

| Middleware | File | Purpose |
|-----------|------|---------|
| `JWTAuth(validator)` | `middleware/jwt.go` | Validates the Bearer token and stores `*jwt.CustomClaims` in the Gin context under `"claims"` |
| `RateLimit(limiter)` | `middleware/ratelimit.go` | Per-client request throttling (Redis-backed), returns `429` when exceeded |

### Packages

| Package | Purpose |
|---------|---------|
| `pkg/base62` | Configurable Base62 encoding/decoding (`NewEncoding`) for short codes |
| `pkg/common` | Shared error definitions/types |
| `pkg/dbutils` | GORM/SQL error mapping helpers |
| `pkg/jwt` | RSA token generator, validator, claims + `GetUserIDFromContext` helper |
| `pkg/jwt/provider` | Loads RSA keys + config and wires the JWT generator/validator (DI) |
| `pkg/logger` | Global Zerolog configuration |
| `pkg/ratelimit` | `Limiter`/`Store` interfaces + Redis-backed implementation |
| `pkg/ratelimit/provider` | Builds a `Limiter` from a Redis client + config (DI) |
| `pkg/redis` | Redis client construction from env config |
| `pkg/requestutils` | Request binding/validation helpers |
| `pkg/response` | Standardized API success/error response formatting |
| `pkg/security` | bcrypt `PasswordHasher` |
| `pkg/shortcode` | Short-code routing: `Classify`, `EncodeSQLCode`, `AddRedisPrefix` |
| `pkg/sqldb` | PostgreSQL (GORM) client + migration runner |
| `pkg/utils` | Random `CodeGenerator` |

All major components are defined as interfaces and ship test mocks (under
`mocks/`) so consuming services can mock them.

## Usage

Each service constructs instances and injects them. Provider packages read
configuration from environment variables under a given prefix.

### JWT auth (provider + middleware)

```go
import (
    "github.com/huypham67/bookmark-common/middleware"
    jwtprovider "github.com/huypham67/bookmark-common/pkg/jwt/provider"
    "github.com/huypham67/bookmark-common/pkg/jwt"
)

// Bootstrap: loads RSA keys + JWT config from env (e.g. AUTH_*).
provider, err := jwtprovider.New("AUTH")
if err != nil { /* handle */ }

// Protect routes:
router.Use(middleware.JWTAuth(provider.Validator()))

// Issue tokens:
token, err := provider.Generator().GenerateToken(userID, displayName, email)

// Inside a handler:
userID, err := jwt.GetUserIDFromContext(c)
```

### Rate limiting (provider + middleware)

```go
import (
    "github.com/huypham67/bookmark-common/middleware"
    ratelimitprovider "github.com/huypham67/bookmark-common/pkg/ratelimit/provider"
    "github.com/huypham67/bookmark-common/pkg/redis"
)

redisClient, _ := redis.NewClient("REDIS")
limiter, _ := ratelimitprovider.New(redisClient, "RATELIMIT")
router.Use(middleware.RateLimit(limiter))
```

### Logger / DB / Redis

```go
import (
    "github.com/huypham67/bookmark-common/pkg/logger"
    "github.com/huypham67/bookmark-common/pkg/sqldb"
)

logger.NewClient("LOG")            // configures the global zerolog logger
db, err := sqldb.NewClient("DB")   // *gorm.DB from env config
```

### Password hashing

```go
import "github.com/huypham67/bookmark-common/pkg/security"

hasher := security.NewBcryptPasswordHasher()
hash, err := hasher.Hash("secret")
err = hasher.Compare(hash, "secret")
```

## Testing

```bash
make test            # run tests + coverage, enforces 80% threshold on business logic
make test-coverage   # open the HTML coverage report
make docker-test     # same run inside Docker (as CI does)
```

- **Verified total coverage: ~98%** on business logic, **80% threshold** enforced by the Makefile.
- **Tested** (business logic): `middleware`, `pkg/base62`, `pkg/jwt`, `pkg/ratelimit`, `pkg/shortcode`.
- **Excluded from the threshold** (infrastructure/wiring, still security-scanned):
  `pkg/common`, `pkg/dbutils`, `pkg/logger`, `pkg/redis`, `pkg/sqldb`,
  `pkg/requestutils`, `pkg/response`, `pkg/security`, `pkg/utils`,
  `pkg/jwt/provider`, `pkg/ratelimit/provider`.

The exclusion lists live in a single source of truth in the `Makefile`
(`INFRA_DIRS` / `SYSTEM_DIRS`) and drive both local coverage filtering and the
SonarCloud `coverage.exclusions`.

### Mocks

`make generate-mocks` regenerates mocks for the interface packages that declare
`go:generate` directives (`pkg/jwt`, `pkg/ratelimit`) via Mockery;
`make clean-mocks` removes them. (`pkg/security`, `pkg/utils` and `pkg/redis`
ship hand-maintained mocks.)

## Code Quality

```bash
export SONAR_TOKEN=...   # required
make docker-sonar        # SonarCloud scan: coverage + SAST + smells
```

## Make Targets

```bash
make help   # list all targets
```

| Target | Description |
|--------|-------------|
| `make test` | Tests + coverage (80% threshold) |
| `make test-coverage` | Open coverage HTML |
| `make generate-mocks` | Regenerate Mockery mocks (jwt, ratelimit) |
| `make clean-mocks` | Remove generated mocks |
| `make fmt` / `vet` / `lint` | Format / vet / golangci-lint |
| `make tidy` / `vendor` | Tidy / vendor dependencies |
| `make docker-test` | Test in Docker with coverage extraction |
| `make docker-sonar` | SonarCloud scan |
| `make install-tools` | Install mockery + golangci-lint |
| `make info` | Show library/module info |
| `make clean` | Remove coverage artifacts |

## Project Structure

```
bookmark-common/
├── middleware/                 # jwt.go, ratelimit.go (+ _test.go)
├── pkg/
│   ├── base62/                 # base62.go, errors.go
│   ├── common/                 # error.go
│   ├── dbutils/                # error.go
│   ├── jwt/                    # claims.go, generator.go, validator.go, mocks/, provider/
│   ├── logger/                 # client.go, config.go
│   ├── ratelimit/              # limiter.go, store.go, redis.go, mocks/, provider/
│   ├── redis/                  # client.go, config.go, mock.go
│   ├── requestutils/           # request.go
│   ├── response/               # api_response.go, error_response.go
│   ├── security/               # password_hasher.go, mocks/
│   ├── shortcode/              # shortcode.go
│   ├── sqldb/                  # client.go, config.go, migration.go, mock.go
│   └── utils/                  # code_generator.go, mocks/
├── .github/workflows/ci.yaml   # docker-test + docker-sonar
├── Dockerfile                  # multi-stage: base → test-exec → test (coverage export)
├── Makefile
├── go.mod / go.sum
└── coverage_report/            # generated
```

## Security

- RSA (asymmetric) JWT signing/validation; keys loaded by `pkg/jwt/provider`.
- bcrypt password hashing via `pkg/security`.
- Redis-backed rate limiting to curb abuse.
- SonarCloud SAST on every CI run.
