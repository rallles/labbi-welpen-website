# Setup

## Voraussetzungen

- Go 1.24.1, passend zu `go.mod`
- Docker mit Compose Plugin, wenn du Compose nutzt
- Git
- Neo4j lokal oder ueber Docker Compose

Pruefen:

```bash
go version
docker version
docker compose version
git status
```

## Environment

Im Projektroot:

```bash
cp .env.example .env
```

Die Vorlage ist fuer den direkten lokalen Start vorbereitet. Platzhalter-Passwoerter
bei Bedarf lokal aendern, aber keine echten Secrets committen.

Fuer lokalen Go-Start ohne Docker sind diese Werte typisch:

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

Beim direkten Start mit `go run ./cmd` wird aus `.env`
`NEO4J_URI=bolt://localhost:7687` verwendet. Beim Docker-Start ueberschreibt
`docker-compose.yml` diesen Wert im `web`-Service mit `bolt://neo4j:7687` und setzt
auch Serveradresse, Upload-, Static- und Template-Verzeichnisse auf Containerwerte.
Die Zugangsdaten werden weiterhin aus `.env` interpoliert.

## Optionales SMTP

Ohne SMTP-Konfiguration speichert die App Kontaktanfragen nur in Neo4j und versucht
keinen Mailversand. Fuer Benachrichtigungen muessen alle Werte gemeinsam gesetzt sein:

```env
SMTP_HOST=smtp.example.de
SMTP_PORT=587
SMTP_USER=sender@example.de
SMTP_PASSWORD=change_me_smtp_password
CONTACT_MAIL_TO=kontakt@example.de
```

Fehlt mindestens ein Wert, gilt SMTP als nicht konfiguriert. Die auskommentierten
Beispiele in `.env.example` sind keine Produktionswerte.

## Lokaler Start ohne Docker

Du brauchst eine erreichbare Neo4j-Instanz.

```bash
go mod download
go test ./...
go run ./cmd
```

Oeffnen:

- `http://localhost:8080/`
- `http://localhost:8080/healthz`
- `http://localhost:8080/admin`

## Neo4j lokal

Schneller Weg mit Docker nur fuer Neo4j:

```bash
docker run --rm --name labbi-neo4j \
  -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/change_me_strong_neo4j_password \
  neo4j:5.26.1-community
```

Dann `.env`:

```env
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=change_me_strong_neo4j_password
```

Beim App-Start werden Constraints erstellt und Parent-Dogs geseedet.

## Start mit Docker Compose

```bash
cp .env.example .env
docker compose config
docker compose build
docker compose up -d
docker compose ps
docker compose logs -f web nginx neo4j
```

Oeffnen:

- HTTP: `http://localhost/`
- Healthcheck: `http://localhost/healthz`

Auf einer produktiven Instanz laeuft der Zugriff ueber HTTPS und Nginx-Zertifikate.

## Wenn Go fehlt

Symptom:

```text
go: command not found
gofmt: command not found
```

Loesung: Go 1.24.1 installieren oder im passenden Devcontainer/Buildcontainer arbeiten. Ohne Toolchain koennen Tests und Formatierung nicht lokal ausgefuehrt werden.
