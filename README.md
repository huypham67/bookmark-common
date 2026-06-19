# bookmark-common

Shared Go library for the bookmark microservices. It centralises cross-cutting
concerns — JWT auth, rate limiting, logging, database/Redis clients, password
hashing, request/response helpers, CSV decoding, and code utilities — so each
service implements them once, consistently.

This module ships **no `main`, no binary, and no migrations of its own** — it is a
library consumed via `go get`.

## Tech Stack

| Component | Technology | Version |
|---|---|---|
| Language | Go | 1.26 |
| ORM / DB | GORM + pgx driver | v1.31.1 / v1.6.0 |
| Cache | go-redis | v9.19.0 |
| Logger | Zerolog | v1.35.1 |
| Auth | JWT (RS256, golang-jwt) | v5.3.1 |
| Migrations | golang-migrate | v4.19.1 |
| Config | envconfig | v1.4.0 |
| Validation | go-playground/validator | v10.30.2 |
| Password hashing | bcrypt (golang.org/x/crypto) | v0.52.0 |
| CSV parsing | gocarina/gocsv | latest |
| Testing | Testify + miniredis | v1.11.1 / v2.38.0 |
| Mocks | Mockery | v2 |

## Installation

```bash
go get github.com/huypham67/bookmark-common@v0.3.0
go mod tidy
```

## What's Inside

### Middleware

| Middleware | File | Purpose |
|---|---|---|
| `JWTAuth(validator)` | `middleware/jwt.go` | Validates the Bearer token; stores `*jwt.CustomClaims` in Gin context under `"claims"` |
| `RateLimit(limiter)` | `middleware/ratelimit.go` | Per-client Redis sliding-window throttle; returns `429` when exceeded |

### Packages

| Package | Purpose |
|---|---|
| `pkg/base62` | Base62 encode/decode for short codes |
| `pkg/common` | `ExitOnError` helper |
| `pkg/csvutil` | Generic CSV decoder with per-row validation (`csvutil.Decode[T]`); built on gocsv + validator |
| `pkg/dbutils` | GORM/SQL error classifier (duplicate, not-found, etc.) |
| `pkg/jwt` | RS256 token generator, validator, claims + `GetUserIDFromContext` |
| `pkg/jwt/provider` | Env-config constructors: `NewIssuer` (private key → signs) and `NewValidator` (public key only → verifies) |
| `pkg/logger` | Global zerolog configuration |
| `pkg/ratelimit` | `Limiter`/`Store` interfaces + Redis sliding-window implementation |
| `pkg/ratelimit/provider` | Builds a `Limiter` from a Redis client + env config |
| `pkg/redis` | go-redis client from env config |
| `pkg/requestutils` | `FormFile` (multipart upload with size/type enforcement) + generic `Bind[T]` (URI + JSON + query + header → struct + validate) |
| `pkg/response` | Standardised API success/error response formatting |
| `pkg/security` | bcrypt `PasswordHasher` |
| `pkg/shortcode` | Prefix-based shortcode routing: `Classify`, `EncodeSQLCode`, `AddRedisPrefix` |
| `pkg/sliceutil` | `SplitIntoBatches[T]` — splits a slice into fixed-size chunks |
| `pkg/sqldb` | GORM PostgreSQL client + golang-migrate runner (`MigratePostgresDB`, `RunMigration`, `ForcePostgresDB`, `PostgresMigrationVersion`) |
| `pkg/utils` | `CodeGenerator` interface + random implementation |

All major components are defined as interfaces and ship test mocks (under `mocks/`) so consuming services can mock them without real dependencies.

## Usage

### JWT auth (provider + middleware)

```go
import (
    "github.com/huypham67/bookmark-common/middleware"
    jwtprovider "github.com/huypham67/bookmark-common/pkg/jwt/provider"
    "github.com/huypham67/bookmark-common/pkg/jwt"
)

// Token issuer (user-service) — signs with private key
generator, err := jwtprovider.NewIssuer("")
token, err := generator.GenerateToken(userID, displayName, email)

// Relying party (bookmark-service) — loads only public key, never signs
validator, err := jwtprovider.NewValidator("")
router.Use(middleware.JWTAuth(validator))

// Inside a handler
userID, err := jwt.GetUserIDFromContext(c)
```

