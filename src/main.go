package main

import (
	"db"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var d db.MysqlDB

func main() {
	d, err := db.NewMySQLDB()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	registerHandlers()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

	d.Close() // Close connection to db
}

func registerHandlers() {
	r := mux.NewRouter()

	r.HandleFunc("/", GetReportsHandler)      // GET
	r.HandleFunc("/report", SetReportHandler) // POST
}

// GetReportsHandler - Load all reports
func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	var reports []db.Report

	allreps, _ := d.ListReports()
	for _, i := range allreps {
		reports = append(reports, *i)
	}
}

// SetReportHandler - Receive a json object and send it to the DB
func SetReportHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	// Create report struct and send it to DB handler
	report := createReportStruct()

	d.AddReport(&report)
}

func createReportStruct() db.Report {
	// Get JSON object and parse through it
	r := db.Report{}

	return r
}
