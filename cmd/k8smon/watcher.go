package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/frohwerk/deputy-backend/internal/task"
)

func k8swatcher(platform string) task.Task {
	watcher, err := executable("k8swatcher")
	if err != nil {
		log.Fatal("Failed to find executable k8swatcher: %s", err)
	}

	t := task.CreateTask(platform, log, func(cancel <-chan interface{}) error {
		cmd := exec.Command(watcher, platform)
		cmd.Stdout = &prefixer{Prefix: fmt.Sprintf("[%s] ", platform), Writer: os.Stdout}
		cmd.Stderr = &prefixer{Prefix: fmt.Sprintf("[%s] ", platform), Writer: os.Stderr}

		result := make(chan error)
		go func() { result <- cmd.Run() }()
		log.Debug("Starting %s %s", watcher, platform)

		select {
		case err := <-result:
			return err
		case <-cancel:
			return fmt.Errorf("task canceled")
		}
	})

	return t
}
