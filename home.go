package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type Home struct {
	ID        int
	SaccoName string
	SaccoID   int
	Vehicles  int
	Route     string
	Manager   string
	Contact   string
}

type SaccoData struct {
	Sacco
	Vehicles int
	Route    []string
	Manager  string
	Contact  string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve all SACCO data including vehicles count
	saccoData, err := getAllSaccoData()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching SACCO data:", err)
		return
	}

	// Execute menu template
	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))

	// Execute menu template
	err = menuTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	// Render the home template with the retrieved SACCO data
	err = tmpl.ExecuteTemplate(w, "home", saccoData)
	if err != nil {
		http.Error(w, "Error rendering the template", http.StatusInternalServerError)
		log.Println("Error rendering the template:", err)
		return
	}
}

func getSaccoDataHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve SACCO data from the database
	saccoData, err := getAllSaccoData()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching SACCO data:", err)
		return
	}

	// Convert SACCO data to JSON format
	saccoDataJSON, err := json.Marshal(saccoData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error encoding SACCO data to JSON:", err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.Write(saccoDataJSON)
}

func getAllSaccoData() ([]SaccoData, error) {
	var saccoData []SaccoData

	// Fetch SACCOs
	saccos, err := getAllSaccos()
	if err != nil {
		return nil, err
	}

	// Iterate over SACCOs
	for _, sacco := range saccos {
		// Fetch vehicles count for the SACCO
		vehiclesCount, err := getVehiclesCountForSacco(sacco.ID)
		if err != nil {
			return nil, err
		}

		// Fetch routes for the SACCO
		routes, err := getRoutesForSacco(sacco.ID)
		if err != nil {
			return nil, err
		}

		// Fetch manager and contact for the SACCO
		manager, contact, err := getManagerAndContactForSacco(sacco.ID)
		if err != nil {
			return nil, err
		}

		// Create SaccoData object and append to the slice
		saccoData = append(saccoData, SaccoData{
			Sacco:    sacco,
			Vehicles: vehiclesCount,
			Route:    routes, // Assign the routes slice directly
			Manager:  manager,
			Contact:  contact,
		})
	}

	return saccoData, nil
}

// Get all vehicles count for a specific sacco
func getVehiclesCountForSacco(saccoID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cars WHERE sacco_id = ?", saccoID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get all routes as a slice for a specific sacco along with checkpoint names
func getRoutesForSacco(saccoID int) ([]string, error) {
	// Query to fetch checkpoint names for routes
	query := `
        SELECT c.checkpoint_name
        FROM route_checkpoints rc
        JOIN checkpoints c ON rc.checkpoint_id = c.id
        JOIN routes r ON rc.route_id = r.id
        WHERE r.sacco_id = ?
    `

	// Execute the query
	rows, err := db.Query(query, saccoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []string
	for rows.Next() {
		var checkpointName string
		if err := rows.Scan(&checkpointName); err != nil {
			return nil, err
		}
		// Append the checkpoint name to the checkpoints slice
		routes = append(routes, checkpointName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// Get managers and contacts for all saccos
func getManagerAndContactForSacco(saccoID int) (string, string, error) {
	var manager, contact sql.NullString
	err := db.QueryRow("SELECT manager, contact FROM saccos WHERE id = ?", saccoID).Scan(&manager, &contact)
	if err != nil {
		return "", "", err
	}

	// Check if manager and contact are valid
	var managerStr, contactStr string
	if manager.Valid {
		managerStr = manager.String
	}
	if contact.Valid {
		contactStr = contact.String
	}

	return managerStr, contactStr, nil
}
