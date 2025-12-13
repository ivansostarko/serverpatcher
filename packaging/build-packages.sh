#!/usr/bin/env bash
set -euo pipefail

# Requires:
#   go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
#   (and `nfpm` in PATH)

VERSION="${VERSION:-0.1.0}"
OUTDIR="${OUTDIR:-dist}"

rm -rf "${OUTDIR}"
mkdir -p "${OUTDIR}"

echo "Building linux/amd64 binary..."
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X github.com/serverpatcher/serverpatcher/internal/version.Version=${VERSION}" -o "${OUTDIR}/serverpatcher" ./cmd/serverpatcher

echo "Packaging DEB..."
nfpm package --packager deb --config packaging/nfpm-deb.yaml --target "${OUTDIR}" --version "${VERSION}" --arch amd64

echo "Packaging RPM..."
nfpm package --packager rpm --config packaging/nfpm-rpm.yaml --target "${OUTDIR}" --version "${VERSION}" --arch x86_64

echo "Artifacts:"
ls -lah "${OUTDIR}"
