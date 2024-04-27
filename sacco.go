package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Sacco struct {
	ID        int
	SaccoName string
	Manager   string
	Contact   string
}

type CarWithTrips struct {
	Car
	Trips int
}

func saccoHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Retrieve saccos from the database
	saccos, err := getAllSaccos()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching saccos:", err)
		return
	}

	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))
	err = menuTemplate.Execute(w, saccos)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	data := struct {
		Saccos []Sacco
	}{
		Saccos: saccos,
	}

	err = tmpl.ExecuteTemplate(w, "sacco", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func getAllSaccos() ([]Sacco, error) {
	var saccos []Sacco

	rows, err := db.Query("SELECT * FROM saccos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sacco Sacco
		err := rows.Scan(&sacco.ID, &sacco.SaccoName, &sacco.Manager, &sacco.Contact)
		if err != nil {
			return nil, err
		}
		saccos = append(saccos, sacco)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return saccos, nil
}

func addSaccoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON data from the request body
	decoder := json.NewDecoder(r.Body)
	var requestData map[string]string
	err := decoder.Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	// Extract sacco name, manager, and contact from the request data
	saccoName, ok := requestData["saccoName"]
	if !ok || saccoName == "" {
		http.Error(w, "Sacco name cannot be empty", http.StatusBadRequest)
		log.Println("Sacco name is missing or empty")
		return
	}
	manager, ok := requestData["manager"]
	if !ok || manager == "" {
		http.Error(w, "Manager name cannot be empty", http.StatusBadRequest)
		log.Println("Manager name is missing or empty")
		return
	}
	contact, ok := requestData["contact"]
	if !ok || contact == "" {
		http.Error(w, "Contact cannot be empty", http.StatusBadRequest)
		log.Println("Contact is missing or empty")
		return
	}

	// Check if the sacco already exists in the system
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM saccos WHERE sacco_name = ?", saccoName).Scan(&count)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error checking if sacco exists:", err)
		return
	}
	if count > 0 {
		http.Error(w, "Sacco already exists", http.StatusBadRequest)
		log.Println("Sacco already exists")
		return
	}

	// Insert the new sacco into the database
	_, err = db.Exec("INSERT INTO saccos (sacco_name, manager, contact) VALUES (?, ?, ?)", saccoName, manager, contact)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error inserting sacco:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

// Retrieves SACCO details based on the provided saccoID
func getSaccoDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse saccoID from query parameters
	saccoIDStr := r.URL.Query().Get("saccoID")
	// if saccoIDStr == "" {
	// 	http.Error(w, "Missing SACCO ID", http.StatusBadRequest)
	// 	log.Println("Sacco ID cannot be empty")
	// 	return
	// }

	saccoID, err := strconv.Atoi(saccoIDStr)
	if err != nil {
		http.Error(w, "Invalid SACCO ID", http.StatusBadRequest)
		log.Println("Invalid sacco ID:", err)
		return
	}

	// Retrieve SACCO details from your data source using the saccoID
	sacco, err := getSaccoByID(saccoID)
	if err != nil {
		http.Error(w, "Failed to retrieve SACCO details", http.StatusInternalServerError)
		log.Println("Error getting sacco details")
		return
	}

	// Convert SACCO details to JSON
	saccoJSON, err := json.Marshal(sacco)
	if err != nil {
		http.Error(w, "Failed to marshal SACCO details to JSON", http.StatusInternalServerError)
		return
	}

	// Set content type and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(saccoJSON)
}

func getSaccoByID(saccoID int) (Sacco, error) {
	var sacco Sacco

	// Query the database to get the SACCO by ID
	err := db.QueryRow("SELECT id, sacco_name, manager, contact FROM saccos WHERE id = ?", saccoID).
		Scan(&sacco.ID, &sacco.SaccoName, &sacco.Manager, &sacco.Contact)
	if err != nil {
		return Sacco{}, err
	}

	return sacco, nil
}

func editSaccoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON data from the request body
	decoder := json.NewDecoder(r.Body)
	var requestData map[string]string
	err := decoder.Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	// Extract SACCO ID and updated details from the request data
	saccoID, ok := requestData["saccoID"]
	if !ok || saccoID == "" {
		http.Error(w, "SACCO ID is missing", http.StatusBadRequest)
		log.Println("SACCO ID is missing")
		return
	}

	saccoName, ok := requestData["saccoName"]
	if !ok || saccoName == "" {
		http.Error(w, "SACCO name cannot be empty", http.StatusBadRequest)
		log.Println("SACCO name is missing or empty")
		return
	}

	manager, ok := requestData["manager"]
	if !ok || manager == "" {
		http.Error(w, "Manager name cannot be empty", http.StatusBadRequest)
		log.Println("Manager name is missing or empty")
		return
	}

	contact, ok := requestData["contact"]
	if !ok || contact == "" {
		http.Error(w, "Contact cannot be empty", http.StatusBadRequest)
		log.Println("Contact is missing or empty")
		return
	}

	// Check if the SACCO exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM saccos WHERE id = ?", saccoID).Scan(&count)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error checking if SACCO exists:", err)
		return
	}
	if count == 0 {
		http.Error(w, "SACCO not found", http.StatusNotFound)
		log.Println("SACCO not found")
		return
	}

	// Update the SACCO in the database
	_, err = db.Exec("UPDATE saccos SET sacco_name = ?, manager = ?, contact = ? WHERE id = ?", saccoName, manager, contact, saccoID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error updating SACCO:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

func deleteSaccoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse sacco ID from query parameters
	saccoIDStr := r.URL.Query().Get("saccoID")
	if saccoIDStr == "" {
		http.Error(w, "Missing sacco ID", http.StatusBadRequest)
		log.Println("Sacco ID cannot be empty")
		return
	}

	saccoID, err := strconv.Atoi(saccoIDStr)
	if err != nil {
		http.Error(w, "Invalid sacco ID", http.StatusBadRequest)
		log.Println("Invalid sacco ID:", err)
		return
	}

	// Check if the sacco exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM saccos WHERE id = ?", saccoID).Scan(&count)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error checking if sacco exists:", err)
		return
	}

	if count == 0 {
		http.Error(w, "Sacco not found", http.StatusNotFound)
		log.Println("Sacco not found")
		return
	}

	// Delete sacco from the database
	_, err = db.Exec("DELETE FROM saccos WHERE id = ?", saccoID)
	if err != nil {
		http.Error(w, "Error deleting sacco", http.StatusInternalServerError)
		log.Println("Failed to delete sacco from DB:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
}

// Define a new handler function to fetch cars and drivers data by SACCO ID
func getCarsAndDriversAndRoutesHandler(w http.ResponseWriter, r *http.Request) {
	// Get the SACCO ID from the query parameters
	saccoID := r.URL.Query().Get("saccoID")

	// Debugging: Print the SACCO ID
	// fmt.Println("SACCO ID:", saccoID)

	// Fetch cars and drivers data for the given SACCO ID
	cars, err := getCarsBySaccoID(saccoID)
	if err != nil {
		http.Error(w, "Failed to fetch cars data", http.StatusInternalServerError)
		fmt.Println("Error fetching cars data:", err)
		return
	}
	// fmt.Println("Retrieved cars:", cars)

	drivers, err := getDriversBySaccoID(saccoID)
	if err != nil {
		http.Error(w, "Failed to fetch drivers data", http.StatusInternalServerError)
		fmt.Println("Error fetching drivers data:", err)
		return
	}
	// fmt.Println("Retrieved drivers:", drivers)

	routes, err := getRoutesBySaccoID(saccoID)
	if err != nil {
		http.Error(w, "|Failed to fetch routes data", http.StatusInternalServerError)
		fmt.Println("Error fetching routes data:", err)
		return
	}
	// fmt.Println("Retrieved routes:", routes)

	// Fetch trips data for all cars collectively
	tripsData := make(map[int]int)
	// Fetch the date from marked_checkpoints table
	dateFromMarkedCheckpoints, err := getDateFromMarkedCheckpoints()
	if err != nil {
		http.Error(w, "Failed to fetch date from marked_checkpoints", http.StatusInternalServerError)
		fmt.Println("Error fetching date from marked_checkpoints:", err)
		return
	}
	fmt.Println("Date from checkpoints: ", dateFromMarkedCheckpoints)

	// Get trips data for the retrieved date
	trips, err := getTripsDataForDay(dateFromMarkedCheckpoints)
	if err != nil {
		http.Error(w, "Failed to fetch trips data", http.StatusInternalServerError)
		fmt.Println("Error fetching trips data:", err)
		return
	}
	fmt.Println("Trips: ", trips)

	// Convert the keys from string to int
	for carIDStr, tripCount := range trips {
		carID, err := strconv.Atoi(carIDStr)
		if err != nil {
			fmt.Println("Error converting car ID to int:", err)
			continue
		}
		tripsData[carID] = tripCount
	}

	// Construct a response containing cars, drivers, routes, and trips data
	responseData := struct {
		Cars      []Car
		Drivers   []Driver
		Routes    []Route
		TripsData map[int]int
	}{
		Cars:      cars,
		Drivers:   drivers,
		Routes:    routes,
		TripsData: tripsData,
	}

	// fmt.Println(cars)

	// Convert the response data to JSON
	jsonResponse, err := json.Marshal(responseData)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		fmt.Println("Error marshaling response:", err)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(jsonResponse)
}

func getCarsBySaccoID(saccoID string) ([]Car, error) {
	var cars []Car

	// Placeholder query logic to retrieve cars for the specified SACCO ID
	rows, err := db.Query("SELECT id, number_plate, make, model FROM cars WHERE sacco_id = ?", saccoID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var car Car
		err := rows.Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model)
		if err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func getDriversBySaccoID(saccoID string) ([]Driver, error) {
	var drivers []Driver

	// Placeholder query logic to retrieve drivers for the specified SACCO ID
	rows, err := db.Query("SELECT d.id, d.name, d.id_number, d.contact FROM drivers d WHERE d.sacco_id = ?", saccoID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var driver Driver
		err := rows.Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact)
		if err != nil {
			return nil, err
		}
		drivers = append(drivers, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drivers, nil
}

func getRoutesBySaccoID(saccoID string) ([]Route, error) {
	var routes []Route

	// Query to join routes with checkpoints to get checkpoint names
	query := `
		SELECT r.id, r.sacco_id, GROUP_CONCAT(r.checkpoint_id) as checkpoint_ids, GROUP_CONCAT(c.checkpoint_name) as checkpoint_names
		FROM routes r
		JOIN checkpoints c ON FIND_IN_SET(c.id, r.checkpoint_id)
		WHERE r.sacco_id = ?
		GROUP BY r.id
	`

	rows, err := db.Query(query, saccoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route Route
		var checkpointIDsStr, checkpointNamesStr string

		err := rows.Scan(&route.ID, &route.SaccoID, &checkpointIDsStr, &checkpointNamesStr)
		if err != nil {
			return nil, err
		}

		checkpointIDs := strings.Split(checkpointIDsStr, ",")
		checkpointNames := strings.Split(checkpointNamesStr, ",")

		for _, idStr := range checkpointIDs {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, err
			}
			route.CheckpointIDs = append(route.CheckpointIDs, id)
		}

		route.Checkpoints = checkpointNames

		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// Function to fetch trip data for each car on a specific day
func getTripsDataForDay(day time.Time) (map[string]int, error) {
	// Define the start and end time of the specified day
	startTime := day.Truncate(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	// Query marked checkpoints for the specified day
	query := `
		SELECT car_id, COUNT(DISTINCT DATE(checkpoint_time)) AS trip_count
		FROM marked_checkpoints
		WHERE checkpoint_time >= ? AND checkpoint_time < ?
		GROUP BY car_id
	`

	rows, err := db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to store trip count for each car
	tripCounts := make(map[string]int)

	// Iterate through the results and count trips for each car
	for rows.Next() {
		var carID string
		var count int
		if err := rows.Scan(&carID, &count); err != nil {
			return nil, err
		}
		tripCounts[carID] = count
		fmt.Println("Trip counts: ", tripCounts)
	}

	return tripCounts, nil
}

func getDateFromMarkedCheckpoints() (time.Time, error) {
	var dateStr string

	// Query to fetch the date from marked_checkpoints table
	query := "SELECT DISTINCT DATE(checkpoint_time) FROM marked_checkpoints ORDER BY DATE(checkpoint_time) DESC LIMIT 1"
	err := db.QueryRow(query).Scan(&dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// Parse the date string into a time.Time value
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
