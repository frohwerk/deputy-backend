package database

type scanner interface {
	Scan(dest ...interface{}) error
}
