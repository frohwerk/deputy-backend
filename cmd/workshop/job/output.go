package job

import (
	"fmt"
)

type Output interface {
	Write(format string, args ...interface{})
}

type OutputBuffer struct {
	buf []string
}

func (o *OutputBuffer) Write(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	defer func() { fmt.Println("o.buf (after): ", o.buf) }()
	fmt.Println("OutputBuffer::Write =>", s)
	fmt.Println("o.buf (before):", o.buf)
	o.buf = append(o.buf, s)
}

func (o *OutputBuffer) Get() []string {
	fmt.Println("OutputBuffer::Get")
	fmt.Println("o.buf", o.buf)
	res := make([]string, len(o.buf))
	copy(res, o.buf)
	return res
}
