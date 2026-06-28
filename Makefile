# Makefile for Bookmark Common (Shared Library)

# =============================================================================
# LIBRARY METADATA
# =============================================================================

APP_NAME    := bookmark-common
MAIN_PKG    := github.com/huypham67/bookmark-common

# =============================================================================
# COVERAGE & QUALITY GATES
# =============================================================================

COVERAGE_DIR       ?= coverage_report
COVERAGE_THRESHOLD ?= 80

# ═══════════════════════════════════════════════════════════════════════════
# SINGLE SOURCE OF TRUTH: Coverage & Quality Gate Exclusions
#
# Strategy:
#   1. SYSTEM_DIRS/FILES: Completely excluded (no scan, no coverage)
#      → Auto-generated, vendored, test infrastructure
#      → Used for sonar.exclusions + local coverage filter
#
#   2. INFRA_DIRS / INFRA_FILES: Exclude from coverage % but INCLUDE in scan
#      → Infrastructure/setup code (DI, config, adapters)
#      → INFRA_DIRS: whole packages excluded from coverage threshold.
#        Pure adapters/wiring with no testable logic, e.g. pkg/logger,
#        pkg/redis, pkg/sqldb, pkg/jwt/provider, pkg/ratelimit/provider
#        (env load, key file I/O, DI), pkg/tracing (NR app init + NR
#        API adapters: Extract/Continue).
#      → INFRA_FILES: surgical per-file exclusion for packages that mix
#        tested logic with wiring/setup. Currently empty — packages are
#        split so each is wholly one category (see pkg/jwt vs pkg/jwt/provider).
#      → Both are still scanned for security vulnerabilities (SonarQube)
#
#   3. Everything else = business logic → MUST be covered:
#      pkg/base62, pkg/shortcode, pkg/jwt (generator, validator, claims),
#      pkg/ratelimit, middleware
#
# Usage:
#   - make test        → filters coverage.out to exclude infrastructure + system
#   - make docker-test → passes COVERAGE_EXCLUDE to Docker build
#   - make docker-sonar → sonar.exclusions (system only)
#                        sonar.coverage.exclusions (system + infra)
# ═══════════════════════════════════════════════════════════════════════════

# Infrastructure dirs: exclude from coverage % but SCAN for security
INFRA_DIRS := \
	pkg/common \
	pkg/dbutils \
	pkg/jwt/provider \
	pkg/ratelimit/provider \
	pkg/logger \
	pkg/redis \
	pkg/requestutils \
	pkg/response \
	pkg/security \
	pkg/sqldb \
	pkg/tracing \
	pkg/utils

# Infrastructure files: surgical per-file coverage exclusion (still SCANNED).
# For packages that mix tested logic with wiring/setup.
INFRA_FILES :=

# System artifacts: auto-generated, vendored, test infrastructure (NO SCAN)
SYSTEM_DIRS := vendor docs bin mocks
SYSTEM_FILES := _test.go .pb.go test_helper.go mock.go

# Format conversion for Makefile
comma := ,
space := $(subst ,, )

