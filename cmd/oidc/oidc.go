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
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
)

var (
	// Routes to other components
	frontendRoute string
	backendRoute  string
	tasksRoute    string
	// OpenID related variables
	providerUri  string
	clientId     string
	clientSecret string
	redirectUrl  string
	// TLS private key and certificate
	serverKey  string
	serverCert string
	// TODOs: Use key and cert files provided by openshift
	// Maybe a wildcard certificate will be needeed for the public route? *.my-openshift-domain
)

var (
	Log logger.Logger = logger.Basic(logger.LEVEL_INFO)
)

type ServerApplication struct {
	http.Server
}

func Getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	frontendRoute = Getenv("FRONTEND_URI", "http://localhost:4200")
	backendRoute = Getenv("BACKEND_URI", "http://localhost:8080")
	tasksRoute = Getenv("TASKS_URI", "http://localhost:8877")

	providerUri = Getenv("OPENID_PROVIDER", "https://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo")
	clientId = Getenv("OPENID_CLIENT_ID", "test")
	clientSecret = Getenv("OPENID_CLIENT_SECRET", "1d319ad6-cd77-48a6-af69-4d18ab28394a")
	redirectUrl = Getenv("OPENID_REDIRECT_URI", "https://127.0.0.1.nip.io/auth/keycloak/callback")

	serverKey = Getenv("TLS_SERVER_KEY", "certificates/127.0.0.1.nip.io/key.pem")
	serverKey = Getenv("TLS_SERVER_CERT", "certificates/127.0.0.1.nip.io/cert.pem")
}

func main() {
	Log.Warn("TODO: Use key and cert files provided by openshift")

	ctx := context.Background()
	defer func() { time.Sleep(500 * time.Millisecond) }()

	wd, _ := os.Getwd()
	fmt.Println("Working directory: ", wd)

	mux := chi.NewMux()
	app := new(ServerApplication)
	app.Addr = ":443"
	app.Handler = mux

	frontend := reverseProxy(frontendRoute)
	backend := reverseProxy(backendRoute)
	taskExecutor := reverseProxy(tasksRoute)

	provider, err := oidc.NewProvider(ctx, providerUri)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Scanln()
		return
	}

	config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       []string{oidc.ScopeOpenID},
	}
	config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader

	jwks := oidc.NewRemoteKeySet(ctx, fmt.Sprintf("%s/protocol/openid-connect/certs", providerUri))

	mux.Handle("/login", keycloak.NewLoginHandler(&config))
	mux.Handle("/logout", keycloak.NewLogoutHandler(providerUri))
	mux.Handle("/auth/keycloak/callback", keycloak.NewCallbackHandler(&config))

	mux.Handle("/whoami", whoami.NewHandler(&config, jwks))

	mux.Handle("/*", security.NewDecorator(&config, frontend))
	mux.Handle("/api/*", security.NewDecorator(&config, backend))
	mux.Handle("/api/tasks/*", security.NewDecorator(&config, taskExecutor))

	app.start()

	app.awaitShutdown()
}

func reverseProxy(route string) *httputil.ReverseProxy {
	target, err := url.Parse(route)
	if err != nil {
		Log.Fatal("error parsing url '%s': %v", route, err)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

func (server *ServerApplication) start() {
	go func() { server.ListenAndServeTLS(serverCert, serverKey) }()
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
