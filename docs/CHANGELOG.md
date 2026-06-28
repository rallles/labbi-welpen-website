# Changelog

## Unreleased

### Dokumentation

- Vollstaendige Projektdokumentation unter `docs/` ergaenzt.
- README als Wiedereinstiegs- und Betriebsuebersicht erstellt.
- ADRs fuer Projektstruktur, Neo4j-Datenmodell, Static/Uploads und Admin-Security angelegt.

### Security und Hardening

- `.env`-Ladelogik in `cmd/main.go` vereinfacht: lokal optional `.env`, Docker/AWS ueber Environment.
- Config-Validierung fuer Neo4j- und Admin-Pflichtwerte ergaenzt.
- Logausgaben vermeiden Secret-Werte.
- Contact Rate Limiter raeumt alte IP-Keys periodisch auf.
- Nginx-Kommentar zu Zertifikaten fuer alle HTTPS-Servernamen ergaenzt.
- CSRF-Tokens fuer Admin-POST-Routen sind Single-Use via `Consume`.

### Tests

- CSRF-Tests finalisiert: kein Skip-Test, direkte Consume-Tests, Wiederverwendung schlaegt fehl.
- Contact-Mail-Header-Test robuster gegen CRLF-Injection gemacht.
- Redundanten Contact-Validation-Test entfernt.
- Puppy-Validation-Tests erweitert: Name, Notizen, Gewicht, Eltern-Normalisierung und Deduplizierung.
- Upload-Tests erweitert: leere Liste, Gesamtlimit, sichere Dateinamen.
- Config-Validation-Tests ergaenzt.
- Rate-Limiter-Cleanup-Test ergaenzt.

### Config und Betrieb

- `.env.example` nutzt Platzhalterwerte fuer Neo4j/Admin/SMTP.
- `.gitignore` ignoriert `.env` und `.env.*`, erlaubt aber `.env.example`.
- Docker Compose definiert Healthchecks fuer Neo4j und Web.

### Datenmodell

- Neo4j-Constraints fuer `Puppy.id`, `Dog.id` und `Contact.id`.
- Parent-Dogs `gandalf`, `anna`, `brina` werden beim Start geseedet.
- `Puppy`-Knoten pflegen `HAS_PARENT`-Relationships zu `Dog`.
- `UpdateMailStatus` meldet fehlende Contact-Knoten mit `ErrContactNotFound`.
- Admin Delete entfernt zugehoerige Dateien sicher aus `UPLOAD_DIR` und meldet Cleanup-Fehler als Warnung.

### Bekannte Einschraenkungen

- In dieser Arbeitsumgebung fehlen `go` und `gofmt`; Tests und Formatierung muessen in einer Umgebung mit Go 1.24.1 nachgeholt werden.
- SMTP ist optional; ohne vollstaendige SMTP-Konfiguration wird gespeichert, aber keine Kontaktmail versendet.
