# bookmark-common

A production-ready shared Go library for bookmark microservices. Provides common middleware, utilities, and infrastructure code used across all bookmark services. Extracted from `bookmark-service-monolithic` to enable independent microservices while maintaining consistent patterns and code quality standards.

## Overview

**bookmark-common** is a reusable library that centralizes authentication, rate limiting, database utilities, logging, caching, and security infrastructure. It eliminates code duplication across microservices and ensures consistent implementation of cross-cutting concerns.

## 🎯 Features

- **JWT Authentication Middleware**: RSA-based token validation and user context extraction
- **Rate Limiting Middleware**: Request throttling with Redis backend
- **Structured Logging**: Zerolog-based logging with built-in configuration
- **Database Utilities**: PostgreSQL connection management via GORM with error mappings
- **Redis Client**: Pre-configured Redis connection with health checks
- **JWT Token Management**: Token generation, validation, and claims handling
- **Password Security**: bcrypt-based password hashing utilities
- **Short Code Generation**: Base62 encoding for shortened URLs
- **HTTP Utilities**: Request binding and validation helpers
- **Response Formatting**: Standardized API response and error formatting
- **Code Generation**: Utilities for ID and code generation
- **Health Check Interfaces**: Pinger abstractions for health monitoring
- **Comprehensive Testing**: All business logic tested with 98% coverage
- **Docker Ready**: Multi-stage Dockerfile for testing and coverage extraction
- **Code Quality Gates**: Enforced 80% coverage threshold on business logic
- **SonarCloud Integration**: Continuous code quality and security scanning

## 📋 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26 |
| Web Framework | Gin | v1.12.0 |
| Database | PostgreSQL | (via GORM) |
| ORM | GORM | v1.31.1 (postgres driver v1.6.0) |
| Cache | Redis | v9.19.0 |
| Logger | Zerolog | v1.35.1 |
| Auth | JWT (RSA) | v5.3.1 |
| Password Hashing | bcrypt | (golang.org/x/crypto) |
| Validation | go-playground/validator | v10.30.2 |
| Testing | Testify | v1.11.1 |
| Code Gen | Mockery | (for test mocks) |

## 🚀 Quick Start

### Prerequisites

- **Go 1.26** or higher
- **Git**

### Installation

1. **Include in your service's go.mod:**
   ```bash
   go get github.com/huypham67/bookmark-common
   ```

2. **Or with version pinning:**
   ```bash
   go get github.com/huypham67/bookmark-common@v1.0.0
   ```

3. **Tidy dependencies:**
   ```bash
   go mod tidy
   ```

## 📦 What's Inside

### Middleware

| Middleware | File | Purpose |
|-----------|------|---------|
| **JWT Auth** | `middleware/jwt.go` | Token validation + user context extraction |
| **Rate Limiting** | `middleware/ratelimit.go` | Request throttling with Redis |

### Packages

| Package | Purpose |
|---------|---------|
| **pkg/base62** | Base62 encoding/decoding for short codes |
| **pkg/common** | Common error definitions and types |
| **pkg/dbutils** | Database error mappings and utilities |
| **pkg/jwt** | JWT token generation, validation, claims management |
| **pkg/jwt/provider** | JWT configuration and RSA key loading (DI) |
| **pkg/logger** | Zerolog-based structured logging configuration |
| **pkg/ratelimit** | Rate limiter store and limiter interfaces |
| **pkg/ratelimit/provider** | Rate limit configuration (DI) |
| **pkg/redis** | Redis client initialization and health checks |
| **pkg/requestutils** | HTTP request binding and validation helpers |
| **pkg/response** | API response and error response formatting |
| **pkg/security** | Password hashing (bcrypt) utilities |
| **pkg/shortcode** | Short code generation logic |
| **pkg/sqldb** | PostgreSQL client configuration and migrations |
| **pkg/utils** | Code generators and general utilities |

## 💡 Usage

### JWT Authentication Middleware

