# Routen

Die Routen werden in `internal/router/router.go` registriert.

| Route | Methode | Handler | Template | Auth | Zweck |
|---|---|---|---|---|---|
| `/` | GET | `HomeHandler` | `index.html` | Nein | Startseite |
| `/about` | GET | `AboutHandler` | `about.html` | Nein | Ueber uns |
| `/dogs` | GET | `DogsHandler` | `dogs.html` | Nein | Hunde/Elternhunde |
| `/puppies` | GET | `MakePuppiesHandler(driver)` | `puppies.html` | Nein | Neo4j-Welpenliste; bei DB-Fehler Hinweis plus feste Galerie mit HTTP 200 |
| `/list-puppies` | GET | `ListPuppiesHandler` | keins | Nein | Permanenter Redirect auf `/puppies` |
| `/contact` | GET | `ContactHandler` | `contact.html` | Nein | Kontaktformular anzeigen |
| `/contact` | POST | `ContactHandler` | `contact_result.html` oder `contact.html` | Nein | Kontakt speichern, optional Mail |
| `/impressum` | GET | `ImpressumHandler` | `impressum.html` | Nein | Impressum |
| `/datenschutz` | GET | `DatenschutzHandler` | `datenschutz.html` | Nein | Datenschutzseite |
| `/robots.txt` | GET | `RobotsHandler` | Datei `robots.txt` | Nein | Crawler-Metadaten |
| `/sitemap.xml` | GET | `SitemapHandler` | Datei `sitemap.xml` | Nein | Sitemap |
| `/healthz` | GET, HEAD | `HealthHandler` | keins | Nein | Healthcheck; GET antwortet `ok`, HEAD ohne Body |
| `/admin` | GET | `AdminDashboardHandler` | `admin/admin_dashboard.html` | Basic Auth | Admin-Dashboard |
| `/admin/puppies` | GET | `ListPuppiesAdminHandler` | `admin/admin_puppies_table.html` | Basic Auth | Welpen im Adminbereich |
| `/admin/puppies/add` | GET | `AddPuppyFormHandler` | `admin/add_puppy.html` | Basic Auth | Formular fuer neuen Welpen |
| `/admin/puppies/add` | POST | `AddPuppyHandler` | Redirect oder Formular | Basic Auth + CSRF | Welpe mit Bildern speichern |
| `/admin/puppies/edit` | GET | `EditPuppyFormHandler` | `admin/admin_puppies_edit.html` | Basic Auth | Welpe bearbeiten |
| `/admin/puppies/edit` | POST | `EditPuppySaveHandler` | Redirect oder Formular | Basic Auth + CSRF | Welpe aktualisieren |
| `/admin/puppies/delete` | POST | `DeletePuppyHandler` | Redirect | Basic Auth + CSRF | Welpe und zugehoerige Upload-Dateien loeschen |
| `/static/...` | GET | `http.FileServer` oder Nginx Alias | Datei | Nein | CSS, Icons, feste Bilder |
| `/uploads/...` | GET | `http.FileServer` oder Nginx Alias | Datei | Nein | Admin-Uploads |

## Admin-Routing

Alle Admin-Routen werden ueber `middleware.AuthMiddleware(cfg, ...)` geschuetzt.

POST-Routen erwarten `csrf_token` im Formular. Die Tokens werden nach Nutzung konsumiert.

## Static und Uploads

In Go:

```go
mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))
```

In Produktion liefert Nginx `/static/` und `/uploads/` direkt per `alias`.

## SEO-Dateien

`robots.txt` und `sitemap.xml` werden im Dockerfile ins Image kopiert. Nginx proxyt diese beiden exakten Routen an die Go-App.
