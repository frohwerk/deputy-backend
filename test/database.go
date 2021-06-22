package test

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"
)

var _mutex sync.Mutex
var _db *sql.DB

func DB() *sql.DB {
	_mutex.Lock()
	defer _mutex.Unlock()

	if _db == nil {
		db, err := sql.Open("postgres", "postgres://test:drowssap@localhost:5433/test?sslmode=disable")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error on open:", err)
			os.Exit(1)
		}
		_db = db
	}

	err := error(nil)
	for i := 0; i < 15; i++ {
		var now sql.NullTime
		err = _db.QueryRow(`SELECT CURRENT_TIMESTAMP`).Scan(&now)
		if err == nil {
			return _db
		}
		if i%5 == 0 {
			fmt.Println("Database not available (yet?). Sleeping...")
		}
		time.Sleep(time.Second)
	}

	fmt.Fprintln(os.Stderr, "error connecting to database:", err)
	os.Exit(1)
	return nil
}
