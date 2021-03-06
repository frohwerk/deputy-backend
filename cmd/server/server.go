package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/frohwerk/deputy-backend/cmd/server/apps"
	"github.com/frohwerk/deputy-backend/cmd/server/components"
	"github.com/frohwerk/deputy-backend/cmd/server/deployments"
	"github.com/frohwerk/deputy-backend/cmd/server/envs"
	"github.com/frohwerk/deputy-backend/cmd/server/platforms"

	"github.com/frohwerk/deputy-backend/internal"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"

	"github.com/go-chi/chi"
	"github.com/spf13/cobra"

	_ "github.com/lib/pq"
)

var (
	command = &cobra.Command{Run: Run}

	rtbase string
	port   int

	k8s internal.Platform
)

func init() {
	command.Flags().IntVarP(&port, "port", "p", 8080, "port this webhook will listen on")
}

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

func flush(resp http.ResponseWriter) bool {
	if f, ok := resp.(http.Flusher); ok {
		f.Flush()
		return true
	} else {
		log.Printf("Error flushing response")
		return false
	}
}

func main() {
	// todos.print()
	fmt.Println("starting...")
	command.Use = os.Args[0]
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}
}

func Run(cmd *cobra.Command, args []string) {
	fmt.Println("database.Open()")
	db := database.Open()
	defer util.Close(db, log.Fatalf)

	as := database.NewAppStore(db)
	cs := database.NewComponentStore(db)
	ds := database.NewDeploymentStore(db)
	es := database.NewEnvStore(db)
	ps := database.NewPlatformStore(db)

	ah := apps.NewHandler(db, as, cs, ds, es, ps)
	ch := components.NewHandler(db, cs, ds)
	dh := deployments.NewHandler(db)

	mux := chi.NewRouter()
	mux.Route("/api/apps", func(r chi.Router) {
		r.Get("/", apps.List(as))
		r.Post("/", apps.Create(as))
		r.Route("/{app}", func(r chi.Router) {
			r.Use(appParameter)
			r.Put("/components", apps.UpdateComponents(as))
			r.Get("/", ah.Get)
			r.Patch("/", ah.Patch)
			r.Delete("/", apps.Delete(as))
		})
	})

	mux.Route("/api/components", func(r chi.Router) {
		r.Get("/", ch.List)
		r.Route("/{component}/dependencies", func(r chi.Router) {
			r.Get("/", ch.GetDependencies)
			r.Patch("/", ch.PatchDependencies)
		})
	})

	mux.Route("/api/deployments", dh.Routes)

	mux.Route("/api/envs", func(r chi.Router) {
		r.Get("/", envs.List(es))
		r.Get("/{env}", envs.Get(es))
		r.Post("/", envs.Create(es))
		r.Put("/{env}", envs.Update(es))
		r.Patch("/{env}", envs.Patch(db))
		r.Delete("/{env}", envs.Delete(es))
		r.Route("/{env}/platforms", func(r chi.Router) {
			r.Get("/", platforms.List(ps))
			r.Post("/", platforms.Create(ps))
			r.Get("/{platform}", platforms.Get(ps))
			r.Put("/{platform}", platforms.Update(ps))
			r.Delete("/{platform}", platforms.Delete(ps))
		})
	})
	mux.Get("/stream", stream)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
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

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}
}
