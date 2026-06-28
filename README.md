# Labbi-Welpen-App

Go-Webapp fuer die Labbi-/Welpen-Website mit oeffentlicher Website, Adminbereich, Neo4j, Bild-Uploads, Kontaktformular, Docker Compose, Nginx Reverse Proxy und AWS-Deployment.

Diese Dokumentation ist als Wiedereinstieg gedacht: Wenn du das Projekt nach Monaten wieder oeffnest, findest du hier Startbefehle, Routen, Datenmodell, Betriebsbefehle und die wichtigsten Fallstricke.

## Projektueberblick

Die Labbi-Welpen-App betreibt eine Website fuer eine Labrador-Hobbyzucht. Sie loest drei praktische Probleme:

- Oeffentliche Inhalte wie Startseite, Ueber uns, Hunde, Welpen, Kontakt, Impressum und Datenschutz werden als Go-Templates ausgeliefert.
- Welpen koennen im Adminbereich angelegt, bearbeitet, geloescht und mit Bildern versehen werden.
- Dynamische Daten wie Welpen, Eltern-Beziehungen und Kontaktanfragen werden in Neo4j gespeichert.

Der Stack:

- Go 1.24.1, Standard `net/http`
- Neo4j 5.26.1 Community
- Dockerfile mit Multi-Stage-Build
- Docker Compose fuer `web`, `neo4j`, `nginx`
- Nginx fuer TLS, Reverse Proxy, `/static/`, `/uploads/` und ACME-Challenge
- Basic Auth und CSRF fuer Admin-POST-Routen

Detaildokumente:

