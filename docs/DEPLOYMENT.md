# Deployment

## Zielumgebung

Das Projekt ist fuer eine AWS-Instanz mit Docker Compose vorbereitet.

Services:

- `labbi-web`: Go-App
- `neo4jLabbi`: Datenbank
- `labbi-nginx`: Reverse Proxy, TLS, Static/Uploads

Persistenz:

- Docker Volume `neo4j_data`
- Docker Volume `uploads`
- Zertifikate unter `/etc/letsencrypt`

## VSCode SSH Workflow

1. Per VSCode Remote SSH auf die AWS-Instanz verbinden.
2. Terminal im Projektroot oeffnen.
3. Status pruefen.

```bash
cd /pfad/zur/labbi-app
git status
git pull
```

4. Compose-Konfiguration pruefen.

```bash
docker compose config
```

Vorsicht: Die Ausgabe kann Secrets aus `.env` enthalten. Nicht in Tickets oder Chat kopieren.

5. Build und Restart.

```bash
docker compose build
docker compose up -d
docker compose ps
```

6. Logs beobachten.

```bash
docker compose logs -f web nginx neo4j
```

## Dateien auf der Instanz

Muss vorhanden sein:

- `docker-compose.yml`
- `Dockerfile`
- `nginx.conf`
- `static/`
- `internal/public/`
- `.env` mit echten Werten
- `/etc/letsencrypt/...` Zertifikate

Darf nicht committed werden:

- `.env`
- `.env.*`
- echte Passwoerter oder Tokens

## Zertifikate und Nginx

In `nginx.conf` sind vier HTTPS-Servernamen konfiguriert:

- `labbi-welpen.de`
- `www.labbi-welpen.de`
- `labbi-hobby.de`
- `www.labbi-hobby.de`

Jeder HTTPS-Serverblock braucht ein passendes Zertifikat oder ein gemeinsames Multi-Domain-Zertifikat. ACME-Challenges werden aus `./internal/public` nach `/var/www/html` gemountet.

Pruefen:

```bash
sudo ls -la /etc/letsencrypt/live
docker compose logs nginx
curl -I https://labbi-welpen.de/healthz
```

## Rollback-Grundidee

Ein einfacher Rollback ist Git-basiert:

```bash
git log --oneline -5
git checkout <alter-commit>
docker compose build
docker compose up -d
```

Danach wieder zur gewuenschten Branch zurueck:

```bash
git checkout main
```

Vor einem produktiven Rollback immer daran denken: Datenbank- und Upload-Zustand bleiben in Volumes erhalten.

## Deployment Smoke Test

Nach jedem Deployment:

```bash
docker compose ps
curl -i http://localhost/healthz
curl -I https://labbi-welpen.de/
curl -I https://labbi-welpen.de/static/css/styles.css
docker compose logs --tail=100 web
docker compose logs --tail=100 nginx
docker compose logs --tail=100 neo4j
```

Admin manuell:

1. `/admin` oeffnen.
2. Basic Auth testen.
3. `/admin/puppies` laden.
4. Kein Test-Delete in Produktion ohne Backup/Absprache.
