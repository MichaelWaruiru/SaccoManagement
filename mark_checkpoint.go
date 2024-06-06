package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func markCheckpointHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	if r.Method == http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checkpoints, err := getAllCheckpoints()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching checkpoints:", err)
		return
	}

	cars, err := getAllCars()
	if err != nil {
		http.Error(w, "Failed to fetch cars", http.StatusInternalServerError)
		log.Println("Error fetching cars:", err)
		return
	}

	routes, err := getRoutes()
	if err != nil {
		http.Error(w, "Failed to fetch routes", http.StatusInternalServerError)
		log.Println("Error fetching routes:", routes)
		return
	}

	// Data to pass to the template
	data := struct {
		Checkpoints []Checkpoint
		Cars        []Car
		Routes      []Route
	}{
		Checkpoints: checkpoints,
		Cars:        cars,
		Routes:      routes,
	}

	// Execute menu template
	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))
	err = menuTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	// Execute the mark_checkpoint template
	err = tmpl.ExecuteTemplate(w, "markCheckpoints", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func addMarkCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Println("Error parsing form:", err)
		return
	}

	// Extract form values
	routeID := r.Form.Get("routeSelect")
	checkpointID := r.Form.Get("checkpointSelect")
	carID := r.Form.Get("carSelect")
	timeStr := r.Form.Get("timePicker")
	dateStr := r.Form.Get("datePicker")

	// Validate form data
	if routeID == "" || checkpointID == "" || carID == "" || timeStr == "" || dateStr == "" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		log.Println("Invalid form data")
		return
	}

	log.Println("Checkpoint ID:", checkpointID)
	// log.Println("Car ID:", carID)
	// log.Println("Time:", timeStr)
	log.Println("Route ID: ", routeID)

	// Validate time format
	if timeStr == "" {
		http.Error(w, "Time is required", http.StatusBadRequest)
		log.Println("Time is empty")
		return
	}

	// Parse time string into time.Time format
	dateWithTime := fmt.Sprintf("%s %s", dateStr, timeStr)
	checkpointTime, err := time.Parse("2006-01-02 15:04", dateWithTime)
	if err != nil {
		http.Error(w, "Invalid time format", http.StatusBadRequest)
		log.Println("Error parsing time:", err)
		return
	}

	// Insert data into the marked_checkpoints table
	err = insertMarkedCheckpoint(routeID, checkpointID, carID, checkpointTime)
	if err != nil {
		http.Error(w, "Failed to insert marked checkpoint", http.StatusInternalServerError)
		log.Println("Error inserting marked checkpoint:", err)
		return
	}

	// Respond with success message
	// w.Write([]byte("Checkpoint marked successfully"))
	// Redirect the user to /mark-checkpoint
	http.Redirect(w, r, "/mark-checkpoint", http.StatusSeeOther)
}

// Function to insert marked checkpoint into the database
func insertMarkedCheckpoint(routeID, checkpointID, carID string, checkpointTime time.Time) error {
	query := "INSERT INTO marked_checkpoints (route_id, checkpoint_id, car_id, checkpoint_time, checkpoint_date) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, routeID, checkpointID, carID, checkpointTime, checkpointTime.Format("2006-01-02"))
	if err != nil {
		return err
	}

	// log.Printf("Inserting marked checkpoint: RouteID=%s, CheckpointID=%s, CarID=%s, Time=%s", routeID, checkpointID, carID, checkpointTime.Format("2006-01-02 15:04:05"))
	return nil
}

// Handler to fetch marked checkpoints data
func getMarkedCheckpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch data parameter from the URL query string
	date := r.FormValue("date")

	// Fetch marked checkpoints from the database or any other data source
	markedCheckpoints, err := getMarkedCheckpointsFromDB(date)
	if err != nil {
		http.Error(w, "Failed to fetch marked checkpoints", http.StatusInternalServerError)
		log.Println("Failed to fetch marked checkpoints from the DB", err)
		return
	}

	// Convert marked checkpoints to JSON
	jsonData, err := json.Marshal(markedCheckpoints)
	if err != nil {
		http.Error(w, "Failed to encode marked checkpoints data", http.StatusInternalServerError)
		log.Println("Error encoding JSON data:", err)
		return
	}

	// Set Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.Write(jsonData)
}

// Function to fetch marked checkpoints from the database
func getMarkedCheckpointsFromDB(date string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
        SELECT mc.checkpoint_id, mc.car_id, mc.checkpoint_time, mc.checkpoint_date, c.checkpoint_name, ca.number_plate
        FROM marked_checkpoints mc
        JOIN checkpoints c ON mc.checkpoint_id = c.id
        JOIN cars ca ON mc.car_id = ca.id
		WHERE mc.checkpoint_date = ?
    `, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markedCheckpoints []map[string]interface{}
	for rows.Next() {
		var checkpointID, carID string
		var checkpointTimeStr, checkpointDateStr, checkpointName, numberPlate sql.NullString
		if err := rows.Scan(&checkpointID, &carID, &checkpointTimeStr, &checkpointDateStr, &checkpointName, &numberPlate); err != nil {
			return nil, err
		}

		var checkpointTime time.Time
		if checkpointTimeStr.Valid {
			checkpointTime, err = time.Parse("2006-01-02 15:04:05", checkpointTimeStr.String)
			if err != nil {
				return nil, err
			}
		}

		var date string
		if checkpointDateStr.Valid {
			date = checkpointDateStr.String
		}

		markedCheckpoint := map[string]interface{}{
			"CheckpointID":   checkpointID,
			"CarID":          carID,
			"Time":           checkpointTime.Format("15:04"),
			"Date":           date,
			"CheckpointName": checkpointName.String,
			"NumberPlate":    numberPlate.String,
		}
		markedCheckpoints = append(markedCheckpoints, markedCheckpoint)
	}

	return markedCheckpoints, nil
}

func filterTripsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	carID := r.URL.Query().Get("carID")
	date := r.URL.Query().Get("date")

	trips, err := getTrips(carID, date)
	if err != nil {
		http.Error(w, "Failed to fetch trips", http.StatusInternalServerError)
		log.Println("Failed to fetch trips from the DB", err)
		return
	}

	jsonData, err := json.Marshal(trips)
	if err != nil {
		http.Error(w, "Failed to encode trips data", http.StatusInternalServerError)
		log.Println("Error encoding JSON data:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func getTrips(carID, date string) ([]map[string]interface{}, error) {
	query := `
        SELECT mc.checkpoint_date, ca.number_plate, COUNT(*) AS trip_count
        FROM marked_checkpoints mc
        JOIN cars ca ON mc.car_id = ca.id
        WHERE 1=1`

	var args []interface{}
	if carID != "" {
		query += " AND mc.car_id = ?"
		args = append(args, carID)
	}
	if date != "" {
		query += " AND mc.checkpoint_date = ?"
		args = append(args, date)
	}

	query += " GROUP BY mc.checkpoint_date, ca.number_plate ORDER BY mc.checkpoint_date, ca.number_plate"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []map[string]interface{}
	for rows.Next() {
		var checkpointDate, numberPlate string
		var tripCount int
		if err := rows.Scan(&checkpointDate, &numberPlate, &tripCount); err != nil {
			return nil, err
		}

		trip := map[string]interface{}{
			"Date":        checkpointDate,
			"NumberPlate": numberPlate,
			"Count":       tripCount,
		}
		trips = append(trips, trip)
	}

	return trips, nil
}
