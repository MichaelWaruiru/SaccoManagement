package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var (
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
	db   *sql.DB
)

func main() {
	var err error
	db, err = sql.Open("mysql", "root:Developer2023@tcp(localhost:3306)/fleet")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Serve all the satic files of HTML, CSS and JS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/get-sacco-data", getSaccoDataHandler)

	http.HandleFunc("/sacco", saccoHandler)
	http.HandleFunc("/add-sacco", addSaccoHandler)
	http.HandleFunc("/get-sacco-details", getSaccoDetailsHandler)
	http.HandleFunc("/edit-sacco", editSaccoHandler)
	http.HandleFunc("/delete-sacco", deleteSaccoHandler)

	http.HandleFunc("/get-cars-and-drivers-routes", getCarsAndDriversAndRoutesHandler)

	http.HandleFunc("/cars", carsHandler)
	http.HandleFunc("/add-car", addCarHandler)
	http.HandleFunc("/edit-car", editCarHandler)
	http.HandleFunc("/get-car-details/", getCarDetailsHandler)
	http.HandleFunc("/delete-car/", deleteCarHandler)
	http.HandleFunc("/filter-cars", filterCarsHandler)

	http.HandleFunc("/drivers", driversHandler)
	http.HandleFunc("/add-driver", addDriverHandler)
	http.HandleFunc("/edit-driver", editDriverHandler)
	http.HandleFunc("/get-driver-details/", getDriverDetailsHandler)
	http.HandleFunc("/delete-driver/", deleteDriverHandler)
	http.HandleFunc("/filter-drivers", filterDriversHandler)

	http.HandleFunc("/checkpoints", checkpointsHandler)
	http.HandleFunc("/add-checkpoint", addCheckpointsHandler)

	http.HandleFunc("/routes", routesHandler)
	http.HandleFunc("/create-route", createRouteHandler)
	http.HandleFunc("/get-routes", getRoutesHandler)
	http.HandleFunc("/edit-route", editRouteHandler)
	http.HandleFunc("/get-route-details", getRouteDetailsHandler)
	http.HandleFunc("/delete-route", deleteRouteHandler)

	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/search-suggestions", searchSuggestionsHandler)
	http.HandleFunc("/get-details", getDetailsHandler)

	http.HandleFunc("/mark-checkpoint", markCheckpointHandler)
	http.HandleFunc("/add-mark-checkpoint", addMarkCheckpoint)
	http.HandleFunc("/marked-checkpoints", getMarkedCheckpointsHandler)

	fmt.Println("Server is running")
	http.ListenAndServe(":8080", nil)
}
