package main

import (
	"db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var d db.MysqlDB

/* EXAMPLE
{
    "ID": 1,
    "Header": "Test",
    "Description": "body",
    "Author": "dan",
    "Lat": 28.602427,
    "Lon": -81.200058,
    "Community": "Food"
}
*/

/* EXAMPLE 2
{
	report := db.Report{ID: 1, Header: "Test", Description: "body", Author: "dan",
		Lat: 28.602427, Lon: -81.200058, Community: "Food"}
}
*/

func main() {
	d, err := db.NewMySQLDB()
	if err != nil {
		log.Fatal(err)
	}

	// reports := []db.Report{}

	// report := db.Report{ID: 1, Header: "Test", Description: "body", Author: "dan",
	// 	Lat: 28.602427, Lon: -81.200058, Community: "Food"}
	// d.AddReport(&report)

	// allreps, _ := d.ListReports()
	// for _, i := range allreps {
	// 	reports = append(reports, *i)
	// }

	// fmt.Println(reports)

	r := mux.NewRouter()
	// r.HandleFunc("/", GetReportsHandler)      // GET
	r.HandleFunc("/report", SetReportHandler) // POST

	log.Fatal(http.ListenAndServe(":8000", r))

	d.Close() // Close connection to db
}

// GetReportsHandler - Load all reports
// func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
// 	var reports []db.Report

// 	allreps, _ := d.ListReports()
// 	for _, i := range allreps {
// 		reports = append(reports, *i)
// 	}

// 	fmt.Println(reports)

// 	// Turn it into a json object and send it to the front-end
// 	json.NewEncoder(w).Encode(reports)
// }

// SetReportHandler - Receive a json object and send it to the DB
func SetReportHandler(w http.ResponseWriter, r *http.Request) {
	var report db.Report
	json.NewDecoder(r.Body).Decode(&report) // Decode json and put into report
	fmt.Println(report)                     // now has the data in the struct

	d.AddReport(report)
	json.NewEncoder(w).Encode(report)
}
