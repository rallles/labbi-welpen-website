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

## Oeffentliche Welpenliste zeigt einen Ladehinweis

Symptom:

- `/puppies` zeigt feste Inhalte und Galerie, aber statt der dynamischen Welpenliste einen Ladehinweis.

Ursache:

- Neo4j ist voruebergehend nicht erreichbar oder die Abfrage ist fehlgeschlagen.

Pruefen:

```bash
docker compose ps
docker compose logs --tail=100 web neo4j
curl -i http://localhost:8081/healthz
```

Der HTTP-Healthcheck bestaetigt die Erreichbarkeit der Go-App, ersetzt aber nicht die
Neo4j-Pruefung. Nach Wiederherstellung der DB-Verbindung laedt `/puppies` beim naechsten
Request wieder dynamische Daten.

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

- SMTP ist absichtlich nicht konfiguriert; dann ist dieses Verhalten korrekt.
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

SMTP ist optional. Nur wenn alle fuenf Werte gesetzt sind, wird ein Versand versucht.
Ohne vollstaendige Config wird die Anfrage in Neo4j gespeichert (`MailSent=false`,
`MailError=smtp_not_configured`). Bei einem echten Versandfehler bleibt sie ebenfalls
gespeichert (`MailError=smtp_send_failed`). In der Ergebnisansicht wird der jeweilige
Zustand ohne technische oder geheime Details angezeigt.

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
