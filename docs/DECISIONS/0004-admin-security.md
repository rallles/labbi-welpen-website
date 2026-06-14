# ADR 0004: Admin-Security mit Basic Auth, CSRF und Validierung

## Kontext

Der Adminbereich ist klein und fuer wenige berechtigte Personen gedacht. Er muss Welpen anlegen, bearbeiten, loeschen und Bilder hochladen koennen.

## Entscheidung

Admin-Security besteht aus:

- HTTP Basic Auth fuer alle `/admin`-Routen
- Credentials aus `ADMIN_USER` und `ADMIN_PASSWORD`
- constant-time Vergleich ueber SHA-256 Hashes
- CSRF-Tokens fuer Admin-POST-Routen
- Single-Use-CSRF via `Consume`
- serverseitiger Formularvalidierung
- Upload-Validierung fuer JPEG/PNG und Groessenlimits

## Konsequenzen

- Einfacher Betrieb ohne Session-Store.
- Browser verwaltet Basic-Auth-Credentials.
- CSRF-Schutz verhindert Wiederverwendung erfolgreicher POST-Tokens.
- Fuer komplexere Benutzer-/Rollenmodelle reicht diese Loesung nicht.

## Alternativen

- Session-Login mit Cookies: flexibler, aber mehr Code und Betrieb.
- OAuth/OIDC: fuer dieses kleine Projekt ueberdimensioniert.
- Nur Basic Auth ohne CSRF: fuer Formular-POSTs schwaecher.
