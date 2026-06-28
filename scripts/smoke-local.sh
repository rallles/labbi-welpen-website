#!/usr/bin/env sh
set -eu

base_url="${BASE_URL:-http://localhost:8081}"

check_get() {
  path="$1"
  echo "==> GET ${base_url}${path}"
  curl -fsS "${base_url}${path}" >/dev/null
}

check_get "/healthz"
check_get "/"
check_get "/puppies"
check_get "/robots.txt"
check_get "/sitemap.xml"

echo "==> HEAD ${base_url}/healthz"
curl -fsS -I "${base_url}/healthz" >/dev/null

echo "Local smoke checks passed."
