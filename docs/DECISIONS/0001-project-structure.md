# ADR 0001: Projektstruktur Go/Docker/Neo4j

## Kontext

Die App soll eine kleine produktionsnahe Website mit Adminbereich, Datenbank und Uploads betreiben. Sie soll lokal entwickelbar und auf einer AWS-Instanz mit Docker Compose deploybar sein.

## Entscheidung

Das Projekt nutzt:

- Go `net/http` mit Einstieg in `cmd/main.go`
- fachliche Packages unter `internal/`
- serverseitige Go-Templates unter `internal/templates`
- statische Assets unter `static/`
- Neo4j als Datenbank
- Dockerfile fuer das Go-Binary und Assets
- Docker Compose fuer `web`, `neo4j`, `nginx`
- Nginx als Reverse Proxy und TLS-Endpunkt

## Konsequenzen

- Die App bleibt einfach und ohne grosses Framework.
- `internal/` schuetzt Anwendungscode vor externer Verwendung.
- Docker Compose bildet Produktion und lokale Integration gut ab.
- Templates und Static Assets werden ins Image kopiert; Uploads bleiben im Volume.
- Betrieb braucht Docker, Compose, `.env` und Zertifikate.

## Alternativen

- Fullstack-Framework: mehr Komfort, aber mehr Abhaengigkeiten.
- Separate Frontend-App: fuer dieses Projekt zu gross.
- SQLite/PostgreSQL statt Neo4j: einfacher fuer relationale Daten, aber das aktuelle Modell nutzt explizite Graph-Beziehungen zwischen Welpen und Elternhunden.
