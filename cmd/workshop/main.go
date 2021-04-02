package main

import (
	"database/sql"
	"log"

	"github.com/frohwerk/deputy-backend/internal/database"
)

func main() {
	db, err := sql.Open("postgres", "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer db.Close()
	store := database.NewFileStore(db)
	f := &database.File{Name: "/app/main.js", Digest: "sha256:76a7059dc31c6bec6d0597bc500a093d4d5d914c35f83dcf58703abf2e6c1fe6"}
	archives, err := store.FindByContent(f)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	for _, a := range archives {
		log.Printf("Found archive %s (%s) containing file %s", a.Name, a.Id, f.Name)
	}
}
