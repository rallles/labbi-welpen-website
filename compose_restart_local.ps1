Write-Host "Docker Compose lokal: Stoppe Container ..."
docker compose down

Write-Host "Docker Compose lokal: Baue Web-Service neu ..."
docker compose build --no-cache web

Write-Host "Docker Compose lokal: Starte Container ..."
docker compose up -d

docker compose ps
