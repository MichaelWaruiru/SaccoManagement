package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

//	type Driver struct {
//		ID          int
//		Name        string
//		IDNumber    int
//		Contact     string
//		CarID       int
//		SaccoID     int
//		NumberPlate string
//		SaccoName   string
//	}
type Driver struct {
	ID          int `json:"id"`
	Name        string
	IDNumber    string
	Contact     string
	CarID       int    `json:"-"`
	NumberPlate string `json:"number_plate"`
	SaccoID     int    `json:"-"`
	SaccoName   string // `json:"sacco_name"`
}

func driversHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Retrieve drivers, saccos, and cars from the database
	drivers, err := getAllDrivers()
	if err != nil {
		http.Error(w, "Error getting all drivers", http.StatusInternalServerError)
		fmt.Println("Error getting all drivers:", err)
		return
	}

	saccos, err := getAllSaccos()
	if err != nil {
		http.Error(w, "Failed to fetch saccos", http.StatusInternalServerError)
		return
	}

	cars, err := getAllCars()
	if err != nil {
		http.Error(w, "Failed to fetch cars", http.StatusInternalServerError)
		return
	}

	// Retrieve flash message
	flashMessage := getFlashMessage(w, r)

	// Prepare data to be passed to the template
	data := struct {
		Cars         []Car
		Saccos       []Sacco
		Drivers      []Driver
		FlashMessage string
	}{
		Cars:         cars,
		Saccos:       saccos,
		Drivers:      drivers,
		FlashMessage: flashMessage,
	}

	// Execute menu template
	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))

	// Execute menu template
	err = menuTemplate.Execute(w, drivers)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	// Render the drivers template with the data
	err = tmpl.ExecuteTemplate(w, "drivers", data)
	if err != nil {
		http.Error(w, "Error rendering the template", http.StatusInternalServerError)
		fmt.Println("Error rendering the template:", err)
		return
	}
}

