package api

import "fmt"

type Event struct {
	EventType string    `json:"type"`
	Object    Component `json:"object,omitempty"`
}

type Component struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Image string `json:"image,omitempty"`
}

func (a *Component) String() string {
	return fmt.Sprintf(
		"{Name: '%v', Type: '%v', Image: '%v'}", a.Name, a.Type, a.Image,
	)
}
