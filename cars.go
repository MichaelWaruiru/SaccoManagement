package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Car struct {
	ID                 int
	NumberPlate        string
	Make               string
	Model              string
	NumberOfPassengers int
	Fare               int
	SaccoID            int
	SaccoName          string
	Trips              int
}

func carsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cars, err := getAllCars()
	if err != nil {
		http.Error(w, "Failed to fetch cars", http.StatusInternalServerError)
		log.Println("Error fetching cars:", err)
		return
	}

	// Fetch SACCO data from the database
	saccos, err := getAllSaccos()
	if err != nil {
		http.Error(w, "Failed to fetch saccos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Cars       []Car
		Saccos     []Sacco
		ActivePage string
	}{
		Cars:   cars,
		Saccos: saccos,
	}

	// Execute menu template
	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))

	// Execute menu template
	err = menuTemplate.Execute(w, cars)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	// Render the HTML template with the data
	err = tmpl.ExecuteTemplate(w, "cars", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error rendering template:", err)
		return
	}
}

func getAllCars() ([]Car, error) {
	rows, err := db.Query("SELECT c.id, c.number_plate, c.make, c.model, c.no_of_passengers, c.fare, c.sacco_id, s.sacco_name AS sacco_name FROM cars c JOIN saccos s ON c.sacco_id = s.id;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []Car

	for rows.Next() {
		var car Car
		err := rows.Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model, &car.NumberOfPassengers, &car.Fare, &car.SaccoID, &car.SaccoName)
		if err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	return cars, nil
}

// Handler for adding a car
func addCarHandler(w http.ResponseWriter, r *http.Request) {
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
	numberPlate := r.FormValue("numberPlate")
	make := r.FormValue("make")
	model := r.FormValue("model")
	numPassengers, _ := strconv.Atoi(r.FormValue("noOfPassengers"))
	fare, _ := strconv.Atoi(r.FormValue("fare"))
	saccoID, _ := strconv.Atoi(r.FormValue("saccoSelect"))

	// Checks if the car exist
	if carExists(numberPlate) {
		http.Error(w, "Car already exists", http.StatusBadRequest)
		return
	}

	// Insert the new car into the database
	_, err = db.Exec("INSERT INTO cars (number_plate, make, model, no_of_passengers, fare, sacco_id) VALUES (?, ?, ?, ?, ?, ?)",
		numberPlate, make, model, numPassengers, fare, saccoID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error inserting car:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)

	// Broadcast update after adding the car
	// updateClients()
}

func carExists(numberPlate string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cars WHERE number_plate = ?", numberPlate).Scan(&count)
	if err != nil {
		fmt.Println("Car already exists:", err)
		return false
	}
	return count > 0
}

// Handler for editing a car
func editCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form:", err)
		return
	}

	carID, err := strconv.Atoi(r.FormValue("editCarID"))
	if err != nil {
		http.Error(w, "Invalid car ID", http.StatusBadRequest)
		fmt.Println("Error converting car ID to integer:", err)
		return
	}

	// Extract form values
	numberPlate := r.FormValue("numberPlate")
	make := r.FormValue("make")
	model := r.FormValue("model")
	numPassengers, _ := strconv.Atoi(r.FormValue("noOfPassengers"))
	fare, _ := strconv.Atoi(r.FormValue("fare"))
	saccoID, _ := strconv.Atoi(r.FormValue("saccoSelect"))

	// Update the car details in the database
	_, err = db.Exec("UPDATE cars SET number_plate = ?, make = ?, model = ?, no_of_passengers = ?, fare = ?, sacco_id = ? WHERE id = ?",
		numberPlate, make, model, numPassengers, fare, saccoID, carID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error updating car:", err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Car updated successfully")
}

