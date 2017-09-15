package gosubscribe

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Connect to a given PostreSQL database.
func Connect(host, user, dbname, password string) *sql.DB {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s user=%s dbname=%s password=%s",
			host, user, dbname, password,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