# Pattern builders for Sonar (Ant-style glob)
SONAR_INFRA_DIRS := $(foreach d,$(INFRA_DIRS),**/$(d)**)
SONAR_INFRA_FILES := $(foreach f,$(INFRA_FILES),**/$(f))
SONAR_SYSTEM_DIRS := $(foreach d,$(SYSTEM_DIRS),**/$(d)**)
SONAR_SYSTEM_FILES := $(foreach f,$(SYSTEM_FILES),**/*$(f))

# Sonar: exclude system artifacts completely (INFRA_FILES intentionally absent → still scanned)
SONAR_EXCLUDE_PATTERNS := $(subst $(space),$(comma),$(strip $(SONAR_SYSTEM_FILES) $(SONAR_SYSTEM_DIRS) $(COVERAGE_DIR)/**))

# Sonar: exclude infrastructure (dirs + files) from coverage % but allow security scan
SONAR_COVERAGE_EXCLUSIONS := $(subst $(space),$(comma),$(strip $(SONAR_INFRA_DIRS) $(SONAR_INFRA_FILES) $(SONAR_SYSTEM_DIRS)))

# Local/Docker: Regex format (coverage.out filtering)
ALL_EXCLUDES := $(INFRA_DIRS) $(INFRA_FILES) $(SYSTEM_DIRS) $(SYSTEM_FILES)
COVERAGE_EXCLUDE := $(subst $(space),|,$(strip $(ALL_EXCLUDES)))

# Go test: Scan all, let grep filter
COVERPKG := ./...

# =============================================================================
# GO COMPILER
# =============================================================================

GO      := go
GOLINT  := golangci-lint

# =============================================================================
# DOCKER
# =============================================================================

DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= huypham053
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)

CACHE_FROM ?= type=local,src=/tmp/.buildx-cache
CACHE_TO ?= type=local,dest=/tmp/.buildx-cache-new,mode=max

# =============================================================================
# MACROS
# =============================================================================

.DEFAULT_GOAL := help

# =============================================================================
# DEVELOPMENT
# =============================================================================

.PHONY: help fmt vet lint tidy vendor

help:
	@echo "Development:"
	@echo "  make fmt             Format code"
	@echo "  make vet             Static analysis"
	@echo "  make lint            Linter"
	@echo "  make tidy            Dependencies"
	@echo "  make vendor          Vendor dependencies"
	@echo ""
	@echo "Testing:"
	@echo "  make test            Local tests + coverage report"
	@echo "  make test-coverage   Open coverage HTML"
	@echo ""
	@echo "Mocks:"
	@echo "  make generate-mocks  Generate all mocks"
	@echo "  make clean-mocks     Clean all mocks"
	@echo ""
	@echo "Docker / CI:"
	@echo "  make docker-test     Test in container"
	@echo "  make docker-sonar    SonarCloud scan"
	@echo ""
	@echo "Utilities:"
	@echo "  make install-tools   Install dev tooling"
	@echo "  make info            Show library info"
	@echo "  make clean           Remove artifacts"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	@which $(GOLINT) > /dev/null || (echo "Error: golangci-lint not found. Run: make install-tools"; exit 1)
	$(GOLINT) run ./...

tidy:
	$(GO) mod tidy

vendor:
	$(GO) mod download
	$(GO) mod vendor

# =============================================================================
# TESTING
# =============================================================================

.PHONY: test test-coverage

test:
	@$(GO) clean -testcache
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test ./... -coverprofile=$(COVERAGE_DIR)/coverage.tmp -covermode=atomic -coverpkg=$(COVERPKG) -p 1
	@head -1 $(COVERAGE_DIR)/coverage.tmp > $(COVERAGE_DIR)/coverage.out
	@grep -vE "$(COVERAGE_EXCLUDE)" $(COVERAGE_DIR)/coverage.tmp | tail -n +2 >> $(COVERAGE_DIR)/coverage.out || true
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@total=$$($(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "FAIL: Below $(COVERAGE_THRESHOLD)% threshold"; exit 1; \
	fi

test-coverage: test
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out

# =============================================================================
# CI / CD
# =============================================================================

.PHONY: docker-test docker-sonar

docker-test:
	@mkdir -p $(COVERAGE_DIR)
	docker buildx build \
		--build-arg COVERAGE_EXCLUDE="$(COVERAGE_EXCLUDE)" \
		--build-arg COVERPKG="$(COVERPKG)" \
		--target test \
		--cache-from=$(CACHE_FROM) \
		--cache-to=$(CACHE_TO) \
		--output type=local,dest=$(COVERAGE_DIR) .
	@if [ -f $(COVERAGE_DIR)/coverage.out ]; then \
		total=$$($(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Coverage: $$total%"; \
		if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "FAIL: Below $(COVERAGE_THRESHOLD)% threshold"; exit 1; \
		fi; \
	else \
		echo "FAIL: coverage.out not found"; exit 1; \
	fi

docker-sonar:
	@[ -n "$(SONAR_TOKEN)" ] || (echo "Error: SONAR_TOKEN not set"; exit 1)
	@docker pull --quiet sonarsource/sonar-scanner-cli:11 || true
	docker run --rm \
		-e SONAR_TOKEN=$(SONAR_TOKEN) \
		-e SONAR_HOST_URL=https://sonarcloud.io \
		-v "$(PWD):/usr/src" \
		sonarsource/sonar-scanner-cli:11 \
		-Dsonar.organization="huypham67" \
		-Dsonar.projectKey="huypham67_bookmark-common" \
		-Dsonar.projectName="$(APP_NAME)" \
		-Dsonar.projectVersion="1.0" \
		-Dsonar.sources="." \
		-Dsonar.tests="." \
		-Dsonar.test.inclusions="**/*_test.go" \
		-Dsonar.test.exclusions="**/vendor/**,**/mocks/**" \
		-Dsonar.exclusions="$(SONAR_EXCLUDE_PATTERNS)" \
		-Dsonar.coverage.exclusions="$(SONAR_COVERAGE_EXCLUSIONS)" \
		-Dsonar.go.coverage.reportPaths="$(COVERAGE_DIR)/coverage.out" \
		-Dsonar.qualitygate.wait=true

# =============================================================================
# MOCKS
# =============================================================================

.PHONY: generate-mocks clean-mocks

generate-mocks:
	@echo "Generating mocks for jwt..."
	cd pkg/jwt && $(GO) generate
	@echo "Generating mocks for ratelimit..."
	cd pkg/ratelimit && $(GO) generate
	@echo "✓ Mocks generated successfully"

clean-mocks:
	@echo "Cleaning mocks..."
	rm -rf pkg/jwt/mocks
	rm -rf pkg/ratelimit/mocks
	@echo "✓ Mocks cleaned"

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: install-tools info clean

install-tools:
	$(GO) install github.com/vektra/mockery/v2@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

info:
	@echo "Library:   $(APP_NAME)"
	@echo "Module:    $(MAIN_PKG)"
	@echo "Go:        $$($(GO) version)"

clean:
	rm -rf $(COVERAGE_DIR)
