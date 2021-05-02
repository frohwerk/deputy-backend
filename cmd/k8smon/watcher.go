package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type watcher struct {
	Id string
	*exec.Cmd
}

func k8swatcher(platform string) watcher {
	path, err := os.Executable()
	if err != nil {
		log.Error("error looking up executable path:", err)
		os.Exit(1)
	}
	dir := filepath.Dir(path)
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	p := fmt.Sprintf("%s%ck8swatcher%s", dir, os.PathSeparator, suffix)

	cmd := exec.Command(p, platform)
	cmd.Stdout = &prefixer{Prefix: fmt.Sprintf("[%s] ", platform), Writer: os.Stdout}
	cmd.Stderr = &prefixer{Prefix: fmt.Sprintf("[%s] ", platform), Writer: os.Stderr}

	log.Debug("Starting %s %s", p, platform)

	if err := cmd.Start(); err != nil {
		log.Error("Error starting:", err)
	}

	return watcher{Id: platform, Cmd: cmd}
}
