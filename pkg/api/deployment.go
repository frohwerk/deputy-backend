package api

import (
	"fmt"
	"time"
)

type Deployment struct {
	ImageRef string    `json:"image,omitempty"`
	Updated  time.Time `json:"updated,omitempty"`
}

func (d *Deployment) String() string {
	return fmt.Sprintf(
		"{ImageRef: '%v', Updated: '%s'}", d.ImageRef, d.Updated,
	)
}
