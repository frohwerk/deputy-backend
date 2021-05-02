package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/task"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log := logger.WithPrefix("[workshop] ", logger.LEVEL_TRACE)
	w := &worker{Logger: log}

	tasks := []task.Task{task.CreateTask("1", log, w.work), task.CreateTask("2", log, w.work), task.CreateTask("3", log, w.work)}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	task.StartAll(tasks)

	for {
		switch s := <-sigs; s {
		case os.Interrupt:
			task.StopAll(tasks)
			task.WaitAll(tasks)
			os.Exit(0)
		case os.Kill:
			log.Info("SIGTERM")
			os.Exit(0)
		}
	}
}

type worker struct {
	logger.Logger
}

func (w *worker) work(cancel <-chan interface{}) error {
	if rand.Intn(1) > 0 {
		return fmt.Errorf("crash on startup")
	}
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop()
	for i := 0; i < rand.Intn(15)+1; i++ {
		select {
		case <-cancel:
			return nil
		case <-ticker.C:
		}
		w.Logger.Trace("working hard: %d", i)
		if rand.Intn(15) > 13 {
			return fmt.Errorf("crash during work")
		}
	}
	return nil
}
