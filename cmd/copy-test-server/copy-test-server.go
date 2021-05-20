package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/server/tasks"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/go-chi/chi"
)

func main() {
	db := database.Open()
	defer db.Close()

	tasksHandler := tasks.CreateHandler(db)

	mux := chi.NewMux()
	mux.Route("/", tasksHandler.Routes)

	server := http.Server{
		Addr:    ":8765",
		Handler: mux,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	go func() { server.ListenAndServe() }()

	for {
		switch sig := <-signals; sig {
		case os.Interrupt:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := server.Shutdown(ctx)
			cancel()
			if err != nil {
				fmt.Println(err)
			}
			return
		case os.Kill:
			server.Close()
			return
		}
	}
}
