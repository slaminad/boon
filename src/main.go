package main

import (
	"db"
	"fmt"
	"log"
)

func main() {
	d, err := db.NewMySQLDB()
	if err != nil {
		log.Fatal(err)
	}

	n, _ := d.AddReport(&db.Report{ID: 0, Header: "Test", Description: "Alert", Author: "Dan"})
	fmt.Println(n)

	r, _ := d.ListReports()
	for _, i := range r {
		fmt.Println(*i)
	}

	// Close connection to db
	d.Close()
}
