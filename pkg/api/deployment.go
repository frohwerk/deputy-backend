package api

import (
	"time"
)

type Deployment struct {
	ComponentId string
	PlatformId  string
	ImageRef    string
	Updated     time.Time
}
