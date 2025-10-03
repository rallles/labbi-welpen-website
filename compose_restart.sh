#!/bin/bash

echo "Docker Compose: Stoppe und entferne laufende Container ..."
docker compose down

echo
echo "Docker Compose: Baue 'web' Service komplett neu (ohne Cache) ..."
docker compose build --no-cache web

echo
echo "Docker Compose: Starte Container im Hintergrund (detached Mode) ..."
docker compose up -d

echo
echo "Fertig! Alle Container laufen jetzt neu gebaut im Hintergrund."
