package database

import "fmt"

type ErrDatabaseAccess struct {
	msg string
}

type ErrNotFound struct {
	ErrDatabaseAccess
}

func (e *ErrDatabaseAccess) Error() string {
	return e.msg
}

func (e *ErrNotFound) Error() string {
	return e.msg
}

func wrap(err error) *ErrDatabaseAccess {
	return &ErrDatabaseAccess{msg: fmt.Sprintf("%v", err)}
}
