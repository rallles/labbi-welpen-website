#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

echo "==> gofmt check"
unformatted="$(gofmt -l $(find . -name '*.go' -not -path './.git/*'))"
if [ -n "$unformatted" ]; then
  echo "The following Go files need gofmt:"
  echo "$unformatted"
  exit 1
fi

echo "==> go test ./..."
go test ./...

echo "==> go vet ./..."
go vet ./...

echo "==> docker compose config"
docker compose config >/dev/null

echo "All local checks passed."