```go
import (
  "github.com/huypham67/bookmark-common/middleware"
  "github.com/huypham67/bookmark-common/pkg/jwt"
)

// In your bootstrap/DI setup:
validator := jwtProvider.NewValidator()  // From jwt/provider
router.Use(middleware.JWTAuth(validator))

// In your handler:
// JWT claims are extracted and available in context
claims := c.MustGet("claims").(*jwt.Claims)
userID := claims.UserID
```

### Rate Limiting Middleware

```go
import (
  "github.com/huypham67/bookmark-common/middleware"
  "github.com/huypham67/bookmark-common/pkg/ratelimit"
)

// In your bootstrap/DI setup:
limiter := rateLimitProvider.NewLimiter()  // From ratelimit/provider
router.Use(middleware.RateLimit(limiter))

// Rate limits by client IP, returns 429 if exceeded
```

### Structured Logging

```go
import "github.com/huypham67/bookmark-common/pkg/logger"

// Initialize logger
log := logger.NewClient()

// Use throughout your service
log.Info().Str("user_id", userID).Msg("User logged in")
log.Error().Err(err).Msg("Database query failed")
```

### Password Hashing

```go
import "github.com/huypham67/bookmark-common/pkg/security"

hasher := security.NewBcryptPasswordHasher()

// Hash password
hash, err := hasher.HashPassword("password123")

// Verify password
valid, err := hasher.VerifyPassword(hash, "password123")
```

### Short Code Generation

```go
import "github.com/huypham67/bookmark-common/pkg/shortcode"

generator := shortcode.NewShortCodeGenerator()
code, err := generator.Generate()  // Returns 8-char Base62 code
```

### JWT Token Management

```go
import "github.com/huypham67/bookmark-common/pkg/jwt"

// In your JWT provider setup:
generator := jwt.NewGenerator(privateKey, issuer, audience, expirationSeconds)
validator := jwt.NewValidator(publicKey, issuer, audience)

// Generate token
token, err := generator.GenerateToken(userID, email)

// Validate token
claims, err := validator.ValidateToken(tokenString)
```

### Database Utilities

```go
import (
  "github.com/huypham67/bookmark-common/pkg/sqldb"
  "github.com/huypham67/bookmark-common/pkg/dbutils"
)

// Initialize database
db, err := sqldb.NewClient(config)

// Handle database errors consistently
if errors.Is(err, dbutils.ErrRecordNotFound) {
  // Handle 404
}
```

### Redis Client

```go
import "github.com/huypham67/bookmark-common/pkg/redis"

// Initialize Redis
redisClient := redis.NewClient(config)

// Use for caching or rate limiting
redisClient.Set(ctx, "key", "value", expiration)
```

## 🏗️ Architecture

### Clean Architecture Pattern

This library follows the same clean architecture principles as the monolithic service:

```
Handler/Service Layer (your service)
    ↓
Middleware (jwt.go, ratelimit.go)
    ↓
Package Utilities (logging, validation, response formatting)
    ↓
Infrastructure (database, redis, security)
```

### Design Principles

- **Single Responsibility**: Each package has one clear purpose
- **Interface-Based**: All major components use interfaces for testability
- **Mockable Infrastructure**: All external dependencies can be mocked
- **No Business Logic**: Pure infrastructure and utilities only
- **Consistent Error Handling**: Common error types and mappings
- **Testable Code**: Comprehensive test coverage (98%)

### Dependency Injection

Each service using this library is responsible for creating instances via constructor injection. See `pkg/jwt/provider` and `pkg/ratelimit/provider` for DI wiring patterns.

## 🧪 Testing

### Run Tests Locally

```bash
# Run all tests with coverage (80% threshold enforced)
make test

# View coverage in HTML
make test-coverage

# Run Go vet
make vet

# Run linter
make lint

# Format code
make fmt
```

### Docker-Based Testing

```bash
# Test in Docker (same as CI pipeline)
make docker-test

# Extract coverage to coverage_report/
```

### Test Coverage

