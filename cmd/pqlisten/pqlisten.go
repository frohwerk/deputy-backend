package main

import (
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/notify"
)

var Log logger.Logger = logger.Basic(logger.LEVEL_DEBUG)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	listener := notify.NewListener()
	defer listener.Close()

	err := listener.Listen("platforms", func(s string) { Log.Info("Incoming on channel platforms: %s", s) })
	if err != nil {
		Log.Fatal("%s", err)
	}

	<-sigs
	return
}
