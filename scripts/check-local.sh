#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

echo "==> gofmt check"
unformatted="$(find . -name '*.go' -not -path './.git/*' -exec gofmt -l {} +)"
if [ -n "$unformatted" ]; then
  echo "The following Go files need gofmt:"
  echo "$unformatted"
  exit 1
fi

echo "==> go test ./..."
go test ./...

echo "==> go vet ./..."
go vet ./...

echo "==> docker compose config (validation only; no containers are started)"
docker compose config >/dev/null

if [ "${SKIP_ASSET_CHECK:-0}" = "1" ]; then
  echo "==> Asset-Check uebersprungen (SKIP_ASSET_CHECK=1)"
else
  echo "==> asset check"
  sh "$(dirname "$0")/check-assets.sh"
fi

echo "All local checks passed."
