package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type logLevel int

const (
	LEVEL_TRACE logLevel = iota
	LEVEL_DEBUG logLevel = iota
	LEVEL_INFO  logLevel = iota
	LEVEL_WARN  logLevel = iota
	LEVEL_ERROR logLevel = iota
)

var (
	levelText = []string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR"}
)

type basic struct {
	level   logLevel
	pattern string
}

type Logger interface {
	Pattern(format string)
	Writer(level logLevel) io.Writer
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

type logger interface {
	Logger
	log(level logLevel, format string, args ...interface{})
}

func Basic(l logLevel) Logger {
	return WithPrefix("", l)
}

func WithPrefix(prefix string, l logLevel) Logger {
	return &basic{l, fmt.Sprintf("%s%s %s %s\n", prefix, "%s", "%s", "%s")}
}

func (l *basic) Level(level logLevel) *basic {
	l.level = level
	return l
}

func (l *basic) Pattern(p string) {
	l.pattern = p
}

func (l *basic) Trace(format string, args ...interface{}) {
	l.log(LEVEL_TRACE, format, args...)
}

func (l *basic) Debug(format string, args ...interface{}) {
	l.log(LEVEL_DEBUG, format, args...)
}

func (l *basic) Info(format string, args ...interface{}) {
	l.log(LEVEL_INFO, format, args...)
}

func (l *basic) Warn(format string, args ...interface{}) {
	l.log(LEVEL_WARN, format, args...)
}

func (l *basic) Error(format string, args ...interface{}) {
	l.log(LEVEL_ERROR, format, args...)
}

func (l *basic) Fatal(format string, args ...interface{}) {
	l.log(LEVEL_ERROR, format, args...)
	os.Exit(1)
}

func (b *basic) Writer(level logLevel) io.Writer {
	return &logWriter{b, level}
}

func (l *basic) log(level logLevel, format string, args ...interface{}) {
	format = strings.TrimSpace(format)
	switch {
	case format == "":
		return
	case l.level <= level:
		fmt.Printf(l.pattern, time.Now().Format("2006-01-02 15:04:05.000"), level, fmt.Sprintf(format, args...))
	}
}

func (l logLevel) String() string {
	return levelText[l]
}