### Rate limiting (provider + middleware)

```go
import (
    "github.com/huypham67/bookmark-common/middleware"
    ratelimitprovider "github.com/huypham67/bookmark-common/pkg/ratelimit/provider"
)

redisClient, _ := redis.NewClient("")
limiter, _ := ratelimitprovider.New(redisClient, "")
router.Use(middleware.RateLimit(limiter))
```

### Logger / DB / Redis

```go
logger.NewClient("")
db, err := sqldb.NewClient("")       // *gorm.DB
rdb, err := redis.NewClient("")      // *redis.Client
```

### CSV decoding

```go
import "github.com/huypham67/bookmark-common/pkg/csvutil"

type Row struct {
    URL         string `csv:"url"  validate:"required,url"`
    Description string `csv:"description"`
}

rows, err := csvutil.Decode[Row](file)
// Returns ErrEmptyCSV or ErrInvalidFormat (with row number) on failure.
```

### Password hashing

```go
hasher := security.NewBcryptPasswordHasher()
hash, err := hasher.Hash("secret")
err = hasher.Compare(hash, "secret")
```

## Testing

```bash
make test            # unit tests + coverage, enforces 80% threshold on business logic
make test-coverage   # open the HTML coverage report
make docker-test     # same run inside Docker (CI parity)
```

- **80% threshold** enforced on business logic.
- **Covered** (business logic): `middleware`, `pkg/base62`, `pkg/jwt`, `pkg/ratelimit`, `pkg/shortcode`, `pkg/csvutil`, `pkg/sliceutil`.
- **Excluded from threshold** (infrastructure/wiring, still security-scanned by SonarCloud): `pkg/common`, `pkg/dbutils`, `pkg/logger`, `pkg/redis`, `pkg/sqldb`, `pkg/requestutils`, `pkg/response`, `pkg/security`, `pkg/utils`, `pkg/jwt/provider`, `pkg/ratelimit/provider`.

The exclusion lists are the single source of truth in the `Makefile` (`INFRA_DIRS` / `SYSTEM_DIRS`) and drive both local coverage filtering and SonarCloud `coverage.exclusions`.

### Mocks

```bash
make generate-mocks   # regenerate via Mockery (pkg/jwt, pkg/ratelimit)
make clean-mocks      # delete generated mocks
```

`pkg/security`, `pkg/utils`, and `pkg/redis` ship hand-maintained mocks.

## Code Quality

```bash
export SONAR_TOKEN=...
make docker-sonar    # SonarCloud scan: coverage + SAST
```

## Make Targets

```
make test            Tests + coverage (80% threshold)
make test-coverage   Open coverage HTML
make generate-mocks  Regenerate Mockery mocks
make clean-mocks     Remove generated mocks
make fmt / vet / lint
make tidy / vendor
make docker-test     Test in Docker with coverage extraction
make docker-sonar    SonarCloud scan
make install-tools   Install mockery + golangci-lint
make info            Show module info
make clean           Remove coverage artifacts
```

## Project Structure

```
bookmark-common/
├── middleware/           # jwt.go, ratelimit.go
├── pkg/
│   ├── base62/           # base62.go, errors.go
│   ├── common/           # error.go
│   ├── csvutil/          # reader.go
│   ├── dbutils/          # error.go
│   ├── jwt/              # claims.go, generator.go, validator.go, mocks/, provider/
│   ├── logger/           # client.go, config.go
│   ├── ratelimit/        # limiter.go, store.go, redis.go, mocks/, provider/
│   ├── redis/            # client.go, config.go, mock.go
│   ├── requestutils/     # file.go (upload), request.go (Bind)
│   ├── response/         # api_response.go, error_response.go
│   ├── security/         # password_hasher.go, mocks/
│   ├── shortcode/        # shortcode.go
│   ├── sliceutil/        # batch.go
│   ├── sqldb/            # client.go, config.go, migration.go, mock.go
│   └── utils/            # code_generator.go, mocks/
├── .github/workflows/ci.yaml
├── Dockerfile
├── Makefile
└── go.mod / go.sum
```
