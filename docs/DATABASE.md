# Datenbank

## Neo4j Setup

Die App nutzt Neo4j ueber `github.com/neo4j/neo4j-go-driver/v5`.

Config:

- `NEO4J_URI`
- `NEO4J_USER`
- `NEO4J_PASSWORD`

Beim Start:

1. Driver wird erstellt.
2. Constraints werden angelegt.
3. Parent-Dogs werden geseedet.

## Constraints

```cypher
CREATE CONSTRAINT puppy_id IF NOT EXISTS
FOR (p:Puppy) REQUIRE p.id IS UNIQUE;

CREATE CONSTRAINT dog_id IF NOT EXISTS
FOR (d:Dog) REQUIRE d.id IS UNIQUE;

CREATE CONSTRAINT contact_id IF NOT EXISTS
FOR (c:Contact) REQUIRE c.id IS UNIQUE;
```

## Seed-Daten

Parent-Dogs aus `internal/models/models.go`:

| ID | Name | Geschlecht | Rolle |
|---|---|---|---|
| `gandalf` | Gandalf | `männlich` | `parent` |
| `anna` | Anna | `weiblich` | `parent` |
| `brina` | Brina | `weiblich` | `parent` |

Seed passiert per `MERGE (d:Dog {id: $id})`.

## Puppy

Label: `:Puppy`

Properties:

| Property | Typ/Format | Quelle |
|---|---|---|
| `id` | UUID String | Server |
| `name` | String, max 80 | Adminformular |
| `geburtsdatum` | Neo4j `date` | Adminformular `YYYY-MM-DD` |
| `geschlecht` | `männlich` oder `weiblich` | Adminformular |
| `farbe` | erlaubte Fellfarbe | Adminformular |
| `gewicht` | Float, > 0 und <= 80 | Adminformular |
| `charakter` | String, max 1000 | Adminformular |
| `geimpft` | Bool | Adminformular |
| `gechippt` | Bool | Adminformular |
| `entwurmt` | Bool | Adminformular |
| `eltern` | Liste normalisierter Parent-IDs | Adminformular |
| `notizen` | String, max 2000 | Adminformular |
| `bilder` | Liste relativer Pfade `/uploads/...` | Upload-Flow |

Relationship:

```cypher
(:Puppy)-[:HAS_PARENT]->(:Dog)
```

Beim Update werden alte `HAS_PARENT`-Beziehungen geloescht und neu gesetzt.

Beim Loeschen liefert `PuppyRepository.Delete` `ErrPuppyNotFound`, wenn kein passender
Knoten geloescht wurde. Der Handler loescht zuerst den Neo4j-Knoten und versucht erst
danach, die zuvor geladenen Upload-Dateien zu entfernen. Fehler beim Upload-Cleanup
werden geloggt und als Warnung angezeigt; der erfolgreiche DB-Delete wird nicht rueckgaengig gemacht.

`PuppyRepository.List` liest diese Knoten fuer die oeffentliche Route `/puppies` und die
Admin-Tabelle `/admin/puppies`. Die feste Galerie in `puppies.html` ist kein Datenbankinhalt.

## Dog

Label: `:Dog`

Properties:

- `id`
- `name`
- `gender`
- `role`

Aktuell dienen Dogs als Elternhunde fuer Welpen.

## Contact

Label: `:Contact`

Properties:

| Property | Beschreibung |
|---|---|
| `id` | UUID String |
| `name` | Name aus Kontaktformular |
| `email` | E-Mail aus Kontaktformular |
| `phone` | optionale Telefonnummer |
| `message` | Nachricht |
| `createdAt` | Neo4j `datetime` |
| `mailSent` | Bool |
| `mailError` | sanitizte Fehlerbeschreibung |

## Debug Queries

Alle Welpen:

```cypher
MATCH (p:Puppy)
RETURN p
ORDER BY p.geburtsdatum DESC, p.name ASC;
```

Welpen mit Eltern:

```cypher
MATCH (p:Puppy)
OPTIONAL MATCH (p)-[:HAS_PARENT]->(d:Dog)
RETURN p.id, p.name, collect(d.id) AS parents
ORDER BY p.name;
```

Parent-Dogs:

```cypher
MATCH (d:Dog)
RETURN d
ORDER BY d.id;
```

Kontaktanfragen:

```cypher
MATCH (c:Contact)
RETURN c
ORDER BY c.createdAt DESC
LIMIT 20;
```

Constraints:

```cypher
SHOW CONSTRAINTS;
```

Anzahl Knoten:

```cypher
MATCH (n)
RETURN labels(n) AS labels, count(*) AS count
ORDER BY labels;
```
