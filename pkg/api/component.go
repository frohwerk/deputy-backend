package api

import (
	"fmt"
)

type Event struct {
	EventType string    `json:"type"`
	Object    Component `json:"object,omitempty"`
}

type Component struct {
	Id          string       `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Deployments []Deployment `json:"deployments,omitempty"`
}

func (a *Component) String() string {
	return fmt.Sprintf(
		"{Name: '%v', Deployments: ['%v']}", a.Name, a.Deployments,
	)
}
