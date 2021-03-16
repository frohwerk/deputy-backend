package filesystem_test

import (
	"bytes"
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/fsexample/filesystem"
)

func TestPositiveResult(t *testing.T) {
	ofs := filesystem.FileDigests{
		"hash-1": []string{"/app/app.js"},
	}
	ifs := filesystem.FileDigests{
		"hash-1": []string{"app.js"},
	}
	if !ofs.Contains(ifs) {
		t.Error("The outer file system should contain the inner file system")
	}
}

func TestNegativeResult(t *testing.T) {
	ofs := filesystem.FileDigests{
		"hash-1": []string{"/app/app.js"},
	}
	ifs := filesystem.FileDigests{
		"hash-2": []string{"app.js"},
	}
	if ofs.Contains(ifs) {
		t.Error("The outer file system should NOT contain the inner file system due to different hash values for app.js")
	}
}

func TestPathMismatch(t *testing.T) {
	// log.SetOutput(&logger{t.Logf})
	ofs := filesystem.FileDigests{
		"hash-1": []string{"/app/app.js"},
		"hash-2": []string{"/etc/lib.js"},
	}
	ifs := filesystem.FileDigests{
		"hash-1": []string{"app.js"},
		"hash-2": []string{"lib.js"},
	}
	if ofs.Contains(ifs) {
		t.Error("The outer file system should NOT contain the inner file system due to different paths")
	}
}
func TestDuplicateFile(t *testing.T) {
	// log.SetOutput(&logger{t.Logf})
	ofs := filesystem.FileDigests{
		"hash-1": []string{"/app/app.js"},
		"hash-2": []string{"/etc/lib.js", "/app/lib.js"},
	}
	ifs := filesystem.FileDigests{
		"hash-1": []string{"app.js"},
		"hash-2": []string{"lib.js"},
	}
	if !ofs.Contains(ifs) {
		t.Error("The outer file system should contain the inner file system")
	}
}

type logger struct {
	logfunc func(format string, args ...interface{})
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.logfunc(bytes.NewBuffer(p).String())
	return len(p), nil
}
