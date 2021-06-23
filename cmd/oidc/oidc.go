package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/frohwerk/deputy-backend/cmd/oidc/internal/keycloak"
	"github.com/frohwerk/deputy-backend/cmd/oidc/internal/security"
	"github.com/frohwerk/deputy-backend/cmd/oidc/internal/whoami"
	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
)

var (
	keys = map[string]string{
		"7tvjtgKw6v8dSYBT2xu433-0g-aEngP_NSYCfsLUpV4": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvmMDe4TWD0nEea7osM+EC60ucOZRl5RwrN3iWZxeVMX2jnRLIea5upTAEEEMsArrOyxSYaJq6JtixGAnogAZEttzxNo/Ccp7CWqYWKSq7sE7fKEMCKksllpXuYTrf1AoMVt+J3v3YTTfJX4W35PhP5fkp7bMv2VthDgc8x7HNOPgNrde+aBKL5BaRkDr6azhcQCvYEf+l6mQIN+Wnv+LGwJX3N/5KBpPmySOdgRRPthUg9FS1BS2eEiGiu2q5ce5hCALX+jZIoq224GG9ZQInJ+RoSHwzv8JzBFTFlTI9hXmd6urZomSVeYFPsWUr5ppYE/51K4sDsIt2wRXqi/nawIDAQAB",
	}
)

type ServerApplication struct {
	http.Server
}

func main() {
	ctx := context.Background()
	defer func() { time.Sleep(500 * time.Millisecond) }()

	wd, _ := os.Getwd()
	fmt.Println("Working directory: ", wd)

	mux := chi.NewMux()
	app := new(ServerApplication)
	app.Addr = ":443"
	app.Handler = mux

	fmt.Println("TODO: add error handling for reverse proxy target host")
	frontendRoute, _ := url.Parse("http://localhost:4200")
	frontend := httputil.NewSingleHostReverseProxy(frontendRoute)
	backendRoute, _ := url.Parse("http://localhost:8080")
	backend := httputil.NewSingleHostReverseProxy(backendRoute)
	tasksRoute, _ := url.Parse("http://localhost:8877")
	taskExecutor := httputil.NewSingleHostReverseProxy(tasksRoute)

	provider, err := oidc.NewProvider(ctx, "https://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Scanln()
		return
	}

	// oidcconfig := oidc.Config{
	// 	ClientID: "test",
	// }
	config := oauth2.Config{
		ClientID:     "test",
		ClientSecret: "1d319ad6-cd77-48a6-af69-4d18ab28394a",
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "https://127.0.0.1.nip.io/auth/keycloak/callback",
		Scopes:       []string{oidc.ScopeOpenID},
	}
	config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader

	jwks := oidc.NewRemoteKeySet(ctx, "https://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo/protocol/openid-connect/certs")

	mux.Handle("/login", keycloak.NewLoginHandler(&config))
	mux.Handle("/logout", keycloak.NewLogoutHandler("https://keycloak-myproject.192.168.178.31.nip.io", "demo"))
	mux.Handle("/auth/keycloak/callback", keycloak.NewCallbackHandler(&config))

	mux.Handle("/whoami", whoami.NewHandler(&config, jwks))

	mux.Handle("/*", security.NewDecorator(&config, frontend))
	mux.Handle("/api/*", security.NewDecorator(&config, backend))
	mux.Handle("/api/tasks/*", security.NewDecorator(&config, taskExecutor))

	app.start()

	app.awaitShutdown()
}

func (server *ServerApplication) start() {
	go func() {
		server.ListenAndServeTLS("certificates/127.0.0.1.nip.io/cert.pem", "certificates/127.0.0.1.nip.io/key.pem")
	}()
}

func (server *ServerApplication) awaitShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	for {
		switch sig := <-sigs; sig {
		case os.Interrupt:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			fmt.Println("Shutting down server...")
			server.Shutdown(ctx)
			cancel()
			fmt.Println("Shutting complete")
			return
		case os.Kill:
			fmt.Println("Killing server")
			err := server.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				time.Sleep(1 * time.Second)
				os.Exit(1)
			}
			return
		}
	}
}
