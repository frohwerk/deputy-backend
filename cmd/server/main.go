package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/cmd/server/apps"
	"github.com/frohwerk/deputy-backend/cmd/server/components"

	"github.com/frohwerk/deputy-backend/internal"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"

	"github.com/go-chi/chi"

	_ "github.com/lib/pq"
)

type response struct {
	EventType string        `json:"eventType"`
	Object    api.Component `json:"object"`
}

func appParameter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(resp, req.WithContext(context.WithValue(req.Context(), params.App, chi.URLParam(req, string(params.App)))))
	})
}

func stream(resp http.ResponseWriter, req *http.Request) {
	encoder := json.NewEncoder(resp)

	for i := 1; i < 6; i++ {
		item := response{
			EventType: "ADDED",
			Object:    api.Component{Name: fmt.Sprintf("loop #%v", i)},
		}

		if err := encoder.Encode(item); err != nil {
			log.Printf("Error sending chunk %v: %v", i, err)
		}

		if !flush(resp) {
			return
		}
	}
}

func getComponents(resp http.ResponseWriter, req *http.Request) {
	id := req.Context().Value(params.App)
	fmt.Printf("Looking up components for app '%s'\n", id)

	if observable, err := k8s.WatchComponents(); err != nil {
		log.Printf("error watching artifacts on kubernetes: %v\n", err)
	} else {
		// When the request is canceled or completed stop watching
		go func() { <-req.Context().Done(); observable.Stop() }()
		// Create json encoder for response
		enc := json.NewEncoder(resp)
		// Watch for events and send them encoded as json to the client
		for event := range observable.Events {
			if err := enc.Encode(response{EventType: event.EventType, Object: event.Object}); err != nil {
				log.Printf("error encoding artifact: %v", err)
			}
			flush(resp)
		}
	}
}

func flush(resp http.ResponseWriter) bool {
	if f, ok := resp.(http.Flusher); ok {
		f.Flush()
		return true
	} else {
		log.Printf("Error flushing response")
		return false
	}
}

var k8s internal.Platform

func main() {
	k, err := kubernetes.WithDefaultConfig()
	if err != nil {
		log.Fatalf("error loading kubernetes configuration: %v\n", err)
	}
	k8s = k

	db, err := sql.Open("postgres", "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer util.Close(db, log.Fatalf)

	as := database.NewAppStore(db)
	cs := database.NewComponentStore(db)

	mux := chi.NewRouter()
	mux.Route("/api/apps", func(r chi.Router) {
		r.Get("/", apps.List(as))
		r.Post("/", apps.Create(as))
		r.Route("/{app}", func(r chi.Router) {
			r.Use(appParameter)
			r.Get("/artifacts", getComponents) // TODO Remove deprecated endpoint after frontend update
			r.Get("/components", getComponents)
			r.Get("/", apps.Get(as, cs))
		})
	})
	mux.Route("/api/components", func(r chi.Router) {
		r.Get("/", components.List(cs))
	})
	mux.Get("/stream", stream)

	server := http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	errors := make(chan error, 1)
	go func() { errors <- server.ListenAndServe() }()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case err := <-errors:
		log.Printf("failed to serve: %v", err)
	case sig := <-interrupt:
		log.Printf("terminating due to os signal: %v", sig)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}
}