- [Architektur](docs/ARCHITECTURE.md)
- [Setup](docs/SETUP.md)
- [Deployment](docs/DEPLOYMENT.md)
- [Betrieb](docs/OPERATIONS.md)
- [Entwicklung](docs/DEVELOPMENT.md)
- [Security](docs/SECURITY.md)
- [Datenbank](docs/DATABASE.md)
- [Routen](docs/ROUTES.md)
- [Assets und Uploads](docs/ASSETS_AND_UPLOADS.md)
- [Tests](docs/TESTING.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [TODO](docs/TODO.md)
- [Changelog](docs/CHANGELOG.md)
- [Architecture Decision Records](docs/DECISIONS/)

## Aktueller Projektstatus

Was funktioniert aktuell:

- Oeffentliche Seiten werden ueber `internal/templates` gerendert.
- `/puppies` laedt Welpen aus Neo4j; `/list-puppies` leitet dauerhaft auf diese kanonische Route um.
- Adminbereich ist per Basic Auth geschuetzt.
- Admin-POST-Routen fuer Add/Edit/Delete nutzen Single-Use-CSRF-Tokens.
- Welpen werden in Neo4j als `(:Puppy)` gespeichert.
- Elternhunde `gandalf`, `anna`, `brina` werden beim Start als `(:Dog)` geseedet.
- Beziehung `(:Puppy)-[:HAS_PARENT]->(:Dog)` wird beim Speichern gepflegt.
- Uploads werden als JPEG/PNG validiert, mit serverseitigem UUID-Dateinamen gespeichert und ueber `/uploads/...` ausgeliefert.
- Kontaktformular speichert `(:Contact)` und kann optional SMTP-Mail senden.
- Kontaktformular hat Honeypot, Rate-Limit und Header-Sanitizing.
- Docker Compose definiert `neo4j`, `web`, `nginx` mit Healthchecks.
- Nginx liefert `/static/` aus `./static`, `/uploads/` aus Volume und ACME-Challenges aus `./internal/public`.

Produktionsnah:

- Docker/Nginx/AWS-Struktur ist vorhanden.
- Security Header sind in Nginx gesetzt: HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy.
- `.env` und `.env.*` sind ignoriert; `.env.example` bleibt versioniert.
- Config-Validierung bricht bei fehlenden Pflichtwerten ab, ohne Secret-Werte zu loggen.

Noch offen:

- Backup-/Restore-Prozess fuer Neo4j und Upload-Volume sollte regelmaessig getestet werden.
- TLS-Zertifikate muessen auf der AWS-Instanz fuer alle Nginx-Servernamen vorhanden sein oder als Multi-Domain-Zertifikat eingerichtet werden.
- Admin-Login ist Basic Auth; fuer wenige Admins okay, aber nicht mit Rollen/Sessions vergleichbar.

Bewusst statisch:

- Layout, Navigation, Hunde-/Zucht-Content und feste Bildreferenzen in Templates/CSS.
- Optimierte Website-Bilder unter `static/images/generated/` sollen versioniert bleiben.
- Fehlende Bilder in diesem Arbeitsstand sind absichtlich kein Anlass, Pfade oder Galerie-Struktur zu entfernen.

Dynamisch aus Neo4j:

- Oeffentliche Welpenliste unter `/puppies` und Admin-Welpentabelle.
- Einzelne Welpen beim Bearbeiten.
- Kontaktanfragen und Mailstatus.
- Eltern-Beziehungen zwischen Welpen und Parent-Dogs.

## Schnellstart Lokal

Voraussetzungen:

- Go 1.24.1 passend zu `go.mod`
- Laufende Neo4j-Instanz oder Docker Compose fuer Neo4j
- `.env` mit lokalen Werten

```bash
cp .env.example .env
go mod download
go test ./...
go run ./cmd
```

Die `.env.example` ist fuer den direkten lokalen Start vorbereitet. Sie verwendet
`NEO4J_URI=bolt://localhost:7687`; passe bei Bedarf insbesondere die Platzhalter-Passwoerter an:

```env
SERVER_ADDRESS=:8080
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=change_me_strong_neo4j_password
ADMIN_USER=admin
ADMIN_PASSWORD=change_me_strong_admin_password
UPLOAD_DIR=data/uploads
STATIC_DIR=static
TEMPLATE_DIR=internal/templates
```

Danach:

- Website: `http://localhost:8080/`
- Healthcheck: `http://localhost:8080/healthz`
- Admin: `http://localhost:8080/admin`

Beim Docker-Start ueberschreibt `docker-compose.yml` die App-interne URI im `web`-Service
mit `NEO4J_URI=bolt://neo4j:7687`. Auch Serveradresse sowie Upload-, Static- und
Template-Verzeichnisse werden dort auf Containerpfade gesetzt.

## Schnellstart mit Docker Compose

Unter Windows/Docker Desktop wird der lokale Override per Semikolon aktiviert:

```powershell
cp .env.example .env
$env:COMPOSE_FILE="docker-compose.yml;docker-compose.local.yml"
docker compose config
docker compose up -d --build
docker compose logs -f web nginx neo4j
```

Wichtige Checks:

```powershell
docker compose ps
docker compose logs --tail=100 web
docker compose logs --tail=100 neo4j
docker compose logs --tail=100 nginx
curl -i http://localhost:8081/healthz
```

Lokale Website: `http://localhost:8081`

`COMPOSE_FILE` verwendet betriebssystemabhaengige Trennzeichen: Windows nutzt ein
Semikolon (`docker-compose.yml;docker-compose.local.yml`), Linux und AWS einen
Doppelpunkt (`docker-compose.yml:docker-compose.local.yml`).

Hinweis: `docker compose config` interpoliert Werte aus `.env`. Ausgabe nicht in Issues, Logs oder Dokumentation kopieren, wenn echte Secrets enthalten sind.

Die `.dockerignore` schliesst Markdown-Dokumentation aus dem Runtime-Build-Kontext aus.
Ausserdem bleiben `.env`, `.env.*`, Schluesseldateien und Secret-Verzeichnisse ausgeschlossen;
nur die versionierte `.env.example` mit Platzhaltern ist ausdruecklich erlaubt.

### Routine nach lokalem Deployment

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

Das Smoke-Script prueft nur lesend Healthcheck, Startseite, Welpenseite, `robots.txt`
und Sitemap auf `http://localhost:8081`. Unter Windows kann es mit
`sh scripts/smoke-local.sh` gestartet werden.

## AWS-Deployment

Typischer Ablauf auf der AWS-Instanz per VSCode SSH:

```bash
cd /pfad/zur/labbi-app
git status
git pull
test -f .env
grep '^COMPOSE_FILE=' .env || true
export COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml
docker compose config
docker compose up -d --build
docker compose ps
docker compose logs -f web nginx neo4j
```

Auf AWS vor jedem Deployment pruefen, dass `.env` vorhanden ist und keine Windows-
`COMPOSE_FILE`-Angabe mit Semikolon aktiv ist. Falls noetig, in `.env`
`COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml` setzen oder den Wert wie oben
fuer den Deployment-Lauf exportieren. Keine Produktionswerte in Git oder Dokumentation uebernehmen.

Auf der Instanz muessen vorhanden sein:

- Repository mit `Dockerfile`, `docker-compose.yml`, `docker-compose.aws.yml` und `nginx.aws.conf`
- `.env` mit echten Produktionswerten, nicht committed
- `/etc/letsencrypt/...` mit Zertifikaten fuer alle konfigurierten Domains oder Multi-Domain-Zertifikat
- Docker und Docker Compose Plugin
- Persistente Docker-Volumes `neo4j_data` und `uploads`

Healthchecks:

```bash
curl -i https://labbi-welpen.de/healthz
curl -I https://labbi-welpen.de/healthz
curl -I https://labbi-welpen.de/static/css/styles.css
curl -I https://labbi-welpen.de/robots.txt
curl -I https://labbi-welpen.de/sitemap.xml
```

Diese Produktions-Checks sind rein lesend und werden bewusst nicht vom lokalen Script ausgefuehrt.

Logs:

```bash
docker compose logs -f web
docker compose logs -f nginx
docker compose logs -f neo4j
```

## Environment-Variablen

| Variable | Pflicht | Beispielwert ohne echte Secrets | Beschreibung |
|---|---:|---|---|
| `SERVER_ADDRESS` | Optional | `0.0.0.0:8080` oder `:8080` | Adresse, auf der die Go-App lauscht. Default im Code: `:8080`. |
| `NEO4J_URI` | Ja fuer App-Start ohne Compose | `bolt://localhost:7687` | URI fuer `go run ./cmd`; Docker Compose ueberschreibt sie im `web`-Service mit `bolt://neo4j:7687`. |
| `NEO4J_USER` | Ja | `neo4j` | Neo4j Benutzer. |
| `NEO4J_PASSWORD` | Ja | `change_me_strong_neo4j_password` | Neo4j Passwort. Niemals committen. |
| `ADMIN_USER` | Ja | `admin_user` | Benutzer fuer Basic Auth im Adminbereich. |
| `ADMIN_PASSWORD` | Ja | `change_me_strong_admin_password` | Passwort fuer Basic Auth. Niemals committen. |
| `SMTP_HOST` | Optional | `smtp.invalid` | SMTP Host fuer Kontaktbenachrichtigungen. |
| `SMTP_PORT` | Optional | `587` | SMTP Port. |
| `SMTP_USER` | Optional | `sender@example.invalid` | SMTP Benutzer und From-Adresse. |
| `SMTP_PASSWORD` | Optional | `change_me_smtp_password` | SMTP Passwort. Niemals committen. |
| `CONTACT_MAIL_TO` | Optional | `contact@example.invalid` | Empfaengeradresse fuer Kontaktmail. |
| `UPLOAD_DIR` | Optional | `/app/data/uploads` | Speicherort fuer Admin-Uploads. Default: `data/uploads`. |
| `STATIC_DIR` | Optional | `/app/static` | Verzeichnis fuer `/static/...`. Default: `static`. |
| `TEMPLATE_DIR` | Optional | `/app/templates` | Template-Verzeichnis. Default: `internal/templates`. |

SMTP ist optional. Wenn SMTP nicht vollstaendig konfiguriert ist, wird die Kontaktanfrage gespeichert, aber keine Benachrichtigung versendet.

## Projektstruktur

| Pfad | Zweck |
|---|---|
| `cmd/` | Einstiegspunkt der Go-App, `.env`-Laden, Config-Validierung, Serverstart. |
| `internal/config` | Environment-Konfiguration und Pflichtwert-Validierung. |
| `internal/database` | Neo4j-Driver, Constraints, Seed der Parent-Dogs. |
| `internal/handlers` | HTTP-Handler, Template-Rendering, Kontaktformular, Admin, Uploads. |
| `internal/middleware` | Basic Auth Middleware fuer Adminbereich. |
| `internal/models` | Domain-Modelle `Puppy`, `Dog`, `Contact`, Fellfarben, Parent-Dogs. |
| `internal/repository` | Neo4j-Zugriff fuer Puppies und Contacts. |
| `internal/router` | Registrierung aller Routen und Static/Upload FileServer. |
| `internal/security` | CSRF-Store mit Single-Use-Consume. |
| `internal/validation` | Validierung fuer Puppy- und Contact-Formulare. |
| `internal/templates` | Go HTML Templates fuer Website und Admin. Im Docker-Image nach `/app/templates` kopiert. |
| `static` | CSS, Icons und feste Bilder. Im Docker-Image nach `/app/static` kopiert. |
| `static/images/generated` | Optimierte Bildderivate; bewusst versioniert. |
| `internal/public` | ACME-Challenge Webroot fuer Nginx. |
| `docs` | Projektdokumentation und ADRs. |

## Wichtige Routen

| Route | Methode | Zweck | Auth | Handler/Template |
|---|---|---|---|---|
| `/` | GET | Startseite | Nein | `HomeHandler`, `index.html` |
| `/about` | GET | Ueber uns | Nein | `AboutHandler`, `about.html` |
| `/dogs` | GET | Eltern-/Hunde-Seite | Nein | `DogsHandler`, `dogs.html` |
| `/puppies` | GET | Oeffentliche Welpenliste aus Neo4j plus feste Galerie | Nein | `MakePuppiesHandler(driver)`, `puppies.html` |
| `/list-puppies` | GET | Permanenter Redirect auf `/puppies` | Nein | `ListPuppiesHandler` |
| `/contact` | GET | Kontaktformular | Nein | `ContactHandler`, `contact.html` |
| `/contact` | POST | Kontakt absenden, speichern, optional mailen | Nein | `ContactHandler`, `contact_result.html` |
| `/impressum` | GET | Impressum | Nein | `ImpressumHandler`, `impressum.html` |
| `/datenschutz` | GET | Datenschutzseite | Nein | `DatenschutzHandler`, `datenschutz.html` |
| `/admin` | GET | Admin-Dashboard | Basic Auth | `AdminDashboardHandler`, `admin/admin_dashboard.html` |
| `/admin/puppies` | GET | Admin-Welpen-Tabelle | Basic Auth | `ListPuppiesAdminHandler`, `admin/admin_puppies_table.html` |
| `/admin/puppies/add` | GET | Welpe anlegen Formular | Basic Auth | `AddPuppyFormHandler`, `admin/add_puppy.html` |
| `/admin/puppies/add` | POST | Welpe speichern mit Uploads | Basic Auth + CSRF | `AddPuppyHandler` |
| `/admin/puppies/edit` | GET | Welpe bearbeiten Formular | Basic Auth | `EditPuppyFormHandler`, `admin/admin_puppies_edit.html` |
| `/admin/puppies/edit` | POST | Welpe aktualisieren | Basic Auth + CSRF | `EditPuppySaveHandler` |
| `/admin/puppies/delete` | POST | Welpe loeschen | Basic Auth + CSRF | `DeletePuppyHandler` |
| `/static/...` | GET | CSS, Icons, feste Bilder | Nein | Go FileServer / Nginx Alias |
| `/uploads/...` | GET | Admin-Uploads | Nein | Go FileServer / Nginx Alias |
| `/robots.txt` | GET | Robots-Datei | Nein | `RobotsHandler` |
| `/sitemap.xml` | GET | Sitemap | Nein | `SitemapHandler` |
| `/healthz` | GET, HEAD | Healthcheck (`HEAD` ohne Body) | Nein | `HealthHandler` |

Mehr Details: [docs/ROUTES.md](docs/ROUTES.md).

## Datenmodell Neo4j

Knoten:

- `(:Puppy)` fuer Welpen
- `(:Dog)` fuer Elternhunde
- `(:Contact)` fuer Kontaktanfragen

Relationship:

- `(:Puppy)-[:HAS_PARENT]->(:Dog)`

Constraints beim Start:

```cypher
CREATE CONSTRAINT puppy_id IF NOT EXISTS FOR (p:Puppy) REQUIRE p.id IS UNIQUE;
CREATE CONSTRAINT dog_id IF NOT EXISTS FOR (d:Dog) REQUIRE d.id IS UNIQUE;
CREATE CONSTRAINT contact_id IF NOT EXISTS FOR (c:Contact) REQUIRE c.id IS UNIQUE;
```

Parent-Dogs werden beim Start geseedet:

- `gandalf`
- `anna`
- `brina`

Wichtige `Puppy`-Properties:

- Pflicht aus Formular/Validierung: `id`, `name`, `geburtsdatum`, `geschlecht`, `farbe`, `gewicht`
- Weitere Felder: `charakter`, `geimpft`, `gechippt`, `entwurmt`, `eltern`, `notizen`, `bilder`

Wichtige `Contact`-Properties:

- `id`, `name`, `email`, `phone`, `message`, `createdAt`, `mailSent`, `mailError`

Mehr Details und Debug-Cypher: [docs/DATABASE.md](docs/DATABASE.md).

## Adminbereich

- Adminrouten liegen unter `/admin`.
- Zugriff per HTTP Basic Auth aus `ADMIN_USER` und `ADMIN_PASSWORD`.
- Vergleich erfolgt constant-time ueber SHA-256 Hashes.
- Admin-POST-Routen nutzen CSRF-Tokens.
- CSRF-Tokens sind Single-Use: Nach erfolgreichem `Consume` ist der Token geloescht.
- Puppy-Formulare werden serverseitig validiert.
- Add-Flow akzeptiert JPEG/PNG Uploads, maximal 10 Bilder, maximal 5 MiB pro Datei und 25 MiB total.
- Upload-Dateinamen werden nicht uebernommen; gespeichert wird UUID plus validierte Extension.
- Delete loescht den `Puppy`-Knoten per `DETACH DELETE`, entfernt aber aktuell nicht automatisch Upload-Dateien vom Volume.

## Uploads und Assets

Klare Trennung:

- `/static/...`: feste Website-Dateien aus `static/`.
- `/uploads/...`: dynamische Admin-Uploads aus `UPLOAD_DIR`.
- `static/images/generated/`: optimierte, versionierte Bildderivate.

Docker:

- `web` schreibt Uploads nach `/app/data/uploads`.
- Compose bindet das Volume `uploads` nach `/app/data/uploads`.
- `nginx` liest dasselbe Volume read-only unter `/var/www/uploads`.
- `nginx` liest `./static` read-only unter `/var/www/static`.

Wichtig: Fehlende Bilder in einem ZIP- oder Arbeitsstand sind kein Grund, Templatepfade oder Galerie-Struktur zu loeschen.

## Tests und Qualitaet

Alles lokal mit einem Befehl pruefen:

```powershell
.\scripts\check-local.ps1
```

Unter Linux, AWS oder Git Bash:

```bash
sh scripts/check-local.sh
```

Die Skripte pruefen Formatierung, Tests, Vet und die Docker-Compose-Konfiguration.
`docker compose config` validiert nur und startet keine Container.

Die bestehende GitHub-Actions-CI unter `.github/workflows/ci.yml` prueft ebenfalls
`gofmt`, `go test ./...`, `go vet ./...` und `docker compose config`.

Einzelschritte:

```bash
gofmt -w $(find . -name '*.go')
go test ./...
go vet ./...
docker compose config
```

Hinweis: `gofmt ./...` ist als Wunsch oft gemeint, aber `gofmt` arbeitet mit Dateien. Nutze den oben stehenden `find`-Befehl.

Aktuell vorhandene Testbereiche:

- Config-Validierung
- Contact-Mail-Header und IP-Erkennung
- Contact-Validation
- CSRF Generate/Valid/Consume
- Puppy-Validation
- Upload-Validierung

## Haeufige Fehler

| Symptom | Wahrscheinliche Ursache | Kurzloesung |
|---|---|---|
| Neo4j startet nicht | `.env` fehlt oder `NEO4J_AUTH` ungueltig | `.env` pruefen, `docker compose logs neo4j` |
| Login funktioniert nicht | `ADMIN_USER`/`ADMIN_PASSWORD` fehlen oder Browser cached Basic Auth | `.env` pruefen, privaten Browser nutzen |
| Bilder werden nicht angezeigt | Bilddateien fehlen absichtlich, falsches Volume, falscher Pfad | `/static/...` vs `/uploads/...` pruefen |
| Uploads verschwinden | Kein persistentes `uploads` Volume oder falsches Compose-Projekt | `docker volume ls`, Compose-Namen pruefen |
| Contact Mail wird nicht gesendet | SMTP unvollstaendig oder Credentials falsch | `SMTP_*` und `CONTACT_MAIL_TO` pruefen |
| Zertifikat/Nginx kaputt | Zertifikat fehlt fuer Domain | `/etc/letsencrypt/live/...` pruefen |
| Go-Befehle gehen nicht | Go-Toolchain fehlt oder falsche Version | `go version`, Go 1.24.1 installieren |
| Compose findet `.env` nicht | `.env` nicht im Projektroot | `cp .env.example .env`, Werte setzen |

Mehr: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md).

