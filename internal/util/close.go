package util

import (
	"io"
)

func Close(c io.Closer, log func(string, ...interface{})) {
	if err := c.Close(); err != nil {
		log("error closing rows: %s\n", err)
	}
}
