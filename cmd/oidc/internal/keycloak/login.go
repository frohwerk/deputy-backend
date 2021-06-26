package keycloak

import (
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type loginHandler struct {
	config    *oauth2.Config
	overrides map[string]string
}

func NewLoginHandler(config *oauth2.Config, overrides map[string]string) http.Handler {
	return &loginHandler{config, overrides}
}

func (l loginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	loginUri := l.config.AuthCodeURL("")
	for baseUri, replacement := range l.overrides {
		loginUri = strings.Replace(loginUri, baseUri, replacement, 1)
	}
	http.Redirect(rw, r, loginUri, http.StatusFound)
}
