#!/usr/bin/env bash
set -euo pipefail

echo "Docker Compose AWS: Stoppe Container ..."
docker compose down

echo "Docker Compose AWS: Baue Web-Service neu ..."
docker compose build --no-cache web

echo "Docker Compose AWS: Starte Container ..."
docker compose up -d

docker compose ps