- **98% coverage** on tested code (middleware + packages with tests)
- **80% threshold** enforced via Makefile
- **Coverage exclusions**: Infrastructure code (configs, DI wiring, database clients)
- **Test types**: Unit tests for all logic, mocks for external dependencies

**Tested components:**
- `middleware/jwt.go` - 100% coverage
- `middleware/ratelimit.go` - 100% coverage
- `pkg/base62/` - 100% coverage
- `pkg/jwt/` (claims, generator, validator) - 94.3% coverage
- `pkg/ratelimit/` (limiter, redis) - 94.1% coverage
- `pkg/shortcode/` - 100% coverage

**Excluded from threshold** (infrastructure):
- `pkg/common`, `pkg/dbutils` - error definitions
- `pkg/logger`, `pkg/redis`, `pkg/sqldb` - connection setup
- `pkg/jwt/provider`, `pkg/ratelimit/provider` - DI configuration
- `pkg/requestutils`, `pkg/response` - formatting utilities
- `pkg/security`, `pkg/utils` - infrastructure helpers

### Test Files

```bash
# Unit tests are alongside source code
pkg/jwt/claims_test.go
pkg/jwt/generator_test.go
pkg/jwt/validator_test.go
middleware/jwt_test.go
middleware/ratelimit_test.go
# ... etc
```

## 📊 Code Quality

### SonarCloud Integration

```bash
# Run SonarCloud code quality scan (requires SONAR_TOKEN)
export SONAR_TOKEN=your_token_here
make docker-sonar
```

**Quality checks include:**
- Code coverage (80% on business logic)
- Security vulnerabilities (SAST scan)
- Code smells and maintainability
- Duplicate code detection
- Complexity analysis

### Code Formatting

```bash
make fmt     # Format all Go files
make vet     # Run go vet analysis
make lint    # Run golangci-lint
```

## 🐳 Docker & CI/CD

### Docker Multi-Stage Build

```dockerfile
Stage 1 (BASE):     Install dependencies, download modules
Stage 2 (TEST-EXEC): Run tests, generate coverage reports
Stage 3 (TEST):     Extract coverage artifacts for CI
```

### CI Pipeline (GitHub Actions)

```yaml
Workflow: CI
├─ Job: ci
   ├─ Step 1: Checkout code
   ├─ Step 2: Setup Docker Buildx
   ├─ Step 3: make docker-test (extract coverage)
   ├─ Step 4: Upload coverage to Codecov
   └─ Step 5: make docker-sonar (SonarCloud scan)
```

**Environment Variables:**
```yaml
CACHE_FROM: type=gha          # Docker layer cache input (GitHub Actions)
CACHE_TO: type=gha,mode=max   # Docker layer cache output
COVERAGE_THRESHOLD: 80        # Minimum coverage percentage
```

**Secrets:**
- `SONAR_TOKEN` - SonarCloud authentication token

## 🛠️ Available Make Targets

```bash
make help              # Display all available targets
```

| Target | Description |
|--------|-------------|
| `make test` | Run tests with coverage (80% threshold) |
| `make test-coverage` | View coverage HTML report |
| `make docker-test` | Test in Docker with coverage extraction |
| `make docker-sonar` | SonarCloud code quality scan |
| `make fmt` | Format code with go fmt |
| `make vet` | Run go vet analysis |
| `make lint` | Run golangci-lint |
| `make tidy` | Tidy dependencies |
| `make clean` | Remove build artifacts and coverage |

## 🔄 Development Workflow

### Local Development

```bash
# 1. Make changes to code
# 2. Format and check
make fmt
make vet

# 3. Run tests locally
make test

# 4. Test in Docker (as CI would)
make docker-test

# 5. Code quality check
make docker-sonar

# 6. Commit when ready
```

### Before Committing

- [ ] Code formatted: `make fmt`
- [ ] No linting issues: `make vet`
- [ ] Tests pass and coverage ≥ 80%: `make test`
- [ ] Docker build succeeds: `make docker-test`
- [ ] No debug prints or commented code

