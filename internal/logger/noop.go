package logger

import (
	"fmt"
	"io"
	"os"
)

const (
	Noop = noop("")
)

type noop string

func (l noop) Pattern(p string) {
	// Do nothing
}

func (l noop) Trace(format string, args ...interface{}) {
	// Do nothing
}

func (l noop) Debug(format string, args ...interface{}) {
	// Do nothing
}

func (l noop) Info(format string, args ...interface{}) {
	// Do nothing
}

func (l noop) Warn(format string, args ...interface{}) {
	// Do nothing
}

func (l noop) Error(format string, args ...interface{}) {
	// Do nothing
}

func (l noop) log(level logLevel, format string, args ...interface{}) {
	// Do nothing
}

func (l noop) Writer(level logLevel) io.Writer {
	return &logWriter{l, level}
}

func (l noop) Fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", format), args...)
	os.Exit(1)
}
