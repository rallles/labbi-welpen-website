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
test -f .env
grep '^COMPOSE_FILE=' .env || true
export COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml
docker compose config
```

Windows trennt mehrere Compose-Dateien mit Semikolon, Linux und AWS mit Doppelpunkt.
Auf AWS darf daher keine lokale Windows-Angabe wie
`docker-compose.yml;docker-compose.local.yml` aktiv sein. Falls erforderlich, in `.env`
`COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml` setzen oder den Wert fuer den
Deployment-Lauf exportieren. Die uebrigen `.env`-Werte nicht ausgeben oder dokumentieren.

Vorsicht: Die Ausgabe kann Secrets aus `.env` enthalten. Nicht in Tickets oder Chat kopieren.

5. Build und Restart.

```bash
docker compose up -d --build
docker compose ps
```

6. Logs beobachten.

```bash
docker compose logs -f web nginx neo4j
```

## Dateien auf der Instanz

Muss vorhanden sein:

- `docker-compose.yml`
- `docker-compose.aws.yml`
- `Dockerfile`
- `nginx.aws.conf`
- `static/`
- `internal/public/`
- `.env` mit echten Werten
- `/etc/letsencrypt/live/labbi-sites/fullchain.pem`
- `/etc/letsencrypt/live/labbi-sites/privkey.pem`

Darf nicht committed werden:

- `.env`
- `.env.*`
- echte Passwoerter oder Tokens

SMTP ist fuer den Betrieb optional. Ohne SMTP-Werte speichert das Kontaktformular
Anfragen ausschliesslich in Neo4j. Eine echte Benachrichtigung wird nur versucht, wenn
`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` und `CONTACT_MAIL_TO` vollstaendig
gesetzt sind. Teilkonfigurationen gelten bewusst als deaktiviert; keine Dummy-Werte einsetzen.

## Zertifikate und Nginx

Auf AWS mounted `docker-compose.aws.yml` die Datei `nginx.aws.conf` nach `/etc/nginx/nginx.conf`.
Die Root-Datei `nginx.conf` ist nicht produktiv genutzt.

In `nginx.aws.conf` sind vier HTTPS-Servernamen konfiguriert:

- `labbi-welpen.de`
- `www.labbi-welpen.de`
- `labbi-hobby.de`
- `www.labbi-hobby.de`

Alle HTTPS-Serverbloecke verwenden das gemeinsame Let's-Encrypt-Zertifikat `labbi-sites`:

- `/etc/letsencrypt/live/labbi-sites/fullchain.pem`
- `/etc/letsencrypt/live/labbi-sites/privkey.pem`

`labbi-welpen.de` bedient die App. `www.labbi-welpen.de`, `labbi-hobby.de` und `www.labbi-hobby.de` redirecten auf `https://labbi-welpen.de$request_uri`.
ACME-Challenges werden aus `./internal/public` nach `/var/www/html` gemountet und bleiben auf HTTP unter `/.well-known/acme-challenge/` erreichbar.

Pruefen:

```bash
sudo ls -la /etc/letsencrypt/live
sudo ls -la /etc/letsencrypt/live/labbi-sites
docker compose logs nginx
curl -I https://labbi-welpen.de/healthz
```

## Rollback-Grundidee

Ein einfacher Rollback ist Git-basiert:

```bash
git log --oneline -5
git checkout <alter-commit>
export COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml
docker compose up -d --build
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
