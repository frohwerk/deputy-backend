package main

import (
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/notify"
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

	listener := notify.NewListener()
	notifications := make(chan string)
	if err = listener.Listen("platforms", func(payload string) { notifications <- payload }); err != nil {
		log.Warn("error starting listener for configuration updates: %s", err)
	}

	log.Debug("Number of platforms: %d", len(platforms))
	ctrl := &controller{tasks: make(map[string]task.Task)}
	for _, p := range platforms {
		ctrl.Start(p.Id, k8swatcher(p.Id))
	}

	ctrl.StartAll()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		select {
		case payload := <-notifications:
			log.Trace("NOTIFY on platforms: '%s'", payload)
			switch op, id := payload[0:1], payload[2:]; op {
			case "I":
				log.Info("New platform added: %s", id)
				ctrl.Start(id, k8swatcher(id))
			case "U":
				log.Info("Platform modified: %s", id)
				ctrl.Restart(id)
			case "D":
				log.Info("Platform removed: %s", id)
				ctrl.Remove(id)
			}
		case sig := <-signals:
			switch sig {
			case os.Interrupt:
				ctrl.Close()
				ctrl.StopAll()
				ctrl.WaitAll()
				listener.Close()
				log.Info("Bye!")
				return
			case os.Kill:
				log.Info("AAAAAARRRGGH!")
				return
			default:
				log.Warn("Received unexpected signal: %v\n", sig)
			}
		}
	}
}
