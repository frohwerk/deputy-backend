package job_test

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/job"
	"github.com/stretchr/testify/assert"
)

type store struct {
	values []string
}

func (s *store) Write(v string) {
	s.values = append(s.values, v)
}

func TestOutputBuffer(t *testing.T) {
	t.Run("write one line", func(t *testing.T) {
		buf := &job.OutputBuffer{}
		buf.Write("Hallo Welt")
		assert.Equal(t, []string{"Hallo Welt"}, buf.Get())
	})

	t.Run("write two lines", func(t *testing.T) {
		buf := &job.OutputBuffer{}
		buf.Write("Hallo Welt!")
		buf.Write("Was geht?")
		assert.Equal(t, []string{"Hallo Welt!", "Was geht?"}, buf.Get())
	})

	t.Run("preserves content", func(t *testing.T) {
		buf := &job.OutputBuffer{}
		buf.Write("Hallo Welt!")
		buf.Get()[0] = ""
		assert.Equal(t, []string{"Hallo Welt!"}, buf.Get())
	})

	// Each job has an OutputBuffer
	// An interested party can subscribe to new messages
	// An interested party can get the whole content of the buffer
}
