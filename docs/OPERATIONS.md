# Betrieb

## Routine nach lokalem Deployment

Die Routine validiert und baut die lokale Compose-Umgebung neu. Der anschliessende
Smoke-Test fuehrt ausschliesslich lesende GET-/HEAD-Requests aus und veraendert keine Daten:

```bash
git status
git pull
docker compose config
docker compose build --no-cache
docker compose up -d
docker compose ps
docker compose logs --tail=100 web nginx neo4j
./scripts/smoke-local.sh
```

Falls das Executable-Bit in einer Windows-Arbeitskopie fehlt, ist derselbe Test mit
`sh scripts/smoke-local.sh` ausfuehrbar. Der lokale Nginx muss auf Port 8081 erreichbar sein.

Auf AWS danach zusaetzlich die oeffentlichen, rein lesenden Endpunkte pruefen:

```bash
curl -i https://labbi-welpen.de/healthz
curl -I https://labbi-welpen.de/healthz
curl -I https://labbi-welpen.de/static/css/styles.css
curl -I https://labbi-welpen.de/robots.txt
curl -I https://labbi-welpen.de/sitemap.xml
```

Das lokale Script nicht ungeprueft gegen die Produktionsdomain umkonfigurieren.

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
curl -I http://localhost/healthz
```

`GET` liefert `ok`; `HEAD` liefert denselben Status 200 ohne Response-Body.

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

Als zusammengehoerigen Sicherungsstand sichern:

- Neo4j Volume `neo4j_data`
- Upload Volume `uploads`
- Produktions-`.env` separat im sicheren Passwort-/Secret-Speicher
- Nginx-Zertifikate werden von Let's Encrypt verwaltet, aber Pfade und Renewal sollten bekannt sein

Volume-Namen pruefen:

```bash
docker volume ls | grep labbi
docker compose config --volumes
```

`neo4j_data` und `uploads` muessen gemeinsam und aus demselben Wartungsfenster gesichert
werden, damit Datenbank-Bildreferenzen und Dateien zusammenpassen. Es gibt bewusst noch
keinen produktiven Backup-Befehl, solange das Verfahren nicht end-to-end getestet ist.

Vor Aufnahme in das Produktions-Runbook:

1. Backup beider Volumes in einen isolierten Testordner erstellen.
2. Beide Sicherungen in einem frischen Compose-Projekt wiederherstellen.
3. Neo4j-Daten, Upload-Dateien und deren Referenzen gemeinsam pruefen.
4. Erst den erfolgreich getesteten Ablauf als produktiven Runbook-Befehl dokumentieren.

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
