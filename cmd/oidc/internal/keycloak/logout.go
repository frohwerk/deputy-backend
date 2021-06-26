package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type logoutHandler struct {
	BaseUri   string
	overrides map[string]string
}

func NewLogoutHandler(baseUri string, overrides map[string]string) http.Handler {
	return &logoutHandler{baseUri, overrides}
}

func (l logoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	http.SetCookie(rw, &http.Cookie{Name: "token", Expires: time.Unix(0, 0)})
	http.SetCookie(rw, &http.Cookie{Name: "claims", Expires: time.Unix(0, 0)})
	redirectUri := url.QueryEscape(fmt.Sprintf("https://%s", r.Host))
	logoutUri := fmt.Sprintf(`%s/protocol/openid-connect/logout?redirect_uri=%s`, l.BaseUri, redirectUri)
	for baseUri, replacement := range l.overrides {
		logoutUri = strings.Replace(logoutUri, baseUri, replacement, 1)
	}
	http.Redirect(rw, r, logoutUri, http.StatusFound)
}
