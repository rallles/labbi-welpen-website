# TODO

Diese Liste ist aus dem aktuellen Code- und Betriebsstand abgeleitet. Keine Fantasie-Features, sondern Punkte, die beim Weiterarbeiten wirklich relevant sind.

## Kritisch

- Produktions-`.env` auf der AWS-Instanz pruefen: Pflichtwerte gesetzt, starke Passwoerter, keine Secrets im Git.
- Backup von `neo4j_data` und `uploads` gemeinsam in einen isolierten Testordner erstellen,
  in einem frischen Compose-Projekt wiederherstellen und Daten plus Bildreferenzen pruefen.
  Erst danach den getesteten Ablauf als produktiven Runbook-Befehl uebernehmen.
- TLS-Zertifikate fuer alle vier Nginx-Servernamen pruefen oder Multi-Domain-Zertifikat einrichten.

## Hoch

- `scripts/smoke-local.sh` nach lokalen Deployments konsequent ausfuehren; AWS-Endpunkte
  nach dem dokumentierten, rein lesenden Ablauf pruefen.

## Mittel

- Admin-Bildverwaltung fuer bestehende Welpen planen, falls Bilder nachtraeglich geaendert oder geloescht werden sollen.
- Zwei UUID-benannte JPGs in `static/images/` pruefen, die aktuell nirgends referenziert werden:
  - `static/images/0ef00736-349c-41f6-bdd1-407ba2cb05f1_1751210248413534949.jpg`
  - `static/images/986fa9f4-26c0-485f-881e-b82520448925_1751219019634248891.jpg`
  Vermutlich alte Upload-Artefakte (UUID-Namensschema wie bei Admin-Uploads). Pruefen, ob sie noch gebraucht werden;
  ggf. nach `/uploads` migrieren oder entfernen. Nicht voreilig loeschen.

## Optional

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
