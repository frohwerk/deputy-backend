package sqlstate

import (
	"github.com/lib/pq"
)

func UniqueViolation(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code == uniqueViolation
	}
	return false
}
