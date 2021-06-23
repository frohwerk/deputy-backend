package test

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/job"
)

func Test(t *testing.T) {
	t.Run("thingy", func(t *testing.T) {
		var j job.Runner = mock{}
		j.Run(job.Params{})
	})
}

type mock struct{}

func (m *mock) Run(p job.Params, out job.Output) error {
	return nil
}
