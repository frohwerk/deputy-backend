package internal

import "github.com/frohwerk/deputy-backend/pkg/api"

type Observable struct {
	Events <-chan api.Event
	Stop   func()
}

type Platform interface {
	GetComponents() ([]api.Component, error)
	WatchComponents() (Observable, error)
}