## Wiederaufnahme nach laengerer Pause

Erste Befehle:

```bash
git status
git pull
cp .env.example .env # nur falls .env fehlt
go version
go test ./...
docker compose config
docker compose up -d --build
docker compose logs -f
```

Dann pruefen:

- Offene Aufgaben: [docs/TODO.md](docs/TODO.md)
- Erste Datei zum Lesen: `README.md`, danach [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Architekturentscheidungen: [docs/DECISIONS/](docs/DECISIONS/)
- Secrets niemals committen: `.env`, `.env.*`, echte Neo4j/Admin/SMTP-Passwoerter
- Website laeuft, wenn `/healthz` `ok` liefert und Nginx/Web/Neo4j `healthy` sind:

```bash
curl -i http://localhost:8081/healthz
docker compose ps
docker compose logs --tail=100 web nginx neo4j
```

## Aktueller naechster Arbeitsstand

Kritisch:

- Produktions-`.env` auf der AWS-Instanz pruefen, ohne Werte zu committen.
- `neo4j_data` und `uploads` gemeinsam in einen Testordner sichern, in einem frischen
  Compose-Projekt wiederherstellen und erst danach einen produktiven Runbook-Befehl festlegen.

Hoch:

- Nginx-Zertifikate fuer alle vier Servernamen pruefen.
- Lokalen Smoke-Test nach Deployments und die dokumentierten AWS-Checks konsequent ausfuehren.

Bewusst verschoben:

- Keine CSP, solange Alpine per CDN eingebunden ist.
- Keine grosse Auth-Architektur statt Basic Auth.
- Keine Umstrukturierung der Bild-/Galeriepfade wegen fehlender Bilddateien.

Details: [docs/TODO.md](docs/TODO.md).
