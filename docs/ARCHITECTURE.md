# Architektur

## Gesamtbild

Die Labbi-Welpen-App ist eine klassische serverseitige Go-Webapp:

```text
Browser
  |
  | HTTPS
  v
Nginx
  |-- /static/  -> ./static
  |-- /uploads/ -> Docker Volume uploads
  |-- ACME      -> ./internal/public
  |
  | Reverse Proxy
  v
Go-App labbi-web
  |
  | Neo4j Bolt
  v
Neo4j
```

Im lokalen Go-Start ohne Docker kann die App auch direkt auf `:8080` laufen und `/static/` sowie `/uploads/` selbst ueber `http.FileServer` ausliefern.

## Go-App

Der Einstieg liegt in `cmd/main.go`.

Startablauf:

1. Optional `.env` aus dem Projektroot laden.
2. `config.LoadConfig()` liest Environment-Variablen.
3. `cfg.Validate()` prueft Pflichtwerte.
4. `database.NewNeo4jDriver(cfg)` erstellt den Driver, legt Constraints an und seedet Parent-Dogs.
5. `router.SetupRoutes(...)` registriert Handler, Admin-Middleware und FileServer.
6. `http.Server` startet mit Timeouts.

Wichtige Packages:

- `internal/config`: Environment-Konfiguration
- `internal/database`: Neo4j Driver, Constraints, Seed
- `internal/router`: HTTP-Routen
- `internal/handlers`: Seiten, Admin, Kontakt, Uploads
- `internal/repository`: Neo4j Queries
- `internal/security`: CSRF Store
- `internal/validation`: Formularvalidierung

## Nginx Reverse Proxy

`nginx.conf` hat:

- Port 80 fuer Redirect nach HTTPS und ACME-Challenge.
- Vier HTTPS-Serverbloecke fuer `labbi-welpen.de`, `www.labbi-welpen.de`, `labbi-hobby.de`, `www.labbi-hobby.de`.
- Hauptdomain proxyt zur Go-App.
- Neben-Domains redirecten auf `https://labbi-welpen.de`.
- `/static/` wird direkt aus `/var/www/static/` geliefert.
- `/uploads/` wird direkt aus `/var/www/uploads/` geliefert.
- `X-Real-IP` und `X-Forwarded-For` werden bewusst auf `$remote_addr` gesetzt.

Wichtig: Jeder HTTPS-Serverblock braucht ein passendes Zertifikat oder ein Multi-Domain-Zertifikat.

## Neo4j

Neo4j speichert:

- `(:Puppy)` fuer Welpen
- `(:Dog)` fuer Elternhunde
- `(:Contact)` fuer Kontaktanfragen
- `(:Puppy)-[:HAS_PARENT]->(:Dog)`

Beim Start werden Constraints fuer `Puppy.id`, `Dog.id` und `Contact.id` erstellt. Parent-Dogs werden per `MERGE` geseedet.

## Docker Compose

Services:

- `neo4j`: Neo4j 5.26.1 Community, Volume `neo4j_data`, Healthcheck via `cypher-shell`.
- `web`: Go-App aus `Dockerfile`, Upload-Volume nach `/app/data/uploads`, Healthcheck auf `/healthz`.
- `nginx`: Nginx Alpine, bindet `nginx.conf`, `internal/public`, `static`, Upload-Volume und `/etc/letsencrypt`.

Netzwerk:

- `labbi-net` Bridge Netzwerk.

Volumes:

- `neo4j_data`: persistente Datenbank.
- `uploads`: persistente Admin-Uploads.

## Static Assets

Static Assets liegen im Repo unter `static/`.

Im Docker-Image:

- `static` -> `/app/static`

In Nginx:

- `./static` -> `/var/www/static:ro`

Die Go-App kennt das Verzeichnis ueber `STATIC_DIR`.

## Upload Volume

Uploads kommen aus dem Admin-Formular `/admin/puppies/add`.

Flow:

1. Multipart-Form wird mit Limits geparst.
2. CSRF wird konsumiert.
3. Puppy-Form wird validiert.
4. Bilder werden als JPEG/PNG erkannt und decodiert.
5. Dateien werden mit UUID-Dateinamen in `UPLOAD_DIR` gespeichert.
6. Neo4j speichert relative Pfade wie `/uploads/<uuid>.jpg`.

Nginx und Go-App koennen `/uploads/` ausliefern. Im Produktivbetrieb liefert Nginx direkt aus dem Volume.

Beim Loeschen eines Welpen laedt der Handler zuerst dessen Bildpfade, loescht danach den
Neo4j-Knoten und entfernt anschliessend nur Dateien mit einem oeffentlichen `/uploads/`-Pfad.
Dateinamen werden dabei mit `filepath.Base` auf das konfigurierte `UPLOAD_DIR` begrenzt.
Fehlschlaegt nur die Dateibereinigung, bleibt der DB-Delete bestehen und der Admin sieht eine Warnung.

## Request Flow

Oeffentliche Seite:

```text
Browser -> Nginx -> Go Router -> Handler -> Template -> Response
```

Welpenliste:

```text
Browser -> Nginx -> Go Handler -> PuppyRepository -> Neo4j -> Template
```

`/puppies` rendert die mit `PuppyRepository.List` geladenen Welpen. Die bestehende feste
Wurfgalerie im selben Template ist davon getrennt und bleibt statischer Website-Content.
`/list-puppies` leitet permanent auf `/puppies` um.

Admin-POST:

```text
Browser -> Nginx -> Basic Auth -> Handler -> CSRF Consume -> Validation -> Repository/Upload -> Redirect
```

Kontakt:

```text
Browser -> Nginx -> Handler -> Honeypot/RateLimit/Validation -> ContactRepository -> optional SMTP -> Result
```
