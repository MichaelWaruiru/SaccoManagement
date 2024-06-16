package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authorized
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Retrieve the search query from the request
	query := r.FormValue("q")
	if query == "" {
		http.Error(w, "Search query is missing", http.StatusBadRequest)
		return
	}

	var drivers []Driver
	var cars []Car
	var saccos []Sacco
	var managers []Sacco

	// Search for drivers
	drivers, err := searchDriverInDB(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error searching drivers:", err)
		return
	}

	// Search for cars
	cars, err = searchCarInDB(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error searching cars:", err)
		return
	}

	// Search for saccos
	saccos, err = searchSaccoInDB(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error searching saccos:", err)
		return
	}

	// Search for managers
	managers, err = searchManagerFromSaccosInDB(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error searching managers:", err)
		return
	}

	// Render the search results in a JSON response
	searchResults := struct {
		Drivers  []Driver
		Cars     []Car
		Saccos   []Sacco
		Managers []Sacco
	}{
		Drivers:  drivers,
		Cars:     cars,
		Saccos:   saccos,
		Managers: managers,
	}

	// Convert searchResults to JSON
	jsonData, err := json.Marshal(searchResults)
	if err != nil {
		http.Error(w, "Error converting search results to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func searchManagerFromSaccosInDB(query string) ([]Sacco, error) {
	var results []Sacco

	// Query the database to search for managers
	rows, err := db.Query("SELECT id, sacco_name, manager, contact FROM saccos WHERE manager LIKE ?", "%"+query+"%")
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
		results = append(results, sacco)
	}

	return results, nil
}

func searchSaccoInDB(query string) ([]Sacco, error) {
	var results []Sacco

	// Query the database to search for saccos
	rows, err := db.Query("SELECT id, sacco_name, manager, contact FROM saccos WHERE sacco_name LIKE ?", "%"+query+"%")
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

		results = append(results, sacco)
	}

	// Log the results
	// log.Println("Sacco search results:", results)

	return results, nil
}

