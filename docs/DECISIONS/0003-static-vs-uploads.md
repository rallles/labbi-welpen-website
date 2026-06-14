# ADR 0003: Trennung von /static und /uploads

## Kontext

Die Website nutzt feste Assets wie CSS, Icons und Websitebilder. Der Adminbereich erzeugt dynamische Uploads fuer Welpenbilder. Diese beiden Arten von Dateien haben unterschiedliche Lebenszyklen.

## Entscheidung

Es gibt eine klare Trennung:

- `/static/...` fuer versionierte Dateien aus `static/`
- `/uploads/...` fuer dynamische Admin-Uploads aus `UPLOAD_DIR`
- `static/images/generated/` fuer optimierte, versionierte Bildderivate

Docker Compose nutzt:

- bind mount `./static:/var/www/static:ro`
- Volume `uploads:/app/data/uploads`
- Volume `uploads:/var/www/uploads:ro`

## Konsequenzen

- Website-Assets koennen mit Git versioniert werden.
- Uploads ueberleben Container-Rebuilds im Volume.
- Backups muessen das Upload-Volume separat sichern.
- Fehlende Bilder in Arbeitsstaenden duerfen nicht automatisch zu Pfad- oder Template-Aenderungen fuehren.

## Alternativen

- Alle Bilder im Git: schlecht fuer dynamische Uploads und Repository-Groesse.
- Alle Bilder als Uploads: feste Website-Assets waeren schlechter reproduzierbar.
- Object Storage: langfristig moeglich, fuer aktuelle kleine AWS-Compose-Struktur aber mehr Komplexitaet.
