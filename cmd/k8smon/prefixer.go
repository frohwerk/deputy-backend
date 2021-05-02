package main

import "io"

type prefixer struct {
	Prefix string
	io.Writer
}

func (t *prefixer) Write(p []byte) (n int, err error) {
	t.Writer.Write([]byte(t.Prefix))
	return t.Writer.Write(p)
}
