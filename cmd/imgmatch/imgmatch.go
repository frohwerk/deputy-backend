package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/cmd/imgmatch/handler"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/matcher"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/security"
	"github.com/frohwerk/deputy-backend/cmd/server/images"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/notify"
	"github.com/spf13/cobra"
)

var (
	imgmatch = &cobra.Command{RunE: run}

	Log logger.Logger

	registry string
)

func Getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	Log = logger.Basic(logger.LEVEL_DEBUG)
	handler.Log = Log
	matcher.Log = logger.Basic(logger.LEVEL_TRACE)
	security.Log = Log

	imgmatch.Flags().StringVarP(&registry, "registry", "r", "", "The base uri of the docker container registry to use")
}

func main() {
	imgmatch.Use = os.Args[0]
	if err := imgmatch.Execute(); err != nil {
		Log.Error("%s", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	db := database.Open()
	defer db.Close()

	if v := os.Getenv("REGISTRY_BASE_URL"); registry == "" {
		registry = v
	}

	var tr http.RoundTripper = http.DefaultTransport
	if buf, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		tr = security.BearerTokenAuthorization(tr, string(buf))
	} else if os.ErrNotExist != err {
		Log.Warn("No serviceaccount token found. Probably running on a local machine...")
	}

	reg := &images.RemoteRegistry{BaseUrl: registry, Transport: tr}

	fs := database.NewFileStore(db)
	m := matcher.New(fs, fs, reg)

	h, l := handler.New(m, database.NewImageStore(db)), notify.NewListener()
	err := l.Listen("images", func(image string) { go h.Accept(image) })
	if err != nil {
		Log.Fatal("error attaching listener to channel 'images': %s", err)
	} else {
		Log.Info("Listening on channel 'images'")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		switch sig := <-signals; sig {
		case os.Interrupt:
			l.Close()
			return nil
		case os.Kill:
			return nil
		}
	}
}
