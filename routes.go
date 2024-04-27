package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Route struct {
	ID             int
	SaccoID        int
	CheckpointIDs  []int // CheckpointIDs is a slice of checkpoint IDs
	SaccoName      string
	Checkpoints    []string
	CheckpointName []string
}

type RouteData struct {
	Saccos      []Sacco
	Checkpoints []Checkpoint
	Routes      []Route
}

func routesHandler(w http.ResponseWriter, r *http.Request) {
	// Checks if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	saccos, err := getAllSaccos()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching saccos:", err)
		return
	}

	checkpoints, err := getAllCheckpoints()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching checkpoints:", err)
		return
	}

	routes, err := getRoutes()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching routes:", err)
		return
	}

	data := RouteData{
		Saccos:      saccos,
		Checkpoints: checkpoints,
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

	// Execute routes template
	err = tmpl.ExecuteTemplate(w, "routes", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing routes template:", err)
		return
	}
}

func createRouteHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Println("Error parsing form data:", err)
		return
	}

	// Extract SACCO ID from form data
	saccoID := r.FormValue("sacco")
	if saccoID == "" {
		http.Error(w, "Sacco ID cannot be empty", http.StatusBadRequest)
		return
	}

	// Extract checkpoints from form data
	checkpoints := r.Form["checkpoints[]"]

	// Join checkpoint IDs into a comma-separated string
	checkpointIDsStr := strings.Join(checkpoints, ",")

	// Insert route into routes table
	result, err := db.Exec("INSERT INTO routes (sacco_id, checkpoint_id) VALUES (?, ?)", saccoID, checkpointIDsStr)
	if err != nil {
		http.Error(w, "Error inserting route in the database", http.StatusInternalServerError)
		log.Println("Error inserting route:", err)
		return
	}

	// Get the ID of the inserted route
	routeID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error getting route ID", http.StatusInternalServerError)
		log.Println("Error getting route ID:", err)
		return
	}

	// Insert route checkpoints into route_checkpoints table
	for _, checkpointID := range checkpoints {
		// Convert checkpointID to int
		checkpointIDInt, err := strconv.Atoi(checkpointID)
		if err != nil {
			http.Error(w, "Invalid checkpoint ID", http.StatusBadRequest)
			log.Println("Invalid checkpoint ID:", err)
			return
		}

		// Insert route checkpoint into route_checkpoints table
		_, err = db.Exec("INSERT INTO route_checkpoints (route_id, checkpoint_id) VALUES (?, ?)", routeID, checkpointIDInt)
		if err != nil {
			http.Error(w, "Error inserting route checkpoint in the database", http.StatusInternalServerError)
			log.Println("Error inserting route checkpoint:", err)
			return
		}
	}

	// Redirect to the routes page
	http.Redirect(w, r, "/routes", http.StatusSeeOther)
}

