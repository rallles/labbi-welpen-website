# Lokale und AWS-Konfiguration

## Prinzip

- `docker-compose.yml` ist die gemeinsame Basis.
- `docker-compose.local.yml` ergaenzt lokale Ports und mounted `nginx.local.conf`.
- `docker-compose.aws.yml` ergaenzt Produktiv-Ports, Restart-Policy und mounted `nginx.aws.conf`.
- `.env` ist pro Maschine unterschiedlich und wird nicht committet.
- Neo4j bekommt kein `env_file`, damit App-Variablen wie `NEO4J_URI` nicht als Neo4j-Settings interpretiert werden.

## Lokal Windows

`.env` im Projektordner aus `.env.example` erstellen und lokale Werte setzen:

```env
COMPOSE_FILE=docker-compose.yml;docker-compose.local.yml
NEO4J_USER=neo4j
NEO4J_PASSWORD=change_me_local_neo4j_password
ADMIN_USER=admin
ADMIN_PASSWORD=change_me_local_admin_password
SMTP_HOST=smtp.invalid
SMTP_PORT=587
SMTP_USER=sender@example.invalid
SMTP_PASSWORD=change_me_local_smtp_password
CONTACT_MAIL_TO=contact@example.invalid
```

Start:

```powershell
docker compose config
docker compose up -d --build
docker compose ps
```

Website:

```text
http://localhost:8081
```

Optionaler Neo4j Browser:

```text
http://localhost:7474
```

## AWS Linux

`.env` im Projektordner aus `.env.example` erstellen und echte Produktionswerte nur auf der Instanz setzen:

```env
COMPOSE_FILE=docker-compose.yml:docker-compose.aws.yml
NEO4J_USER=neo4j
NEO4J_PASSWORD=change_me_strong_production_neo4j_password
ADMIN_USER=admin
ADMIN_PASSWORD=change_me_strong_production_admin_password
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=sender@example.com
SMTP_PASSWORD=change_me_strong_production_smtp_password
CONTACT_MAIL_TO=kontakt@example.com
```

Start:

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

Auf AWS muss `/etc/letsencrypt` vorhanden sein. Der AWS-Override mounted es read-only in den Nginx-Container und published nur Nginx auf `80:80` und `443:443`; Neo4j-Ports werden dort nicht nach aussen freigegeben.

## Wichtig

In keinem Service steht `env_file: .env`. Die Werte werden in `docker-compose.yml` gezielt in `environment:` an die Container weitergegeben. Dadurch bekommt Neo4j keine App-Variable wie `NEO4J_URI` mehr und startet sauber.