func searchCarInDB(query string) ([]Car, error) {
	var results []Car

	// Query database to search for cars
	rows, err := db.Query(`
        SELECT cars.id, cars.number_plate, cars.make, cars.model,
               cars.no_of_passengers, cars.fare, cars.sacco_id,
               COALESCE(saccos.sacco_name, '') AS sacco_name
        FROM cars
        LEFT JOIN saccos ON cars.sacco_id = saccos.id
        WHERE cars.number_plate LIKE ? OR cars.make LIKE ? OR cars.model LIKE ?
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var car Car
		var saccoID int // Temporary variable to hold the scanned value
		// Scan the results into temporary variables
		err := rows.Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model,
			&car.NumberOfPassengers, &car.Fare, &saccoID, &car.SaccoName)
		if err != nil {
			return nil, err
		}

		// Set default value if SaccoID is 0
		if saccoID == 0 {
			saccoID = -1
		}

		// Assign values to the Car struct
		car.SaccoID = saccoID

		// Append the car to the results slice
		results = append(results, car)
	}

	// Log the results
	// log.Println("Car search results:", results)

	return results, nil
}

// Search for drivers in the database and retrieve relevant information
func searchDriverInDB(query string) ([]Driver, error) {
	var results []Driver

	// Query database to search for drivers along with related car and sacco information
	rows, err := db.Query(`
        SELECT drivers.id, drivers.name, drivers.id_number, drivers.contact,
               COALESCE(cars.number_plate, '') AS number_plate,
               COALESCE(saccos.sacco_name, '') AS sacco_name
        FROM drivers
        LEFT JOIN cars ON drivers.car_id = cars.id
        LEFT JOIN saccos ON drivers.sacco_id = saccos.id
        WHERE drivers.name LIKE ? OR drivers.id_number LIKE ? OR drivers.contact LIKE ?
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var driver Driver
		// Scan the results into the Driver struct
		err := rows.Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact,
			&driver.NumberPlate, &driver.SaccoName)
		if err != nil {
			return nil, err
		}
		// Append the driver to the results slice
		results = append(results, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func searchSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter from URL
	query := r.URL.Query().Get("q")

	// Perform search suggestions based on the query
	suggestions, err := searchSuggestions(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error searching suggestions:", err)
		return
	}

	// Encode suggestions as JSON and write to response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func searchSuggestions(query string) ([]string, error) {
	var suggestions []string

	// List of tables to search
	tables := []string{"saccos", "cars", "drivers"}

	// Iterate over tables and perform search in each table
	for _, table := range tables {
		tableSuggestions, err := searchTableSuggestions(table, query)
		if err != nil {
			return nil, err
		}
		suggestions = append(suggestions, tableSuggestions...)
	}

	return suggestions, nil
}

func searchTableSuggestions(table, query string) ([]string, error) {
	var tableSuggestions []string

	switch table {
	case "saccos":
		// Search suggestions in the saccos table
		saccoSuggestions, err := searchSaccoSuggestions(query)
		if err != nil {
			return nil, err
		}
		tableSuggestions = append(tableSuggestions, saccoSuggestions...)
	case "cars":
		// Search suggestions in the cars table
		carSuggestions, err := searchCarSuggestions(query)
		if err != nil {
			return nil, err
		}
		tableSuggestions = append(tableSuggestions, carSuggestions...)
	case "drivers":
		// Search suggestions in the drivers table
		driverSuggestions, err := searchDriverSuggestions(query)
		if err != nil {
			return nil, err
		}
		tableSuggestions = append(tableSuggestions, driverSuggestions...)
	// Add cases for other tables as needed
	default:
		// Handle unknown table
		return nil, fmt.Errorf("unknown table: %s", table)
	}

	return tableSuggestions, nil
}

func searchSaccoSuggestions(query string) ([]string, error) {
	var suggestions []string

	// Query the database to search for sacco name suggestions
	rows, err := db.Query("SELECT sacco_name, manager, contact FROM saccos WHERE sacco_name LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and append sacco name suggestions to the slice
	for rows.Next() {
		var saccoName, manager, contact string
		if err := rows.Scan(&saccoName, &manager, &contact); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, saccoName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If no sacco name suggestions found, query for manager suggestions
	if len(suggestions) == 0 {
		rows, err := db.Query("SELECT manager FROM saccos WHERE manager LIKE ?", "%"+query+"%")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Iterate over the rows and append manager suggestions to the slice
		for rows.Next() {
			var manager string
			if err := rows.Scan(&manager); err != nil {
				return nil, err
			}
			suggestions = append(suggestions, manager)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return suggestions, nil
}

func searchCarSuggestions(query string) ([]string, error) {
	var suggestions []string

	// Query database to search for car suggestions
	rows, err := db.Query(`
        SELECT number_plate
        FROM cars
        WHERE number_plate LIKE ? OR make LIKE ? OR model LIKE ?
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and append suggestions to the slice
	for rows.Next() {
		var suggestion string
		if err := rows.Scan(&suggestion); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return suggestions, nil
}

func searchDriverSuggestions(query string) ([]string, error) {
	var suggestions []string

	// Query database to search for driver suggestions
	rows, err := db.Query(`
        SELECT name
        FROM drivers
        WHERE name LIKE ? OR id_number LIKE ? OR contact LIKE ?
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and append suggestions to the slice
	for rows.Next() {
		var suggestion string
		if err := rows.Scan(&suggestion); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return suggestions, nil
}

func getDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the suggestion from the request
	suggestion := r.URL.Query().Get("suggestion")

	// Fetch details based on the suggestion
	details, err := getDetailsFromSuggestion(suggestion)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error fetching details:", err)
		return
	}

	// Convert details to JSON and write to response
	jsonData, err := json.Marshal(details)
	if err != nil {
		http.Error(w, "Error converting details to JSON", http.StatusInternalServerError)
		log.Println("Error converting details to JSON:", err)
		return
	}

	// Encode details as JSON and write to response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func getDetailsFromSuggestion(suggestion string) (interface{}, error) {
	// First, check if the suggestion directly matches a known driver, sacco, or car
	details, err := getDirectDetails(suggestion)
	if err != nil {
		return nil, err
	}
	if details != nil {
		// fmt.Println("Details found:", details)
		return details, nil
	}

	// If no direct match found, search each table for potential matches
	tables := []string{"saccos", "cars", "drivers"}
	for _, table := range tables {
		details, err := getDetailsFromTable(table, suggestion)
		if err != nil {
			return nil, err
		}
		if details != nil {
			// fmt.Println("Details found:", details)
			return details, nil
		}
	}

	// If no match found in any table, return an error
	fmt.Println("No details found for suggestion:", suggestion)
	return nil, fmt.Errorf("unknown suggestion: %s", suggestion)
}

func getDirectDetails(suggestion string) (interface{}, error) {
	// Check if suggestion directly matches a driver
	driverDetails, err := getDriverDetails(suggestion)
	if err == nil && len(driverDetails) > 0 {
		return driverDetails, nil
	}

	// Check if suggestion directly matches a sacco
	saccoDetails, err := getSaccoDetails(suggestion)
	if err == nil && saccoDetails != nil {
		return saccoDetails, nil
	}

	// Check if suggestion directly matches a car
	carDetails, err := getCarDetails(suggestion)
	if err == nil && len(carDetails) > 0 {
		return carDetails, nil
	}

	// Check if suggestion matches manager's name
	managerDetails, err := getSaccoDetailsByManager(suggestion)
	if err == nil && managerDetails != nil {
		return managerDetails, nil
	}

	// If no direct match found, return nil without error
	return nil, nil
}

// Function to fetch details from a specific table based on the suggestion
func getDetailsFromTable(table, suggestion string) (interface{}, error) {
	switch table {
	case "saccos":
		return getSaccoDetails(suggestion)
	case "cars":
		return getCarDetails(suggestion)
	case "drivers":
		return getDriverDetails(suggestion)
	default:
		return nil, fmt.Errorf("unknown table: %s", table)
	}
}

// Fetch details from the saccos table
func getSaccoDetails(suggestion string) (*Sacco, error) {
	// log.Printf("Executing query to get sacco details for suggestion '%s'\n", suggestion)

	var sacco Sacco

	// Query to retrieve sacco details
	query := "SELECT id, sacco_name, manager, contact FROM saccos WHERE sacco_name = ?"

	// Execute the query
	err := db.QueryRow(query, suggestion).Scan(&sacco.ID, &sacco.SaccoName, &sacco.Manager, &sacco.Contact)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Suggestion not found, return nil without error
		}
		return nil, err
	}

	return &sacco, nil
}

// Fetch details from the cars table
func getCarDetails(suggestion string) ([]Car, error) {
	// log.Printf("Executing query to get car details for suggestion '%s'\n", suggestion)

	var cars []Car
	query := `
        SELECT id, number_plate, make, model, no_of_passengers, fare, sacco_id
        FROM cars
        WHERE number_plate = ? OR make = ? OR model = ?
    `

	// Execute the query
	rows, err := db.Query(query, suggestion, suggestion, suggestion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var car Car
		err := rows.Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model, &car.NumberOfPassengers, &car.Fare, &car.SaccoID)
		if err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}
	// fmt.Println("Cars: ", cars)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

// Fetch details from the drivers table
func getDriverDetails(suggestion string) ([]Driver, error) {
	// log.Printf("Executing query to get driver details for suggestion '%s'\n", suggestion)

	var drivers []Driver
	query := `
        SELECT drivers.id, drivers.name, drivers.id_number, drivers.contact,
               COALESCE(cars.number_plate, '') AS number_plate,
               COALESCE(saccos.sacco_name, '') AS sacco_name
        FROM drivers
        LEFT JOIN cars ON drivers.car_id = cars.id
        LEFT JOIN saccos ON drivers.sacco_id = saccos.id
        WHERE drivers.name = ? OR drivers.id_number = ? OR drivers.contact = ?
    `

	// Execute the query
	rows, err := db.Query(query, suggestion, suggestion, suggestion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var driver Driver
		err := rows.Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact, &driver.NumberPlate, &driver.SaccoName)
		if err != nil {
			return nil, err
		}
		drivers = append(drivers, driver)
	}
	// fmt.Println("Drivers: ", drivers)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drivers, nil
}

// Fetch managers details from the saccos table
func getSaccoDetailsByManager(manager string) (*Sacco, error) {
	// log.Printf("Executing query to get sacco details for manager '%s'\n", manager)

	var sacco Sacco

	// Query to retrieve manager's details
	query := "SELECT id, sacco_name, manager, contact FROM saccos WHERE manager = ?"

	// Execute the query
	err := db.QueryRow(query, manager).Scan(&sacco.ID, &sacco.SaccoName, &sacco.Manager, &sacco.Contact)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Suggestion not found, return nil without error
		}
		return nil, err
	}

	return &sacco, nil
}
