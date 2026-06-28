# Testing

## Unit Tests

Ausfuehren:

```bash
go test ./...
```

Mit Formatierung:

```bash
gofmt -w $(find . -name '*.go')
go test ./...
```

Optional:

```bash
go vet ./...
```

## Vorhandene Tests

| Datei | Abdeckung |
|---|---|
| `internal/config/config_test.go` | Pflicht-Config, keine Secret-Leaks in Fehlern, SMTP optional |
| `internal/handlers/add_puppy_handler_test.go` | Upload-Typen, Limits, leere Liste, sichere Dateinamen |
| `internal/handlers/contact_handler_test.go` | Mail-Header-Sanitizing, Reply-To, IP-Auswertung, Rate-Limiter-Cleanup, POST-Body-Limit |
| `internal/handlers/health_handler_test.go` | GET-, HEAD- und Method-Not-Allowed-Verhalten |
| `internal/security/csrf_test.go` | Token Generate, Valid, Consume, Single-Use |
| `internal/validation/contact_validation_test.go` | Kontaktformular-Validation |
| `internal/validation/puppy_validation_test.go` | Puppy-Validation, Eltern-Normalisierung, Deduplizierung |

## Manuelle Tests

### Oeffentliche Seiten

```bash
curl -I http://localhost/
curl -I http://localhost/about
curl -I http://localhost/dogs
curl -I http://localhost/puppies
curl -I http://localhost/contact
curl -I http://localhost/healthz
```

### Admin-Flow-Test

1. `/admin` oeffnen.
2. Mit `ADMIN_USER`/`ADMIN_PASSWORD` anmelden.
3. `/admin/puppies` oeffnen.
4. Add-Formular oeffnen.
5. Testwelpen nur in Entwicklungsumgebung anlegen.
6. Editieren.
7. Loeschen.

Wichtig: In Produktion nicht ohne Backup mit echten Daten testen.

### Kontaktformular-Test

1. `/contact` oeffnen.
2. Gueltige Anfrage absenden.
3. Neo4j `(:Contact)` pruefen.
4. Wenn SMTP konfiguriert ist: Mailzustellung pruefen.
5. Logs ansehen:

```bash
docker compose logs --tail=100 web
```

### Upload-Test

1. Admin Add-Formular oeffnen.
2. Kleine JPEG/PNG-Datei hochladen.
3. Nach Redirect `/admin/puppies` pruefen.
4. Bild-URL `/uploads/...` oeffnen.
5. Volume pruefen:

```bash
docker compose exec web ls -la /app/data/uploads
docker compose exec nginx ls -la /var/www/uploads
```

### Docker-Test

```bash
docker compose config
docker compose build
docker compose up -d
docker compose ps
curl -i http://localhost/healthz
```

### Deployment-Smoke-Test

```bash
curl -I https://labbi-welpen.de/
curl -I https://labbi-welpen.de/healthz
curl -I https://labbi-welpen.de/static/css/styles.css
docker compose ps
docker compose logs --tail=100 web nginx neo4j
```

## Umgebungsvoraussetzungen

Die Checks benoetigen die zu `go.mod` passende Go-Toolchain und fuer Compose-Pruefungen
das Docker-Compose-Plugin. Bei fehlenden Werkzeugen schlagen die Befehle unmittelbar fehl.
