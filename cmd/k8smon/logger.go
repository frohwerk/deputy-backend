package main

import (
	"fmt"
	"os"
)

type LogLevel int

const (
	LOG_DEBUG LogLevel = iota
	LOG_INFO  LogLevel = iota
	LOG_WARN  LogLevel = iota
	LOG_ERROR LogLevel = iota
)

type basicLogger struct {
	prefix string
	level  LogLevel
}

type logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

func (l basicLogger) Debug(format string, args ...interface{}) {
	l.log(LOG_DEBUG, format, args...)
}

func (l basicLogger) Info(format string, args ...interface{}) {
	l.log(LOG_INFO, format, args...)
}

func (l basicLogger) Warn(format string, args ...interface{}) {
	l.log(LOG_WARN, format, args...)
}

func (l basicLogger) Error(format string, args ...interface{}) {
	l.log(LOG_ERROR, format, args...)
}

func (l basicLogger) Fatal(format string, args ...interface{}) {
	l.log(LOG_ERROR, format, args...)
	os.Exit(1)
}

func (l basicLogger) log(level LogLevel, format string, args ...interface{}) {
	if l.level <= level {
		fmt.Printf(fmt.Sprintf("[%s] %s\n", l.prefix, format), args...)
	}
}
