# Troubleshooting

## Neo4j startet nicht

Symptom:

- `neo4j` bleibt unhealthy.
- `web` startet nicht, weil `depends_on` auf healthy wartet.

Ursachen:

- `.env` fehlt.
- `NEO4J_USER`/`NEO4J_PASSWORD` fehlen.
- Bestehendes Neo4j-Volume hat anderes Passwort als `.env`.

Loesung:

```bash
docker compose logs neo4j
docker compose config --no-interpolate
docker volume ls | grep neo4j
```

Bei Passwortwechsel mit bestehendem Volume nicht einfach `.env` aendern und erwarten, dass Neo4j das alte Passwort vergisst. Backup/Restore oder bewusstes Volume-Reset planen.

## Login funktioniert nicht

Symptom:

- Browser fragt immer wieder nach Login.
- `/admin` liefert `401 Unauthorized`.

Ursachen:

- `ADMIN_USER` oder `ADMIN_PASSWORD` fehlt.
- Falsche Credentials.
- Browser cached Basic Auth.

Loesung:

```bash
docker compose logs web
grep ADMIN_USER .env
```

Kein Passwort in Logs oder Chat kopieren. Bei Browser-Caching privaten Tab nutzen.

## Bilder werden nicht angezeigt

Symptom:

- 404 fuer `/static/...` oder `/uploads/...`.

Ursachen:

- Bilddateien fehlen im Arbeitsstand absichtlich.
- `STATIC_DIR` oder `UPLOAD_DIR` falsch.
- Upload-Volume leer.
- Nginx-Mount falsch.

Loesung:

```bash
docker compose exec web ls -la /app/static
docker compose exec web ls -la /app/data/uploads
docker compose exec nginx ls -la /var/www/static
docker compose exec nginx ls -la /var/www/uploads
curl -I http://localhost/static/css/styles.css
```

Wichtig: Keine Bildreferenzen aus Templates entfernen, nur weil Dateien im ZIP fehlen.

## Uploads verschwinden nach Container-Neustart

Symptom:

- Nach `docker compose up -d --build` sind Upload-Bilder weg.

Ursachen:

- Uploads wurden nicht ins Volume geschrieben.
- Compose-Projektname/Volume hat gewechselt.
- Container wurde ohne Compose-Volume gestartet.

Loesung:

```bash
docker compose config --volumes
docker volume ls | grep uploads
docker compose exec web ls -la /app/data/uploads
```

## Contact Mail wird nicht gesendet

Symptom:

- Kontaktformular speichert Anfrage, aber Mail kommt nicht an.

Ursachen:

- SMTP-Konfiguration unvollstaendig.
- SMTP-Credentials falsch.
- Provider blockiert Login oder Port.
- `CONTACT_MAIL_TO` fehlt.

Loesung:

```bash
docker compose logs --tail=200 web
```

Pruefen:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USER`
- `SMTP_PASSWORD`
- `CONTACT_MAIL_TO`

SMTP ist optional; ohne vollstaendige Config wird gespeichert, aber nicht gesendet.

## Zertifikat/Nginx-Probleme

Symptom:

- HTTPS startet nicht.
- Nginx Container beendet sich.
- Browser meldet Zertifikatsfehler.

Ursachen:

- Zertifikatspfad fehlt.
- Zertifikat passt nicht zur Domain.
- Einer der vier HTTPS-Serverbloecke hat kein Zertifikat.

Loesung:

```bash
docker compose logs nginx
sudo ls -la /etc/letsencrypt/live
curl -Iv https://labbi-welpen.de/
```

Die ACME-Challenge liegt unter `internal/public/.well-known/acme-challenge`.

## Go-Version/Toolchain-Probleme

Symptom:

```text
go: command not found
gofmt: command not found
```

Ursache:

- Go ist nicht installiert oder nicht im `PATH`.

Loesung:

- Go 1.24.1 installieren.
- Shell neu oeffnen.
- `go version` pruefen.

## Docker Compose findet `.env` nicht

Symptom:

- Compose meldet `NEO4J_USER is required`.
- Services starten nicht.

Ursache:

- `.env` fehlt im Projektroot.

Loesung:

```bash
cp .env.example .env
```

Danach Werte setzen. Keine echten Secrets committen.

## `docker compose config` zeigt Secrets

Symptom:

- Ausgabe enthaelt interpolierte Werte.

Ursache:

- Compose liest `.env` und zeigt finale Config.

Loesung:

- Ausgabe nicht teilen.
- Fuer Strukturcheck ohne Interpolation:

```bash
docker compose config --no-interpolate
```
