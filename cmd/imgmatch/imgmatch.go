package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/frohwerk/deputy-backend/cmd/imgmatch/handler"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/matcher"
	"github.com/frohwerk/deputy-backend/cmd/server/images"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/spf13/cobra"
)

var (
	imgmatch = &cobra.Command{RunE: run}
	port     int
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
	imgmatch.Flags().IntVarP(&port, "port", "p", 8092, "port number the server process will listen on")
	imgmatch.Flags().StringVarP(&registry, "registry", "r", "", "The base uri of the docker container registry to use")
}

func main() {
	imgmatch.Use = os.Args[0]
	if err := imgmatch.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	db := database.Open()
	defer db.Close()

	if v := os.Getenv("SERVER_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err != nil {
			port = i
		}
	}
	if v := os.Getenv("REGISTRY_BASE_URL"); v != "" {
		registry = v
	}

	reg := &images.RemoteRegistry{BaseUrl: registry}
	fs := database.NewFileStore(db)
	m := matcher.New(fs, fs, reg)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: handler.New(m, database.NewImageStore(db)),
	}

	go func() { server.ListenAndServe() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		fmt.Println("Server up on port", port)
		switch sig := <-signals; sig {
		case os.Interrupt:
			server.Shutdown(context.Background())
			return nil
		case os.Kill:
			server.Close()
			return nil
		}
	}
}
