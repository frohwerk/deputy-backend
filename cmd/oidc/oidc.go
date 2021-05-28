package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/go-oidc"
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
	mux := chi.NewMux()
	app := new(ServerApplication)
	app.Addr = ":8080"
	app.Handler = mux

	provider, err := oidc.NewProvider(ctx, "http://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo")
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
		ClientSecret: "43640e4d-b00f-4f33-a9f9-edb99645ba08",
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8080/auth/keycloak/callback",
		Scopes:       []string{oidc.ScopeOpenID},
	}

	jwks := oidc.NewRemoteKeySet(ctx, "http://keycloak-myproject.192.168.178.31.nip.io/auth/realms/demo/protocol/openid-connect/certs")

	mux.Handle("/auth/keycloak/callback", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		oauthToken, err := config.Exchange(r.Context(), code)
		if err != nil {
			rw.Write([]byte(fmt.Sprintln(err)))
			return
		}

		if err := SetToken(rw, oauthToken); err != nil {
			rw.Write([]byte(fmt.Sprintln(err)))
			return
		}

		http.Redirect(rw, r, "/whoami", http.StatusFound)
	}))

	mux.Handle("/login", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, config.AuthCodeURL(""), http.StatusFound)
	}))

	mux.Handle("/logout", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.SetCookie(rw, &http.Cookie{Name: "token"})
		rw.Write([]byte(`<html><head></head><body><p>Bye!<p><a href="/login">Login with Keycloak</a></body></html>`))
	}))

	mux.Handle("/whoami", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		oauthToken, err := GetToken(r)
		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}

		if !oauthToken.Valid() {
			oauthToken, err = config.TokenSource(r.Context(), oauthToken).Token()
			if err != nil {
				rw.Write([]byte(err.Error()))
				return
			}
			if err := SetToken(rw, oauthToken); err != nil {
				rw.Write([]byte(fmt.Sprintln(err)))
				return
			}
		}

		rw.Write([]byte(fmt.Sprintf("Token expires: %v", oauthToken.Expiry)))
		rw.Write([]byte("\n\n"))
		rw.Write([]byte(oauthToken.AccessToken))
		rw.Write([]byte("\n\n"))
		rw.Write([]byte(oauthToken.RefreshToken))
		rw.Write([]byte("\n\n"))

		payload, err := jwks.VerifySignature(r.Context(), oauthToken.AccessToken)
		if err != nil {
			rw.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}

		rw.Write([]byte(payload))
		rw.Write([]byte("\n\n"))

		claims := &Claims{}
		if err := json.Unmarshal(payload, claims); err != nil {
			rw.Write([]byte(fmt.Sprintf("%v", err)))
			return
		}

		rw.Write([]byte(fmt.Sprintf("ID: %s\n", claims.ID)))
		rw.Write([]byte(fmt.Sprintf("Issuer: %s\n", claims.Issuer)))
		rw.Write([]byte(fmt.Sprintf("Subject: %s\n", claims.Subject)))
		rw.Write([]byte(fmt.Sprintf("Audience: %s\n", claims.Audience)))
		rw.Write([]byte(fmt.Sprintf("NotBefore: %v\n", claims.NotBefore)))
		rw.Write([]byte(fmt.Sprintf("IssuedAt: %v\n", time.Unix(int64(*claims.IssuedAt), 0))))
		rw.Write([]byte(fmt.Sprintf("Expiry: %v\n", time.Unix(int64(*claims.Expiry), 0))))
		rw.Write([]byte(fmt.Sprintf("PreferredUsername: %s\n", claims.PreferredUsername)))
	}))

	mux.Handle("/*", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(fmt.Sprintln(r.Method, r.URL, r.Proto)))
	}))

	app.start()

	app.awaitShutdown()
}

func (server *ServerApplication) start() {
	go func() { server.ListenAndServe() }()
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
