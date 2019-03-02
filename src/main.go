package main

import (
	"db"
	"log"
)

func main() {
	d, err := db.NewMySQLDB()
	if err != nil {
		log.Fatal(err)
	}

	// Close connection to db
	d.Close()
}
