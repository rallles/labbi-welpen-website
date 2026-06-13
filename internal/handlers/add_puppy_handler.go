package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"labbi-app/internal/config"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// AddPuppyHandler verarbeitet das Admin-Formular (POST) und speichert den neuen Welpen in Neo4j.
func AddPuppyHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext, cfg config.Config) {
	// Nur POST zulassen
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	// Multipart-Form parsen (max. 20 MB im RAM)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("Fehler beim Lesen des Formulars: %v", err)
		http.Error(w, "Fehler beim Lesen des Formulars", http.StatusBadRequest)
		return
	}

	// Formularwerte auslesen
	name := r.FormValue("name")
	birthdate := r.FormValue("geburtsdatum")
	gender := r.FormValue("geschlecht")
	color := r.FormValue("farbe")
	weight := r.FormValue("gewicht")
	character := r.FormValue("charakter")
	vaccinated := r.FormValue("geimpft") == "true"
	chipped := r.FormValue("gechippt") == "true"
	dewormed := r.FormValue("entwurmung") == "true"
	notes := r.FormValue("notizen")

	// Eltern (Checkboxen können mehrfach übergeben werden)
	parents := r.Form["eltern"]

	// Bilder verarbeiten
	files := r.MultipartForm.File["images"]
	imagePaths, err := saveUploadedImages(files, cfg.UploadDir)
	if err != nil {
		log.Printf("Fehler beim Speichern der Bilder: %v", err)
		http.Error(w, "Fehler beim Speichern der Bilder", http.StatusInternalServerError)
		return
	}

	// Neo4j-Verbindung und Insert
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{})
	defer session.Close(context.Background())

	puppyID := uuid.NewString()
	// Welpenknoten erstellen
	_, err = session.Run(context.Background(),
		`CREATE (p:Puppy {
		    id: $id,
		    name: $name,
		    geburtsdatum: date($birthdate),
		    geschlecht: $gender,
		    farbe: $color,
		    gewicht: toFloat($weight),
		    charakter: $character,
		    geimpft: $vaccinated,
		    gechippt: $chipped,
		    entwurmt: $dewormed,
		    eltern: $parents,
		    notizen: $notes,
		    bilder: $images
		})`, map[string]interface{}{
			"id":         puppyID,
			"name":       name,
			"birthdate":  birthdate,
			"gender":     gender,
			"color":      color,
			"weight":     weight,
			"character":  character,
			"vaccinated": vaccinated,
			"chipped":    chipped,
			"dewormed":   dewormed,
			"parents":    parents,
			"notes":      notes,
			"images":     imagePaths,
		})

	// Elternbeziehungen anlegen, falls vorhanden
	if len(parents) > 0 {
		for _, parent := range parents {
			_, err = session.Run(context.Background(),
				`MATCH (p:Puppy {id: $puppyID}), (parent:Puppy {id: $parentID})
				CREATE (p)-[:HAS_PARENT]->(parent)`,
				map[string]interface{}{
					"puppyID":  puppyID,
					"parentID": parent,
				})
		}
	}
	// Fehler beim Speichern in Neo4j
	if err != nil {
		log.Printf("Fehler beim Speichern in Neo4j: %v", err)
		http.Error(w, "Fehler beim Speichern", http.StatusInternalServerError)
		return
	}

	// Bei Erfolg zurück zum Dashboard
	http.Redirect(w, r, "/admin?success=true", http.StatusSeeOther)
}

// saveUploadedImages speichert alle hochgeladenen Dateien und liefert ihre Pfade zurück
func saveUploadedImages(files []*multipart.FileHeader, uploadDir string) ([]string, error) {
	log.Printf("Upload-Verzeichnis: %s", uploadDir)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("FEHLER: Upload-Verzeichnis konnte nicht angelegt werden: %v", err)
		return nil, err
	}

	log.Printf("Anzahl hochgeladener Dateien: %d", len(files))

	var paths []string
	for idx, fh := range files {
		log.Printf("Verarbeite Datei %d: Originalname: %s", idx+1, fh.Filename)

		file, err := fh.Open()
		if err != nil {
			log.Printf("FEHLER: Datei %s konnte nicht geöffnet werden: %v", fh.Filename, err)
			return nil, err
		}

		ext := filepath.Ext(fh.Filename)
		name := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixNano(), ext)
		target := filepath.Join(uploadDir, name)
		log.Printf("Speichere nach: %s", target)

		out, err := os.Create(target)
		if err != nil {
			log.Printf("FEHLER: Datei konnte nicht angelegt werden: %v", err)
			file.Close()
			return nil, err
		}

		written, err := io.Copy(out, file)
		if err != nil {
			log.Printf("FEHLER: Datei %s konnte nicht gespeichert werden: %v", name, err)
			out.Close()
			file.Close()
			return nil, err
		}
		log.Printf("Datei %s gespeichert (%d Bytes)", name, written)

		paths = append(paths, name)

		// WICHTIG: Dateien gleich schließen (nicht mit defer im Loop)
		out.Close()
		file.Close()
	}
	return paths, nil
}
