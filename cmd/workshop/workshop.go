package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/server/envs"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mock struct{}

func (m mock) Create(name string) (*database.Env, error) {
	return &database.Env{Id: uuid.NewString(), Name: name}, nil
}

func main() {
	mux := chi.NewRouter()
	mux.Route("/api/envs", func(r chi.Router) {
		r.Post("/", envs.Create(new(mock)))
	})
	server := http.Server{Addr: "localhost:8099", Handler: mux}

	go func() {
		err := server.ListenAndServe()
		log.Printf("Server stopped listening: %s", err)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	for {
		select {
		case s := <-signals:
			switch s {
			case os.Interrupt:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := server.Shutdown(ctx)
				cancel()
				if err != nil {
					log.Printf("error during server shutdown: %s", err)
				}
				return
			case os.Kill:
				err := server.Close()
				if err != nil {
					log.Printf("error during server shutdown: %s", err)
				}
				return
			}
		}
	}
}
