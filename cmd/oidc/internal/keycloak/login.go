package keycloak

import (
	"net/http"

	"golang.org/x/oauth2"
)

type loginHandler struct {
	config *oauth2.Config
}

func NewLoginHandler(config *oauth2.Config) http.Handler {
	return &loginHandler{config}
}

func (l loginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, l.config.AuthCodeURL(""), http.StatusFound)
}
