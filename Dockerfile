# Stage 1: BASE - Install Dependencies and Prepare Source
FROM golang:1.26-alpine AS base

RUN apk add --no-cache build-base git

WORKDIR /opt/app

# Separate copy go.mod/sum for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Stage 2: TEST-EXEC - Run Tests and Generate Coverage Reports
FROM base AS test-exec

ARG COVERAGE_EXCLUDE
ARG COVERPKG=./...
ENV _OUTPUTDIR=/tmp/coverage

RUN mkdir -p ${_OUTPUTDIR}

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go test ./... \
      -coverprofile=coverage.tmp \
      -covermode=atomic \
      -coverpkg=${COVERPKG} \
      -p 1 && \
    grep -v -E "${COVERAGE_EXCLUDE}" coverage.tmp > ${_OUTPUTDIR}/coverage.out && \
    go tool cover -html=${_OUTPUTDIR}/coverage.out -o ${_OUTPUTDIR}/coverage.html

# Stage 3: TEST-REPORT - Extract Coverage Reports for CI/CD
FROM scratch AS test

COPY --from=test-exec /tmp/coverage/coverage.out /coverage.out
COPY --from=test-exec /tmp/coverage/coverage.html /coverage.html
