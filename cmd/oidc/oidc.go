package main

import (
	"context"
	"fmt"
	"net"
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
	// server configuration
	serverAddr string
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
	// TODOs: Find a smarter way to do this...
	overrides = map[string]string{
		"http://keycloak.myproject.svc:8080": "https://keycloak-myproject.192.168.178.31.nip.io",
	}
)

var (
	Log logger.Logger = logger.Basic(logger.LEVEL_TRACE)
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
	serverAddr = Getenv("SERVER_ADDR", ":443")
	frontendRoute = Getenv("FRONTEND_URI", "http://localhost:4200")
	backendRoute = Getenv("BACKEND_URI", "http://localhost:8080")
	tasksRoute = Getenv("TASKS_URI", "http://localhost:8877")

	providerUri = Getenv("OPENID_PROVIDER", "https://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo")
	clientId = Getenv("OPENID_CLIENT_ID", "test")
	clientSecret = Getenv("OPENID_CLIENT_SECRET", "1d319ad6-cd77-48a6-af69-4d18ab28394a")
	redirectUrl = Getenv("OPENID_REDIRECT_URI", "https://127.0.0.1.nip.io/auth/keycloak/callback")

	serverKey = Getenv("TLS_SERVER_KEY", "certificates/127.0.0.1.nip.io/key.pem")
	serverCert = Getenv("TLS_SERVER_CERT", "certificates/127.0.0.1.nip.io/cert.pem")
}

func main() {
	if addrs, err := net.LookupHost("keycloak.myproject.svc"); err != nil {
		Log.Warn("DNS lookup failed: %s", addrs)
	} else {
		for _, addr := range addrs {
			fmt.Println(addr)
		}
	}

	Log.Warn(">>>> TODO: Use key and cert files provided by openshift")

	Log.Trace("ctx := context.Background()")
	ctx := context.Background()
	Log.Trace("defer func() { time.Sleep(500 * time.Millisecond) }()")
	defer func() { time.Sleep(500 * time.Millisecond) }()

	Log.Trace("wd, _ := os.Getwd()")
	wd, _ := os.Getwd()
	Log.Info("Working directory: %s", wd)

	mux := chi.NewMux()
	app := new(ServerApplication)
	app.Addr = serverAddr
	app.Handler = mux

	frontend := reverseProxy(frontendRoute)
	frontend.Transport = http.DefaultTransport
	backend := reverseProxy(backendRoute)
	taskExecutor := reverseProxy(tasksRoute)

	Log.Info("Connecting to OpenID provider at %s", providerUri)
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

	Log.Info("Loading OpenID provider certificates from %s/protocol/openid-connect/certs", providerUri)
	jwks := oidc.NewRemoteKeySet(ctx, fmt.Sprintf("%s/protocol/openid-connect/certs", providerUri))

	Log.Info("Configuring enpdoints...")
	mux.Handle("/login", keycloak.NewLoginHandler(&config, overrides))
	mux.Handle("/logout", keycloak.NewLogoutHandler(providerUri, overrides))
	mux.Handle("/auth/keycloak/callback", keycloak.NewCallbackHandler(&config))

	mux.Handle("/whoami", whoami.NewHandler(&config, jwks))

	mux.Handle("/*", security.NewDecorator(&config, frontend))
	mux.Handle("/api/*", security.NewDecorator(&config, backend))
	mux.Handle("/api/tasks/*", security.NewDecorator(&config, taskExecutor))

	Log.Info("Starting application...")
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
	go func() {
		Log.Trace("server.ListenAndServeTLS('%s', '%s')", serverCert, serverKey)
		err := server.ListenAndServeTLS(serverCert, serverKey)
		switch {
		case err == nil:
			// Do nothing
		case err == http.ErrServerClosed:
			Log.Info("server closed: %s", err)
		default:
			Log.Info("error in server goroutine: %s", err)
		}
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
