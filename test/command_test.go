package test

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

type thingy struct {
}

func (t *thingy) Write() {

}

// docker compose up --build --quiet-pull --force-recreate --renew-anon-volumes --detach database

func TestInvalidCommand(t *testing.T) {
	cmd := exec.Command("ducker", "--help")
	cmd.Stdout = logger.Writer(logger.LEVEL_INFO)
	cmd.Stderr = logger.Writer(logger.LEVEL_ERROR)
	err := errors.Unwrap(cmd.Start())
	if err != nil {
		if !assert.Equal(t, exec.ErrNotFound, err) {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Println("This command does not exist. Obviously")
		}
	}
}

func TestDockerCommand(t *testing.T) {
	ready := make(chan interface{})
	markers := []string{
		`listening on IPv4 address "0.0.0.0", port 5432`,
		`database system is ready to accept connections`,
	}

	cmd := exec.Command("docker", "compose", "up", "--build", "--quiet-pull", "--force-recreate", "--renew-anon-volumes", "database")

	output := make(chan string, 10)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal("failed to connect to stdout of process:", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal("failed to connect to stderr of process:", err)
	}

	done := make(chan interface{}, 1)
	if err := cmd.Start(); err != nil {
		logger.Default.Error("error starting command: %s", err)
		return
	}

	go func() {
		r := bufio.NewReader(stdout)
		line, err := r.ReadString('\n')
		for ; err == nil; line, err = r.ReadString('\n') {
			output <- fmt.Sprintf("stdout: %s", strings.TrimSpace(string(line)))
		}
	}()

	go func() {
		r := bufio.NewReader(stderr)
		line, err := r.ReadString('\n')
		for ; err == nil; line, err = r.ReadString('\n') {
			output <- fmt.Sprintf("stderr: %s", strings.TrimSpace(string(line)))
		}
	}()

	go func() {
		for s := range output {
			if len(markers) > 0 {
				if strings.HasSuffix(s, markers[0]) {
					markers = markers[1:]
				}
				if len(markers) == 0 {
					ready <- nil
				}
			}
			fmt.Println(s)
		}
	}()

	go func() {
		r := bufio.NewReader(stderr)
		line, err := r.ReadString('\n')
		for ; err == nil; line, err = r.ReadString('\n') {
			s := strings.TrimSpace(string(line))
			if strings.Contains(s, "database system is ready to accept connections") {
				fmt.Printf(">>>%s<<<\n", s)
				logger.Default.Info("BINGO!!!")
			}
			fmt.Println(line)
		}
	}()

	go func() {
		logger.Default.Info("cmd.Wait()")
		cmd.Wait()
		done <- nil
	}()

	timeout := time.NewTimer(15 * time.Second)
	select {
	case <-ready:
		logger.Default.Info("Database is ready. YAY!")
		timeout.Stop()
	case <-timeout.C:
		logger.Default.Warn("Database did not get ready in time...")
	}

	logger.Default.Info("Sending SIGINT")
	cmd.Process.Signal(os.Interrupt)

	timeout = time.NewTimer(10 * time.Second)
	select {
	case <-done:
		timeout.Stop()
	case <-timeout.C:
		logger.Default.Info("Sending SIGTERM")
		cmd.Process.Signal(os.Kill)
		cmd.Wait()
		logger.Default.Info("done waiting")
	}

	logger.Default.Info("The End(?)")
}
