package main

import (
	"database/sql"
	"io"
	"log"

	_ "github.com/lib/pq"
)

type application struct {
	id   string
	name string
}

func main() {
	db, err := sql.Open("postgres", "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer close(db)

	rows, err := db.Query("SELECT * FROM apps")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer close(rows)

	for rows.Next() {
		a := new(application)
		if err := rows.Scan(&a.id, &a.name); err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Found row: %s => %s", a.id, a.name)
	}
}

func close(db io.Closer) {
	if err := db.Close(); err != nil {
		log.Printf("%v\n", err)
	}
}
