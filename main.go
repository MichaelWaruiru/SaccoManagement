package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
	db   *sql.DB
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get database credentials from environment variables
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	// Construct database connection string
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	// Open database connection
	db, err = sql.Open("mysql", dbURI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ping database to check connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to database!")

	// Serve all the satic files of HTML, CSS and JS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/home", sessionMiddleware(homeHandler))
	http.HandleFunc("/get-sacco-data", sessionMiddleware(getSaccoDataHandler))

	http.HandleFunc("/sacco", sessionMiddleware(saccoHandler))
	http.HandleFunc("/add-sacco", sessionMiddleware(addSaccoHandler))
	http.HandleFunc("/get-sacco-details", sessionMiddleware(getSaccoDetailsHandler))
	http.HandleFunc("/edit-sacco", sessionMiddleware(editSaccoHandler))
	http.HandleFunc("/delete-sacco", sessionMiddleware(deleteSaccoHandler))

	http.HandleFunc("/get-cars-and-drivers-routes", sessionMiddleware(getCarsAndDriversAndRoutesHandler))
	http.HandleFunc("/filter-trips", sessionMiddleware(filterTripsHandler))

	http.HandleFunc("/cars", sessionMiddleware(carsHandler))
	http.HandleFunc("/add-car", sessionMiddleware(addCarHandler))
	http.HandleFunc("/edit-car", sessionMiddleware(editCarHandler))
	http.HandleFunc("/get-car-details/", sessionMiddleware(getCarDetailsHandler))
	http.HandleFunc("/delete-car/", sessionMiddleware(deleteCarHandler))
	http.HandleFunc("/filter-cars", sessionMiddleware(filterCarsHandler))

	http.HandleFunc("/drivers", sessionMiddleware(driversHandler))
	http.HandleFunc("/add-driver", sessionMiddleware(addDriverHandler))
	http.HandleFunc("/edit-driver", sessionMiddleware(editDriverHandler))
	http.HandleFunc("/get-driver-details/", sessionMiddleware(getDriverDetailsHandler))
	http.HandleFunc("/delete-driver/", sessionMiddleware(deleteDriverHandler))
	http.HandleFunc("/filter-drivers", sessionMiddleware(filterDriversHandler))

	http.HandleFunc("/checkpoints", sessionMiddleware(checkpointsHandler))
	http.HandleFunc("/add-checkpoint", sessionMiddleware(addCheckpointsHandler))

	http.HandleFunc("/routes", sessionMiddleware(routesHandler))
	http.HandleFunc("/create-route", sessionMiddleware(createRouteHandler))
	http.HandleFunc("/get-routes", sessionMiddleware(getRoutesHandler))
	http.HandleFunc("/edit-route", sessionMiddleware(editRouteHandler))
	http.HandleFunc("/get-route-details", sessionMiddleware(getRouteDetailsHandler))
	http.HandleFunc("/delete-route", sessionMiddleware(deleteRouteHandler))

	http.HandleFunc("/search", sessionMiddleware(searchHandler))
	http.HandleFunc("/search-suggestions", sessionMiddleware(searchSuggestionsHandler))
	http.HandleFunc("/get-details", sessionMiddleware(getDetailsHandler))

	http.HandleFunc("/mark-checkpoint", sessionMiddleware(markCheckpointHandler))
	http.HandleFunc("/add-mark-checkpoint", sessionMiddleware(addMarkCheckpoint))
	http.HandleFunc("/marked-checkpoints", sessionMiddleware(getMarkedCheckpointsHandler))

	fmt.Println("Server is running")
	http.ListenAndServe(":8080", nil)
}
