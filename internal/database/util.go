package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

const conninfo = "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable"

func Open() *sql.DB {
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return db
}

func NewListener() *pq.Listener {
	return pq.NewListener(conninfo, 5*time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
		switch event {
		case pq.ListenerEventConnected:
			fmt.Println("ListenerEventConnected:", err)
		case pq.ListenerEventDisconnected:
			fmt.Println("ListenerEventDisconnected:", err)
		case pq.ListenerEventReconnected:
			fmt.Println("ListenerEventReconnected:", err)
		case pq.ListenerEventConnectionAttemptFailed:
			fmt.Println("ListenerEventConnectionAttemptFailed:", err)
		default:
			fmt.Println("Unknown event type", event, err)
		}

	})
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("ERROR db.Close() failed: %s\n", err)
	}
}
