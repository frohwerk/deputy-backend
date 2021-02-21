package database

import (
	"database/sql"
	"log"
)

func Open() *sql.DB {
	db, err := sql.Open("postgres", "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return db
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("ERROR db.Close() failed: %s\n", err)
	}
}
