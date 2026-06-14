# Betrieb

## Container pruefen

```bash
docker compose ps
docker compose logs --tail=100 web
docker compose logs --tail=100 nginx
docker compose logs --tail=100 neo4j
```

Live-Logs:

```bash
docker compose logs -f web nginx neo4j
```

## Healthcheck

Go-App:

```bash
curl -i http://localhost/healthz
```

Ueber Nginx/TLS:

```bash
curl -I https://labbi-welpen.de/healthz
```

Docker Healthchecks:

```bash
docker compose ps
```

`web` nutzt `wget -qO- http://localhost:8080/healthz`. `neo4j` nutzt `cypher-shell`.

## Neustart

Nur Web neu bauen/starten:

```bash
docker compose build web
docker compose up -d web
```

Alles neu starten:

```bash
docker compose up -d --build
```

Nginx neu laden ueber Container-Neustart:

```bash
docker compose restart nginx
```

## Neo4j pruefen

Im Container:

```bash
docker compose exec neo4j cypher-shell -a bolt://localhost:7687 -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" "RETURN 1"
```

Wenn Shell-Environment die Variablen nicht kennt, Werte nicht in Shell-History tippen. Besser temporaer ueber sichere Session oder Neo4j Browser mit bekannten Credentials arbeiten.

Debug-Queries stehen in [DATABASE.md](DATABASE.md).

## Backups

Mindestens sichern:

- Neo4j Volume `neo4j_data`
- Upload Volume `uploads`
- Produktions-`.env` separat im sicheren Passwort-/Secret-Speicher
- Nginx-Zertifikate werden von Let's Encrypt verwaltet, aber Pfade und Renewal sollten bekannt sein

Volume-Namen pruefen:

```bash
docker volume ls | grep labbi
docker compose config --volumes
```

Einfache Backup-Idee fuer Uploads:

```bash
docker run --rm \
  -v labbi-app_uploads:/data:ro \
  -v "$PWD/backups":/backup \
  alpine tar czf /backup/uploads-$(date +%Y%m%d).tar.gz -C /data .
```

Neo4j-Backups sollten konsistent erstellt werden. Fuer Community-Edition im Zweifel Container stoppen oder Neo4j-Dump/Export-Prozess separat festlegen und testen.

## Upload-Volume sichern

Uploads sind nicht im Git. Sie liegen im Docker-Volume `uploads` und werden als `/uploads/...` gespeichert.

Pruefen:

```bash
docker compose exec web ls -la /app/data/uploads
docker compose exec nginx ls -la /var/www/uploads
```

## Log-Hygiene

- Keine `.env` oder `docker compose config` Ausgabe mit echten Secrets weitergeben.
- App-Config-Fehler nennen nur fehlende Variablennamen.
- SMTP-/Neo4j-Fehler koennen betriebliche Details enthalten; vor Weitergabe pruefen.

## Regelmaessige Checks

- `docker compose ps`
- `docker compose logs --tail=100 web nginx neo4j`
- HTTPS-Zertifikate und Renewal pruefen
- Backup-Restore testweise durchspielen
- `/admin` Login pruefen
- Kontaktformular testen, wenn SMTP geaendert wurde
