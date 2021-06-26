package job

import (
	"fmt"

	"github.com/frohwerk/deputy-backend/internal/logger"
)

var Log logger.Logger = logger.Default

type Output interface {
	Write(format string, args ...interface{})
}

type OutputBuffer struct {
	buf []string
}

func (o *OutputBuffer) Write(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	defer func() { fmt.Println("o.buf (after): ", o.buf) }()
	Log.Trace("OutputBuffer::Write => %s", s)
	Log.Trace("o.buf (before): %s", o.buf)
	o.buf = append(o.buf, s)
}

func (o *OutputBuffer) Get() []string {
	Log.Trace("OutputBuffer::Get")
	Log.Trace("o.buf %s", o.buf)
	res := make([]string, len(o.buf))
	copy(res, o.buf)
	return res
}