func getRoutesHandler(w http.ResponseWriter, r *http.Request) {
	// Call the getRoutes function to fetch routes data from the database
	routes, err := getRoutes()
	if err != nil {
		http.Error(w, "Failed to fetch routes", http.StatusInternalServerError)
		log.Println("Error fetching routes:", err)
		return
	}

	// Convert routes data to the format expected by the template
	var formattedRoutes []map[string]interface{}
	for _, route := range routes {
		// Fetch SACCO name based on ID
		saccoName, err := getSaccoName(route.SaccoID)
		if err != nil {
			http.Error(w, "Failed to fetch SACCO name", http.StatusInternalServerError)
			log.Println("Error fetching SACCO name:", err)
			return
		}

		// Fetch checkpoint names based on IDs
		var checkpointNames []string
		for _, checkpointID := range route.CheckpointIDs {
			checkpointName, err := getCheckpointName(checkpointID)
			if err != nil {
				http.Error(w, "Failed to fetch checkpoint name", http.StatusInternalServerError)
				log.Println("Error fetching checkpoint name:", err)
				return
			}
			checkpointNames = append(checkpointNames, checkpointName)
		}

		// Create formatted route map
		formattedRoute := map[string]interface{}{
			"sacco_name":  saccoName,
			"checkpoints": checkpointNames,
			"route_id":    route.ID,
			"sacco_id":    route.SaccoID,
		}
		formattedRoutes = append(formattedRoutes, formattedRoute)
	}

	// Log the formattedRoutes for debugging
	// log.Println("Formatted Routes:", formattedRoutes)

	// Convert the formatted routes data to JSON format
	jsonData, err := json.Marshal(formattedRoutes)
	if err != nil {
		http.Error(w, "Failed to marshal routes data", http.StatusInternalServerError)
		log.Println("Error marshaling routes data:", err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.Write(jsonData)
}

func getCheckpointName(checkpointID int) (string, error) {
	var checkpointName string
	err := db.QueryRow("SELECT checkpoint_name FROM checkpoints WHERE id = ?", checkpointID).Scan(&checkpointName)
	if err != nil {
		return "", err
	}
	return checkpointName, nil
}

func getRoutes() ([]Route, error) {
	// Query to fetch routes with their checkpoints
	rows, err := db.Query(`
        SELECT r.id, r.sacco_id, GROUP_CONCAT(rc.checkpoint_id) AS checkpoint_ids
        FROM routes r 
        INNER JOIN route_checkpoints rc ON r.id = rc.route_id
        GROUP BY r.id, r.sacco_id;
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var routeID, saccoID int
		var checkpointIDs string
		if err := rows.Scan(&routeID, &saccoID, &checkpointIDs); err != nil {
			return nil, err
		}
		// fmt.Println("Route ID:", routeID)
		// fmt.Println("Sacco ID:", saccoID)
		// fmt.Println("Checkpoint IDs:", checkpointIDs) // Print out the fetched checkpoint IDs
		// Convert checkpointIDs string to []int
		checkpointIDsArr := make([]int, 0)
		ids := strings.Split(checkpointIDs, ",")
		for _, id := range ids {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				return nil, err
			}
			checkpointIDsArr = append(checkpointIDsArr, idInt)
		}
		// fmt.Println("Parsed Checkpoint IDs:", checkpointIDsArr) // Print out the parsed checkpoint IDs
		routes = append(routes, Route{
			ID:            routeID,
			SaccoID:       saccoID,
			CheckpointIDs: checkpointIDsArr,
		})
	}

	return routes, nil
}

func getSaccoName(saccoID int) (string, error) {
	var saccoName string
	err := db.QueryRow("SELECT sacco_name FROM saccos WHERE id = ?", saccoID).Scan(&saccoName)
	if err != nil {
		return "", err
	}
	return saccoName, nil
}

func editRouteHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Println("Error parsing form data:", err)
		return
	}

	// Extract form values
	editRouteID := r.FormValue("editRouteID")
	editSaccoID := r.FormValue("editSaccoSelect")
	editCheckpoints := r.Form["editCheckpointsSelect[]"]

	// Join the editCheckpoints slice into a comma-separated string
	checkpointsStr := ""
	if len(editCheckpoints) > 0 {
		checkpointsStr = strings.Join(editCheckpoints, ",")
	}

	fmt.Println("Route ID:", editRouteID)
	fmt.Println("Sacco ID:", editSaccoID)
	fmt.Println("Checkpoints:", checkpointsStr)

	// Update route in the database
	_, err = db.Exec("UPDATE routes SET sacco_id = ?, checkpoint_id = ? WHERE id = ?", editSaccoID, checkpointsStr, editRouteID)
	if err != nil {
		http.Error(w, "Error updating route in the database", http.StatusInternalServerError)
		log.Println("Error updating route:", err)
		return
	}

	// Delete existing route checkpoints
	_, err = db.Exec("DELETE FROM route_checkpoints WHERE route_id = ?", editRouteID)
	if err != nil {
		http.Error(w, "Error deleting existing route checkpoints", http.StatusInternalServerError)
		log.Println("Error deleting existing route checkpoints:", err)
		return
	}

	// Insert updated route checkpoints
	for _, checkpointID := range editCheckpoints {
		// Convert checkpointID to int
		checkpointIDInt, err := strconv.Atoi(checkpointID)
		if err != nil {
			http.Error(w, "Invalid checkpoint ID", http.StatusBadRequest)
			log.Println("Invalid checkpoint ID:", err)
			return
		}

		// Insert route checkpoint into route_checkpoints table
		_, err = db.Exec("INSERT INTO route_checkpoints (route_id, checkpoint_id) VALUES (?, ?)", editRouteID, checkpointIDInt)
		if err != nil {
			http.Error(w, "Error inserting route checkpoint in the database", http.StatusInternalServerError)
			log.Println("Error inserting route checkpoint:", err)
			return
		}
	}

	// Handle successful update (redirect or respond with a success message)
	http.Redirect(w, r, "/routes", http.StatusFound)
}

func getRouteDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract route ID from query parameter
	routeID := r.URL.Query().Get("id")
	if routeID == "" {
		http.Error(w, "Route ID is required", http.StatusBadRequest)
		return
	}

	// Fetch route details from the data source based on the route ID
	routeDetails, err := getRouteDetailsFromDB(routeID)
	if err != nil {
		http.Error(w, "Failed to fetch route details", http.StatusInternalServerError)
		log.Println("Error fetching route details:", err)
		return
	}
	// fmt.Println("Route details from DB:", routeDetails)

	// Marshal the route details into JSON format
	jsonData, err := json.Marshal(routeDetails)
	if err != nil {
		http.Error(w, "Failed to marshal route details", http.StatusInternalServerError)
		log.Println("Error marshaling route details:", err)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.Write(jsonData)
}

func getRouteDetailsFromDB(routeID string) (Route, error) {
	var route Route

	// Convert routeID to int
	routeIDInt, err := strconv.Atoi(routeID)
	if err != nil {
		return route, err
	}

	// Query to fetch route details
	row := db.QueryRow(`
        SELECT r.sacco_id, GROUP_CONCAT(rc.checkpoint_id) AS checkpoint_ids
        FROM routes r 
        INNER JOIN route_checkpoints rc ON r.id = rc.route_id
        WHERE r.id = ?
        GROUP BY r.sacco_id;
    `, routeIDInt)

	// Initialize variables to store the fetched data
	var saccoID int
	var checkpointIDs string

	// Scan the row into variables
	err = row.Scan(&saccoID, &checkpointIDs)
	if err != nil {
		return route, err
	}

	// Convert checkpointIDs string to []int
	checkpointIDsArr := make([]int, 0)
	ids := strings.Split(checkpointIDs, ",")
	for _, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return route, err
		}
		checkpointIDsArr = append(checkpointIDsArr, idInt)
	}

	// Fetch SACCO name based on SACCO ID
	saccoName, err := getSaccoName(saccoID)
	if err != nil {
		return route, err
	}

	// Populate the route struct
	route = Route{
		ID:            routeIDInt,
		SaccoID:       saccoID,
		CheckpointIDs: checkpointIDsArr,
		SaccoName:     saccoName,
	}

	return route, nil
}

func deleteRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract route ID from query parameter
	routeID := r.URL.Query().Get("routeid")
	log.Println("Deleting route with ID:", routeID)

	// Delete associated records in route_checkpoints table
	_, err := db.Exec("DELETE FROM route_checkpoints WHERE route_id = ?", routeID)
	if err != nil {
		http.Error(w, "Error deleting associated route checkpoints", http.StatusInternalServerError)
		log.Println("Failed to delete associated route checkpoints:", err)
		return
	}

	// Execute the delete query
	_, err = db.Exec("DELETE FROM routes WHERE id = ?", routeID)
	if err != nil {
		http.Error(w, "Error deleting route", http.StatusInternalServerError)
		log.Println("Failed to delete route:", err)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Route deleted successfully"))
}
