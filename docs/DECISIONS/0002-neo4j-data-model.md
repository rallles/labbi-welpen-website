# ADR 0002: Neo4j-Datenmodell mit Puppy, Dog, Contact

## Kontext

Die Website braucht dynamische Welpendaten, Elternhunde und Kontaktanfragen. Elternbeziehungen sollen explizit auswertbar sein.

## Entscheidung

Neo4j speichert:

- `(:Puppy)` fuer Welpen
- `(:Dog)` fuer Elternhunde
- `(:Contact)` fuer Kontaktanfragen
- `(:Puppy)-[:HAS_PARENT]->(:Dog)` fuer Elternbeziehungen

Zusatz:

- `Puppy.id` ist eindeutig.
- `Dog.id` ist eindeutig.
- Parent-Dogs `gandalf`, `anna`, `brina` werden beim App-Start geseedet.
- `Puppy.eltern` bleibt zusaetzlich als normalisierte Liste gespeichert, waehrend `HAS_PARENT` die Graph-Beziehung abbildet.

## Konsequenzen

- Elternbeziehungen koennen direkt per Cypher abgefragt werden.
- Seed-Daten sind im Code nachvollziehbar.
- Beim Update muessen Beziehungen ersetzt werden.
- Contact-Daten sind im selben Neo4j-System, aber aktuell ohne eigene Adminverwaltung.

## Alternativen

- Eltern nur als Stringliste auf `Puppy`: einfacher, aber weniger Graph-Nutzen.
- Separate relationale Datenbank: vertrauter, aber nicht die aktuelle Projektentscheidung.
- Contacts extern per Mail only: weniger Datenhaltung, aber Kontaktanfragen waeren bei Mailproblemen verloren.
