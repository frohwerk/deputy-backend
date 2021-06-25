package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type logoutHandler struct {
	BaseUri string
}

func NewLogoutHandler(BaseUri string) http.Handler {
	return &logoutHandler{BaseUri}
}

func (l logoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	http.SetCookie(rw, &http.Cookie{Name: "token", Expires: time.Unix(0, 0)})
	http.SetCookie(rw, &http.Cookie{Name: "claims", Expires: time.Unix(0, 0)})
	redirectUri := url.QueryEscape(fmt.Sprintf("https://%s", r.Host))
	logoutUri := fmt.Sprintf(`%s/protocol/openid-connect/logout?redirect_uri=%s`, l.BaseUri, redirectUri)
	http.Redirect(rw, r, logoutUri, http.StatusFound)
}
