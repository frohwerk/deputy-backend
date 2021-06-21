package test

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
)

var _mutex sync.Mutex
var _db *sql.DB

func DB() *sql.DB {
	_mutex.Lock()
	defer _mutex.Unlock()

	if _db == nil {
		db, err := sql.Open("postgres", "postgres://test:drowssap@database:5432/test?sslmode=disable")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_db = db
	}

	return _db
}
