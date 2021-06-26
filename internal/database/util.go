package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const local = "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable"

func GetConninfo() (string, error) {
	url, user, password := os.Getenv("POSTGRESQL_URL"), os.Getenv("POSTGRESQL_USER"), os.Getenv("POSTGRESQL_PASSWORD")
	switch {
	case url == "":
		fmt.Fprintf(os.Stderr, "Invalid database url: %s", url)
		return local, nil
	case user == "":
		fmt.Fprintf(os.Stderr, "Invalid database user: %s", url)
		return local, nil
	case password == "":
		fmt.Fprintf(os.Stderr, "Invalid database password: %s", url)
		return local, nil
	}
	if i := strings.Index(url, "://") + 3; 3 < i && i < len(url) {
		return fmt.Sprintf("%s%s:%s@%s", url[:i], user, password, url[i:]), nil
	} else {
		return "", fmt.Errorf("Invalid database url: %s", url)
	}
}

func Open() *sql.DB {
	conninfo, err := GetConninfo()
	if err != nil {
		log.Println(err)
		<-time.NewTimer(15 * time.Second).C
		conninfo = local
		os.Exit(1)
	}
	fmt.Println("Connecting to database:", conninfo)
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Println(err)
		<-time.NewTimer(15 * time.Second).C
		os.Exit(1)
	}
	return db
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("ERROR db.Close() failed: %s\n", err)
	}
}
