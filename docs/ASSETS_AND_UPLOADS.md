# Assets und Uploads

## Grundregel

Nicht vermischen:

- `static/` ist fuer versionierte, feste Website-Dateien.
- `uploads/` ist fuer dynamische Admin-Uploads und liegt nicht im Git.
- `static/images/generated/` sind optimierte, versionierte Bildderivate.

Konkrete Zuordnung:

- `static/images/generated/` enthaelt optimierte Website-Bilder (Derivate fuer das Webdesign).
- `static/images/` enthaelt feste Website-Originale sowie Icon-/Favicon-Dateien.
- `/uploads/` enthaelt Admin-Uploads und liegt zur Laufzeit im Docker-Volume.

Wichtig:

- Admin-Uploads gehoeren nicht nach `static/images/`. Sie werden ueber den Adminbereich
  erzeugt und im Upload-Volume unter `/uploads/...` abgelegt.
- Grosse Originalbilder in `static/images/` erhoehen die Groesse des Docker-Images, da
  `static/` in das Runtime-Image kopiert wird.
- Optional spaeter: Originale aus dem Runtime-Image ausschliessen oder in
  `static/images/originals/` verschieben und nur die `generated/`-Derivate ausliefern.

Fehlende Bilder in einem Arbeitsstand sind bewusst moeglich. Keine Template-Referenzen oder Galerie-Strukturen entfernen, nur weil Dateien lokal fehlen.

## Static Assets

Pfad im Repo:

```text
static/
  css/
  images/
  images/generated/
```

Verwendung:

- CSS: `/static/css/...`
- Icons/Favicons: `/static/images/...`
- feste Bilder: `/static/images/...`
- optimierte Derivate: `/static/images/generated/...`

Dockerfile:

```dockerfile
COPY --from=builder /app/static ./static
ENV STATIC_DIR="/app/static"
```

Compose/Nginx:

```yaml
- ./static:/var/www/static:ro
```

Nginx:

```nginx
location /static/ {
    alias /var/www/static/;
}
```

## Generated Images

`static/images/generated/` ist in `.gitignore` explizit wieder erlaubt:

```gitignore
!static/images/generated/
```

Diese Dateien sind Website-Assets, keine Admin-Uploads.

## Uploads

Uploads entstehen im Adminbereich beim Anlegen eines Welpen.

Konfiguration:

- lokal: `UPLOAD_DIR=data/uploads`
- Docker: `UPLOAD_DIR=/app/data/uploads`

Compose:

```yaml
volumes:
  - uploads:/app/data/uploads
```

Nginx:

```yaml
- uploads:/var/www/uploads:ro
```

```nginx
location /uploads/ {
    alias /var/www/uploads/;
}
```

In Neo4j werden relative Pfade gespeichert:

```text
/uploads/<uuid>.jpg
/uploads/<uuid>.png
```

## Upload-Validierung

Aktuell:

- JPEG/PNG
- Content-Type-Erkennung
- DecodeConfig-Pruefung
- maximal 10 Bilder
- maximal 5 MiB je Datei
- maximal 25 MiB gesamt
- UUID-Dateiname statt Client-Dateiname

## Typische Fehler bei Bildpfaden

| Symptom | Ursache | Loesung |
|---|---|---|
| `/static/...` 404 | Datei fehlt oder `STATIC_DIR` falsch | Pfad und Env pruefen |
| `/uploads/...` 404 | Upload-Volume leer/falsch | `docker compose exec web ls /app/data/uploads` |
| Bilder nach Restart weg | Uploads nicht im Volume gespeichert | Compose-Volume `uploads` pruefen |
| Nginx liefert Upload nicht | Alias/Volume falsch | `nginx.conf` und Compose-Mount pruefen |
| ZIP enthaelt keine Bilder | bewusst kleiner Arbeitsstand | Pfade nicht entfernen |

## Was versionieren?

Versionieren:

- CSS
- SVG/Icons
- feste Website-Bilder, wenn sie Teil des Webdesigns sind
- `static/images/generated`

Nicht versionieren:

- `.env`
- dynamische Uploads aus dem Adminbereich
- Datenbankdaten
- Backup-Archive
