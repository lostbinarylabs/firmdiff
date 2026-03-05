#!/usr/bin/env bash
set -e

VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS="-X github.com/lostbinarylabs/firmdiff/internal/app.Version=$VERSION \
         -X github.com/lostbinarylabs/firmdiff/internal/app.GitCommit=$COMMIT \
         -X github.com/lostbinarylabs/firmdiff/internal/app.BuildDate=$DATE"

go build -ldflags "$LDFLAGS" -o bin/firmdiff ./cmd/firmdiff