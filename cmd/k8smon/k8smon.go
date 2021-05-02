package main

import (
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/internal/database"
)

var (
	log = &basicLogger{prefix: "k8smon", level: LOG_DEBUG}
)

func main() {
	db := database.Open()
	defer database.Close(db)

	ps := database.NewPlatformStore(db)

	platforms, err := ps.List()

	if err != nil {
		log.Fatal("error reading platforms: %s", err)
	}

	watchers := []watcher{}
	for _, p := range platforms {
		watchers = append(watchers, k8swatcher(p.Id))
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		switch sig := <-signals; sig {
		case os.Interrupt:
			for _, c := range watchers {
				if err := c.Wait(); err != nil {
					log.Error("error during shutdown of k8swatcher %s:", c.Id, err)
				}
			}
			os.Exit(0)
		case os.Kill:
			os.Exit(0)
		default:
			log.Warn("Received unexpected signal: %v\n", sig)
		}
	}
}
