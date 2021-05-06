package main

import (
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/task"
)

var (
	log = logger.WithPrefix("[k8smon] ", logger.LEVEL_DEBUG)
)

func main() {
	log.Info("Starting k8smon...")
	db := database.Open()
	defer database.Close(db)

	ps := database.NewPlatformStore(db)

	log.Debug("Reading platforms...")
	platforms, err := ps.List()
	if err != nil {
		log.Fatal("error reading platforms: %s", err)
	}

	log.Debug("Number of platforms: %d", len(platforms))
	watchers := make([]task.Task, len(platforms))
	for i, p := range platforms {
		watchers[i] = k8swatcher(p.Id)
	}

	task.StartAll(watchers)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		switch sig := <-signals; sig {
		case os.Interrupt:
			task.StopAll(watchers)
			task.WaitAll(watchers)
			os.Exit(0)
		case os.Kill:
			os.Exit(0)
		default:
			log.Warn("Received unexpected signal: %v\n", sig)
		}
	}
}