// Handler for getting car details for editing
func getCarDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract car ID from URL
	carID := r.URL.Path[len("/get-car-details/"):]

	// Convert carID to integer
	id, err := strconv.Atoi(carID)
	if err != nil {
		http.Error(w, "Invalid car ID", http.StatusBadRequest)
		return
	}

	// Fetch car details from the database by car ID
	car, err := getCarByID(id)
	if err != nil {
		http.Error(w, "Failed to get car details", http.StatusInternalServerError)
		log.Println("Error getting car details:", err)
		return
	}

	// Fetch SACCO data from the database
	saccos, err := getAllSaccos()
	if err != nil {
		http.Error(w, "Failed to fetch saccos", http.StatusInternalServerError)
		log.Println("Error fetching saccos:", err)
		return
	}

	// Combine car details and SACCO data into a single response object
	response := struct {
		Car    Car
		Saccos []Sacco
	}{
		Car:    car,
		Saccos: saccos,
	}

	// Marshal response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		log.Println("Error marshaling response:", err)
		return
	}

	// Set response content type and write JSON data to response body
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Function to get car details by ID from the database
func getCarByID(id int) (Car, error) {
	var car Car
	// Query the database to get the car details by ID
	err := db.QueryRow("SELECT c.id, c.number_plate, c.make, c.model, c.no_of_passengers, c.fare, c.sacco_id, s.sacco_name AS sacco_name FROM cars c JOIN saccos s ON c.sacco_id = s.id WHERE c.id = ?", id).
		Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model, &car.NumberOfPassengers, &car.Fare, &car.SaccoID, &car.SaccoName)
	if err != nil {
		return Car{}, err
	}
	return car, nil
}

// Handler for deleting a car
func deleteCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract car ID from query parameters
	carID, _ := strconv.Atoi(r.URL.Query().Get("carid"))

	// Temporarily disable foreign key constraints
	// _, err := db.Exec("SET FOREIGN_KEY_CHECKS=0")
	// if err != nil {
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	log.Println("Error disabling foreign key constraints:", err)
	// 	return
	// }

	// Delete the car from the database
	_, err := db.Exec("DELETE FROM cars WHERE id = ?", carID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error deleting car:", err)
		return
	}

	// _, err = db.Exec("DELETE FROM cars WHERE id = ?", carID)
	// if err != nil {
	// 	// Re-enable foreign key constraints if there's an error
	// 	_, err := db.Exec("SET FOREIGN_KEY_CHECKS=1")
	// 	if err != nil {
	// 		log.Println("Error re-enabling foreign key constraints:", err)
	// 	}
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	log.Println("Error deleting car:", err)
	// 	return
	// }

	// Re-enable foreign key constraints after successful deletion
	// _, err = db.Exec("SET FOREIGN_KEY_CHECKS=1")
	// if err != nil {
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	log.Println("Error re-enabling foreign key constraints:", err)
	// 	return
	// }

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Car deleted successfully"}`)
}

// Handler for filtering cars by SACCO
func filterCarsHandler(w http.ResponseWriter, r *http.Request) {
	saccoID, err := strconv.Atoi(r.URL.Query().Get("saccoID"))
	if err != nil {
		http.Error(w, "Invalid SACCO ID", http.StatusBadRequest)
		return
	}

	// Fetch filtered cars from the database
	cars, err := getFilteredCars(saccoID)
	if err != nil {
		http.Error(w, "Failed to fetch filtered cars", http.StatusInternalServerError)
		log.Println("Error fetching filtered cars:", err)
		return
	}

	// Convert cars to JSON and send response
	jsonResponse, err := json.Marshal(cars)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		log.Println("Error marshaling response:", err)
		return
	}

	// Set response content type and write JSON data to response body
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Function to get filtered cars by SACCO from the database
func getFilteredCars(saccoID int) ([]Car, error) {
	// Query the database to get the filtered cars by SACCO ID
	rows, err := db.Query("SELECT c.id, c.number_plate, c.make, c.model, c.no_of_passengers, c.fare, c.sacco_id, s.sacco_name AS sacco_name FROM cars c JOIN saccos s ON c.sacco_id = s.id WHERE c.sacco_id = ?", saccoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []Car

	for rows.Next() {
		var car Car
		err := rows.Scan(&car.ID, &car.NumberPlate, &car.Make, &car.Model, &car.NumberOfPassengers, &car.Fare, &car.SaccoID, &car.SaccoName)
		if err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	return cars, nil
}

// // Function to broadcast update to clients
// func updateClients() {
// 	// Count vehicles and routes
// 	vehiclesCount, err := countVehicles()
// 	if err != nil {
// 		log.Println("Error counting vehicles:", err)
// 		return
// 	}

// 	// Create message payload
// 	message := struct {
// 		VehiclesCount int
// 	}{
// 		VehiclesCount: vehiclesCount,
// 	}

// 	// Convert message to JSON
// 	messageJSON, err := json.Marshal(message)
// 	if err != nil {
// 		log.Println("Error marshaling message:", err)
// 		return
// 	}

// 	// Broadcast message to clients
// 	broadcastMessage(messageJSON)
// }

// // Function to count vehicles
// func countVehicles() (int, error) {
// 	var count int
// 	err := db.QueryRow("SELECT COUNT(*) FROM cars").Scan(&count)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return count, nil
// }