## 📄 Project Structure

```
bookmark-common/
├── middleware/
│   ├── jwt.go                  # JWT authentication middleware
│   ├── jwt_test.go            # Middleware tests
│   ├── ratelimit.go            # Rate limiting middleware
│   └── ratelimit_test.go       # Middleware tests
├── pkg/
│   ├── base62/
│   │   ├── base62.go           # Base62 encoding/decoding
│   │   └── base62_test.go
│   ├── common/
│   │   └── error.go            # Common error types
│   ├── dbutils/
│   │   └── error.go            # Database error mappings
│   ├── jwt/
│   │   ├── claims.go           # JWT claims structure
│   │   ├── claims_test.go
│   │   ├── generator.go        # Token generation logic
│   │   ├── generator_test.go
│   │   ├── validator.go        # Token validation logic
│   │   ├── validator_test.go
│   │   ├── mocks/              # Test mocks
│   │   └── provider/           # DI wiring
│   ├── logger/
│   │   ├── client.go           # Zerolog configuration
│   │   └── config.go
│   ├── ratelimit/
│   │   ├── limiter.go          # Rate limiter logic
│   │   ├── limiter_test.go
│   │   ├── redis.go            # Redis store
│   │   ├── redis_test.go
│   │   ├── store.go            # Store interface
│   │   ├── mocks/              # Test mocks
│   │   └── provider/           # DI wiring
│   ├── redis/
│   │   ├── client.go           # Redis client
│   │   ├── config.go
│   │   └── mock.go
│   ├── requestutils/
│   │   └── request.go          # Request binding helpers
│   ├── response/
│   │   ├── api_response.go     # Response formatting
│   │   └── error_response.go
│   ├── security/
│   │   ├── password_hasher.go  # bcrypt hashing
│   │   └── mocks/              # Test mocks
│   ├── shortcode/
│   │   ├── shortcode.go        # Short code generation
│   │   └── shortcode_test.go
│   ├── sqldb/
│   │   ├── client.go           # PostgreSQL connection
│   │   └── migration.go
│   └── utils/
│       ├── code_generator.go   # Utility code generation
│       └── mocks/              # Test mocks
├── .github/workflows/
│   └── ci.yaml                 # GitHub Actions CI pipeline
├── Dockerfile                  # Multi-stage Docker build
├── Makefile                    # Build and test automation
├── go.mod / go.sum             # Go module definition
├── .gitignore
├── README.md                   # This file
└── coverage_report/            # Generated coverage reports
```

## 🔐 Security Considerations

- **JWT Tokens**: RSA-based, asymmetric cryptography
- **Password Security**: bcrypt with appropriate salt rounds
- **Input Validation**: Struct-tag based validation
- **Error Handling**: Consistent error types for security
- **Rate Limiting**: Protects against abuse and DoS
- **Dependency Pinning**: All external actions pinned to commit SHAs in CI
- **Code Scanning**: SonarCloud SAST scanning on all code

## 🔗 Integration with Services

This library is designed to be imported by:
- `user-service` - JWT auth, password hashing, logging
- `bookmark-service` - JWT auth, rate limiting, caching, logging
- Any other bookmark microservice

See the individual service READMEs for integration examples.

## 📚 Monolithic Origin

This library was extracted from `bookmark-service-monolithic`. It maintains exact code structure and patterns to ensure consistency across the microservices architecture.

**Key design decisions inherited from mono:**
- Clean architecture with layered separation
- Repository pattern for data access
- Service pattern for business logic
- Middleware pattern for cross-cutting concerns
- Constructor-based dependency injection
- Comprehensive test coverage with gating

## 📞 Support

For issues, questions, or suggestions about this library, please create an issue in the bookmark services repository.

## 🔗 Useful Links

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Introduction](https://jwt.io/introduction)
- [Redis Documentation](https://redis.io/docs/)
- [Zerolog Logger](https://github.com/rs/zerolog)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
