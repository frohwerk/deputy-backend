package rollout

import (
	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
)

func Strategy(lookup dependencies.Lookup) *strategy {
	return &strategy{Lookup: lookup}
}

type strategy struct {
	Lookup dependencies.Lookup
}
