package database

import (
	"database/sql"
	"log"
)

var db *sql.DB

func init() {
	connStr := "user=postgres password=102030dD host=localhost port=5432 dbname=avito sslmode=disable"
	var err error
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
