package main

import (
	"db"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

var d db.MysqlDB

func main() {
	d, err := db.NewMySQLDB()
	if err != nil {
		log.Fatal(err)
	}

	// n, _ := d.AddReport(&db.Report{ID: 0, Header: "Test", Description: "Alert", Author: "Dan",
	// 	Lat: 51.747169, Lon: -114.333801})
	// fmt.Println(n)

	d.DeleteReport(1)

	r, _ := d.ListReports()
	for _, i := range r {
		fmt.Println(*i)
	}

	// Close connection to db
	d.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	registerHandlers()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func registerHandlers() {
	r := mux.NewRouter()

	r.HandleFunc("/", GetReportsHandler) // GET
	// r.HandleFunc("/{filter}", FilterGetReportsHandler)    // GET
	r.HandleFunc("/report/{lat}/{lon}", SetReportHandler) // POST

}

// GetReportsHandler - Handles getting reports
func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	var reports []db.Report

	allreps, _ := d.ListReports()
	for _, i := range allreps {
		reports = append(reports, *i)
	}
}

// // FilterGetReportsHandler - Filters through reports
// func FilterGetReportsHandler(w http.ResponseWriter, r *http.Request) {
// 	var reports []db.Report
// 	f := mux.Vars(r)["filter"]

// 	filters := strings.Split(f, "-")

// 	allreps, _ := d.ListReports()
// 	for _, i := range allreps {
// 		reports = append(reports, *i)
// 	}
// }

// SetReportHandler - Handles setting a new report
func SetReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	la, _ := strconv.ParseFloat(vars["lat"], 64)
	lo, _ := strconv.ParseFloat(vars["lon"], 64)

	d.AddReport(&db.Report{ID: 0, Header: " ", Description: " ", Author: " ",
		Lat: la, Lon: lo})
}
