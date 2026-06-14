# Entwicklung

## Workflow

Typischer Ablauf:

```bash
git status
git pull
cp .env.example .env # nur falls .env fehlt
go mod download
gofmt -w $(find . -name '*.go')
go test ./...
docker compose config
```

Wenn Docker verwendet wird:

```bash
docker compose up -d --build
docker compose logs -f web
```

## Branches

Es gibt keine besondere Branch-Logik im Code. Praktisch:

- Kleine Fixes auf Feature-Branch entwickeln.
- Vor Merge `gofmt`, `go test ./...`, `docker compose config`.
- `.env` und echte Secrets niemals committen.

## Wo kommt neuer Code hin?

| Art | Ziel |
|---|---|
| Neue Route | `internal/router/router.go` |
| Neuer Seitenhandler | `internal/handlers/*_handler.go` |
| Neues Template | `internal/templates/*.html` oder `internal/templates/admin/*.html` |
| Neues CSS | `static/css/*.css` |
| Neue Domain-Struktur | `internal/models` |
| Neue Neo4j-Abfrage | `internal/repository` |
| Neue Formularvalidierung | `internal/validation` |
| Neue Security-Hilfe | `internal/security` oder `internal/middleware` |
| Neue Betriebsdoku | `docs/` |

## Neue Seite ergaenzen

1. Template in `internal/templates/<seite>.html` anlegen.
2. Handler in `internal/handlers/<seite>_handler.go` anlegen.
3. Route in `internal/router/router.go` registrieren.
4. Navigation in `internal/templates/base.html` ergaenzen, falls oeffentlich.
5. Test oder manuellen Smoke-Test ergaenzen.

## Neue Admin-Seite ergaenzen

1. Template unter `internal/templates/admin/`.
2. In `renderAdminTemplate` wird automatisch `admin_base.html` mitgeladen.
3. Route unter `/admin/...` immer mit `middleware.AuthMiddleware(cfg, ...)` schuetzen.
4. POST-Routen mit CSRF schuetzen.
5. Eingaben in `internal/validation` validieren.

## Templates

Oeffentliche Templates nutzen `base.html`.

Admin-Templates nutzen `admin_base.html`.

Die Template-Funktionen aus `base_handler.go` enthalten unter anderem:

- `contains`
- `parentDogName`

## Upload-Entwicklung

Upload-Logik liegt in `internal/handlers/add_puppy_handler.go`.

Aktuelle Limits:

- maximal 10 Bilder
- maximal 5 MiB pro Datei
- maximal 25 MiB gesamt
- JPEG und PNG
- serverseitiger UUID-Dateiname

Beim Bearbeiten werden Bilder aktuell beibehalten; es gibt keinen separaten Bild-Edit-Flow.

## Tests

Vor Commit:

```bash
gofmt -w $(find . -name '*.go')
go test ./...
go vet ./...
docker compose config
```

Bestehende Testdateien:

- `internal/config/config_test.go`
- `internal/handlers/add_puppy_handler_test.go`
- `internal/handlers/contact_handler_test.go`
- `internal/security/csrf_test.go`
- `internal/validation/contact_validation_test.go`
- `internal/validation/puppy_validation_test.go`

## Keine Bildpfade aufraeumen, nur weil Dateien fehlen

In manchen Arbeitsstaenden fehlen Bilder absichtlich, damit ZIPs kleiner bleiben. Template-Referenzen, Galerie-Struktur und `static/images/generated` deshalb nicht entfernen.
