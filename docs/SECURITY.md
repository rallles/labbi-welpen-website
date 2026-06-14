# Security

## Secrets

Secrets gehoeren in `.env` oder in den Secret-Mechanismus der Zielumgebung.

Nie committen:

- `.env`
- `.env.*`
- echte Neo4j-Passwoerter
- echte Admin-Passwoerter
- echte SMTP-Passwoerter

`.gitignore` ignoriert `.env` und `.env.*`, erlaubt aber `.env.example`.

## Config-Validierung

Pflichtwerte:

- `NEO4J_URI`
- `NEO4J_USER`
- `NEO4J_PASSWORD`
- `ADMIN_USER`
- `ADMIN_PASSWORD`

Wenn sie fehlen, bricht die App beim Start mit einer klaren Fehlermeldung ab. Die Fehlermeldung nennt Variablennamen, keine Werte.

## Admin Basic Auth

Adminrouten unter `/admin` laufen durch `middleware.AuthMiddleware`.

Eigenschaften:

- Basic Auth Realm `Admin Bereich`
- Credentials aus `ADMIN_USER` und `ADMIN_PASSWORD`
- Vergleich ueber SHA-256 und `subtle.ConstantTimeCompare`

Grenzen:

- Keine Rollen
- Keine Sessions
- Kein Logout ausser Browser-Credentials entfernen
- Fuer wenige Admins okay, fuer komplexere Anforderungen spaeter ersetzen

## CSRF

CSRF liegt in `internal/security/csrf.go`.

Eigenschaften:

- Token 32 Bytes random, base64url kodiert
- TTL 2 Stunden
- `Valid` prueft nur
- `Consume` akzeptiert frische Tokens und loescht sie sofort
- Admin-POST-Routen nutzen `Consume`

Geschuetzte POSTs:

- `/admin/puppies/add`
- `/admin/puppies/edit`
- `/admin/puppies/delete`

## Upload-Sicherheit

Upload-Schutz:

- `http.MaxBytesReader` fuer Gesamtlimit
- `ParseMultipartForm` mit Speicherlimit
- maximal 10 Bilder
- maximal 5 MiB pro Datei
- maximal 25 MiB gesamt
- `http.DetectContentType`
- `jpeg.DecodeConfig` / `png.DecodeConfig`
- keine Uebernahme von Client-Dateinamen
- Speicherung als UUID plus `.jpg` oder `.png`

Noch bewusst offen:

- Delete entfernt aktuell den Neo4j-Knoten, aber nicht automatisch die Upload-Dateien.
- Keine Virenscanner-/Content-Moderation-Pipeline.

## Kontaktformular-Schutz

Schutzmechanismen:

- Honeypot-Feld `Website`
- Rate-Limiter pro IP
- periodischer Cleanup alter IP-Keys
- serverseitige Validation
- Header-Sanitizing gegen CRLF-Injection
- `clientIP` akzeptiert nur einzelne gueltige IPs aus `X-Real-IP` oder `X-Forwarded-For`

Nginx setzt `X-Real-IP` und `X-Forwarded-For` auf `$remote_addr`, damit Forwarded-Header nicht als freie Client-Eingabe durchgereicht werden.

## Nginx Security Header

Aktuell gesetzt:

- `Strict-Transport-Security`
- `X-Content-Type-Options`
- `X-Frame-Options`
- `Referrer-Policy`
- `Permissions-Policy`

Keine CSP einfuehren, solange externe Skripte wie Alpine CDN genutzt werden, ohne die Policy bewusst zu planen.

## Vor Produktivbetrieb pruefen

- Echte starke Werte in `.env`
- `.env` nicht getrackt
- `docker compose config` nicht mit Secrets weitergeben
- Zertifikate fuer alle Domains
- Admin-Passwort im Passwortmanager
- Backup fuer Neo4j und Uploads
- SMTP-Konfiguration, falls Kontaktmails erwartet werden
- `go test ./...` erfolgreich in einer Umgebung mit Go 1.24.1
