// handlers/mistakes.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"app/mistakes/internal/database"
	"app/mistakes/internal/models"
)

// GetMistakes retrieves all mistakes from the database.
func GetMistakes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db := database.GetDB() 
	rows, err := db.Query("SELECT id, title, description, category FROM mistakes ORDER BY id ASC")
	if err != nil {
		log.Printf("Error querying mistakes: %v", err)
		http.Error(w, "Failed to retrieve mistakes", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Ensure rows are closed even if errors occur later

	mistakes := []models.Mistake{} // Initialize empty slice to hold results
	for rows.Next() {
		var m models.Mistake
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Category); err != nil {
			log.Printf("Error scanning mistake row: %v", err)
			http.Error(w, "Error processing data", http.StatusInternalServerError)
			return
		}
		mistakes = append(mistakes, m)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating mistake rows: %v", err)
		http.Error(w, "Error processing data", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(mistakes); err != nil {
		log.Printf("Error encoding mistakes to JSON: %v", err)

	}
}

// CreateMistake adds a new mistake to the database.
func CreateMistake(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var mistake models.Mistake

	if err := json.NewDecoder(r.Body).Decode(&mistake); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if mistake.Title == "" || mistake.Description == "" {
		http.Error(w, "Title and Description are required", http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	
	sqlStatement := `
	INSERT INTO mistakes (title, description, category)
	VALUES ($1, $2, $3)
	RETURNING id`

	var insertedID int
	err := db.QueryRow(sqlStatement, mistake.Title, mistake.Description, mistake.Category).Scan(&insertedID)
	if err != nil {
		log.Printf("Error inserting mistake into database: %v", err)
		// Check for specific DB errors if needed (e.g., unique constraint violation)
		if err == sql.ErrNoRows { // Should not happen with INSERT...RETURNING unless something is very wrong
			http.Error(w, "Failed to create mistake (no ID returned)", http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to create mistake", http.StatusInternalServerError)
		}
		return
	}

	mistake.ID = insertedID
	log.Printf("Successfully created mistake with ID: %d", insertedID)


	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(mistake); err != nil {
		log.Printf("Error encoding created mistake to JSON: %v", err)
	}
}