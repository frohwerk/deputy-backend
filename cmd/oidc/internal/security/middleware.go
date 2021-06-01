package security

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

type middleware struct {
	config *oauth2.Config
	next   http.Handler
}

func NewDecorator(config *oauth2.Config, next http.Handler) http.Handler {
	return &middleware{config, next}
}

func (m *middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	token, err := GetToken(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get oauth token from cookie", err)
		rw.Write([]byte(`<!DOCTYPE html><html><head></head><body><p>Hi!<p><a href="/login">Login with Keycloak</a></body></html>`))
		return
	}

	if !token.Valid() {
		token, err = m.config.TokenSource(r.Context(), token).Token()

		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to refresh oauth token", err)
			rw.Write([]byte(`<!DOCTYPE html><html><head></head><body><p>Hi!<p><a href="/login">Login with Keycloak</a></body></html>`))
			return
		}

		if err := SetToken(rw, token); err != nil {
			fmt.Fprintln(os.Stderr, "failed to update token:", err)
		} else {
			fmt.Println("refresh for oauth token successful")
		}
	}

	m.next.ServeHTTP(rw, r)
}
