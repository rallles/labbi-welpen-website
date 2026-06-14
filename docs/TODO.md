# TODO

Diese Liste ist aus dem aktuellen Code- und Betriebsstand abgeleitet. Keine Fantasie-Features, sondern Punkte, die beim Weiterarbeiten wirklich relevant sind.

## Kritisch

- Go 1.24.1 Toolchain lokal/CI verfuegbar machen und `gofmt -w $(find . -name '*.go')` sowie `go test ./...` ausfuehren.
- Produktions-`.env` auf der AWS-Instanz pruefen: Pflichtwerte gesetzt, starke Passwoerter, keine Secrets im Git.
- Backup- und Restore-Prozess fuer `neo4j_data` und `uploads` testen und dokumentiert ablegen.
- TLS-Zertifikate fuer alle vier Nginx-Servernamen pruefen oder Multi-Domain-Zertifikat einrichten.

## Hoch

- Entscheiden, ob Upload-Dateien beim Loeschen eines Welpen entfernt werden sollen. Aktuell loescht `DeletePuppyHandler` nur den Neo4j-Knoten.
- Contact-ID-Constraint in Neo4j pruefen/ergaenzen, falls Kontaktanfragen langfristig verwaltet werden.
- Deployment-Smoke-Test nach jedem Build konsequent ausfuehren: `/healthz`, `/static/...`, `/uploads/...`, Admin Login.
- Testlauf in echter Go-Umgebung nachholen, weil diese Arbeitsumgebung zuletzt keine Go-Toolchain hatte.

## Mittel

- Backup-Befehle fuer Neo4j Community finalisieren und Restore einmal in Testumgebung ueben.
- Admin-Bildverwaltung fuer bestehende Welpen planen, falls Bilder nachtraeglich geaendert oder geloescht werden sollen.
- Kleine Integrationstests oder Smoke-Scripts fuer Docker Compose ergaenzen.
- `go vet ./...` in den Standard-Workflow aufnehmen, sobald die Toolchain vorhanden ist.

## Optional

- CI einrichten, die `gofmt`, `go test`, `go vet` und `docker compose config --no-interpolate` ausfuehrt.
- CSP planen, falls externe Skripte wie Alpine CDN ersetzt oder explizit freigegeben werden.
- Admin Auth langfristig durch Session-/Rollenmodell ersetzen, falls mehrere Benutzer oder Audit-Anforderungen entstehen.

## Spaeter

- Kontaktanfragen im Adminbereich anzeigen/verwalten, falls gewuenscht.
- Strukturierte Logs einfuehren, falls Betrieb/Monitoring umfangreicher wird.
- Metriken fuer Kontakt-Rate-Limit, Uploadfehler und Mailversand ergaenzen.

## Bewusst nicht anfassen

- Keine Bildreferenzen oder Galerie-Struktur entfernen, nur weil Bilder in einem Arbeitsstand fehlen.
- Keine echten Datenschutz-/Rechtstexte erfinden.
- Keine grosse Architektur-Migration ohne konkreten Bedarf.