func getAllDrivers() ([]Driver, error) {
	var drivers []Driver

	rows, err := db.Query(`
	SELECT d.id, d.name, d.id_number, d.contact, d.car_id, d.sacco_id, c.number_plate, s.sacco_name
	FROM drivers d
	JOIN cars c ON d.car_id = c.id
	JOIN saccos s ON d.sacco_id = s.id
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var driver Driver
		err := rows.Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact, &driver.CarID, &driver.SaccoID, &driver.NumberPlate, &driver.SaccoName)
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

func addDriverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form")
		return
	}

	// Extract form values
	name := r.FormValue("name")
	idNumber := r.FormValue("idNumber")
	contact := r.FormValue("contact")
	assignedCar := r.FormValue("assignedCar")
	assignedSacco := r.FormValue("assignedSacco")

	// fmt.Println("Assigned car:", assignedCar)
	existingDriverID, err := getDriverByIDNumber(idNumber)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error checking existing driver:", err)
		return
	}
	if existingDriverID != 0 {
		setFlashMessage(w, r, "Driver with ID number "+idNumber+" already exists")
		http.Redirect(w, r, "/drivers", http.StatusSeeOther)
		return
	}

	carID, err := getCarIDByNumberPlate(assignedCar)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error getting car ID by number plate:", err)
		return
	}

	saccoID, err := strconv.Atoi(assignedSacco)
	if err != nil {
		http.Error(w, "Invalid sacco ID", http.StatusBadRequest)
		fmt.Println("Error converting sacco to interger:", err)
		return
	}

	// Insert the new driver into the database
	_, err = db.Exec("INSERT INTO drivers (name, id_number, contact, car_id, sacco_id) VALUES (?, ?, ?, ?, ?)", name, idNumber, contact, carID, saccoID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error inserting driver:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Driver added successfully")
}

// Function to get car ID by number plate
func getCarIDByNumberPlate(numberPlate string) (int, error) {
	// Query the database to get the car ID by number plate
	var carID int
	err := db.QueryRow("SELECT id FROM cars WHERE UPPER(number_plate) = UPPER(?)", numberPlate).Scan(&carID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no car found with the number plate: %s", numberPlate)
		}
		return 0, err
	}

	return carID, nil
}

func getDriverByIDNumber(idNumber string) (int, error) {
	var driverID int
	err := db.QueryRow("SELECT id FROM drivers WHERE id_number = ?", idNumber).Scan(&driverID)
	if err != nil {
		return 0, nil
	}

	return driverID, err
}

func editDriverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form")
		return
	}

	// Retrieve driver ID from the edit form
	driverID, err := strconv.Atoi(r.FormValue("editDriverID"))
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		fmt.Println("Error converting driver ID to integer:", err)
		return
	}

	name := r.FormValue("name")
	idNumber := r.FormValue("idNumber")
	contact := r.FormValue("contact")
	assignedCar := r.FormValue("assignedCar")
	assignedSacco := r.FormValue("assignedSacco")

	// Update the database
	_, err = db.Exec("UPDATE drivers SET name = ?, id_number = ?, contact = ?, car_id = ?, sacco_id = ? WHERE id = ?", name, idNumber, contact, assignedCar, assignedSacco, driverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error updating driver:", err)
		return
	}

	http.Redirect(w, r, "/drivers", http.StatusSeeOther)
}

func getDriverDetailsHandler(w http.ResponseWriter, r *http.Request) {
	driverID, err := strconv.Atoi(r.URL.Path[len("/get-driver-details/"):])
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		fmt.Println("Error converting driver ID to integer:", err)
		return
	}

	// Fetch driver details from the database using driver ID
	driver, err := getDriverByID(driverID)
	if err != nil {
		http.Error(w, "Failed to fetch driver details", http.StatusInternalServerError)
		fmt.Println("Error fetching driver details:", err)
		return
	}

	driverJSON, err := json.Marshal(driver)
	if err != nil {
		http.Error(w, "Failed to encode driver details", http.StatusInternalServerError)
		fmt.Println("Error encoding driver details:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(driverJSON)
}

func getDriverByID(driverID int) (Driver, error) {
	var driver Driver

	err := db.QueryRow(`
	SELECT d.id, d.name, d.id_number, d.contact, d.car_id, d.sacco_id, c.number_plate, s.sacco_name
		FROM drivers d
		JOIN cars c ON d.car_id = c.id
		JOIN saccos s ON d.sacco_id = s.id
		WHERE d.id = ?
	`, driverID).Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact,
		&driver.CarID, &driver.SaccoID, &driver.NumberPlate, &driver.SaccoName)
	if err != nil {
		return Driver{}, nil
	}

	return driver, nil
}

func deleteDriverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	// Extract driver ID from the query parameters
	driverID, _ := strconv.Atoi(r.URL.Query().Get("driverid"))

	// Delete the driver from the database
	_, err := db.Exec("DELETE FROM drivers WHERE id = ?", driverID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error deleting driver:", err)
		return
	}
}

func filterDriversHandler(w http.ResponseWriter, r *http.Request) {
	// Get the SACCO ID from the query parameters
	saccoID, err := strconv.Atoi(r.URL.Query().Get("saccoID"))
	if err != nil {
		http.Error(w, "Invalid SACCO ID", http.StatusBadRequest)
		fmt.Println("Error converting SACCO ID to integer:", err)
		return
	}

	// Retrieve filtered drivers from the database
	drivers, err := getFilteredDrivers(saccoID)
	if err != nil {
		http.Error(w, "Failed to fetch filtered drivers", http.StatusInternalServerError)
		fmt.Println("Error fetching filtered drivers:", err)
		return
	}

	// Convert drivers to JSON and send response
	jsonResponse, err := json.Marshal(drivers)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		fmt.Println("Error marshaling response:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func getFilteredDrivers(saccoID int) ([]Driver, error) {
	// Query the database to get the filtered drivers by SACCO ID
	rows, err := db.Query(`
        SELECT d.id, d.name, d.id_number, d.contact, d.car_id, d.sacco_id, c.number_plate, s.sacco_name
        FROM drivers d
        JOIN cars c ON d.car_id = c.id
        JOIN saccos s ON d.sacco_id = s.id
        WHERE d.sacco_id = ?
    `, saccoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []Driver

	for rows.Next() {
		var driver Driver
		err := rows.Scan(&driver.ID, &driver.Name, &driver.IDNumber, &driver.Contact, &driver.CarID, &driver.SaccoID, &driver.NumberPlate, &driver.SaccoName)
		if err != nil {
			return nil, err
		}
		drivers = append(drivers, driver)
	}

	return drivers, nil
}
