package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func executable(name string) (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(path)
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	return fmt.Sprintf("%s%c%s%s", dir, os.PathSeparator, name, suffix), nil
}
