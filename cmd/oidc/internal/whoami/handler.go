package whoami

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/frohwerk/deputy-backend/cmd/oidc/internal/security"
	"golang.org/x/oauth2"
)

type handler struct {
	config *oauth2.Config
	jwks   oidc.KeySet
}

func NewHandler(config *oauth2.Config, keySet oidc.KeySet) http.Handler {
	return &handler{config, keySet}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	oauthToken, err := security.GetToken(r)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	}

	if !oauthToken.Valid() {
		oauthToken, err = h.config.TokenSource(r.Context(), oauthToken).Token()
		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}
		if err := security.SetToken(rw, oauthToken); err != nil {
			rw.Write([]byte(fmt.Sprintln(err)))
			return
		}
	}

	rw.Write([]byte(`<html><head></head><body><pre>`))

	rw.Write([]byte(fmt.Sprintf("Token expires: %v", oauthToken.Expiry)))
	rw.Write([]byte("\n\n"))
	rw.Write([]byte(oauthToken.AccessToken))
	rw.Write([]byte("\n\n"))
	rw.Write([]byte(oauthToken.RefreshToken))
	rw.Write([]byte("\n\n"))

	payload, err := h.jwks.VerifySignature(r.Context(), oauthToken.AccessToken)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}

	rw.Write([]byte(payload))
	rw.Write([]byte("\n\n"))

	claims := &security.Claims{}
	if err := json.Unmarshal(payload, claims); err != nil {
		rw.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}

	rw.Write([]byte(fmt.Sprintf("ID: %s\n", claims.ID)))
	rw.Write([]byte(fmt.Sprintf("Issuer: %s\n", claims.Issuer)))
	rw.Write([]byte(fmt.Sprintf("Subject: %s\n", claims.Subject)))
	rw.Write([]byte(fmt.Sprintf("Audience: %s\n", claims.Audience)))
	rw.Write([]byte(fmt.Sprintf("IssuedAt: %v\n", time.Unix(int64(*claims.IssuedAt), 0))))
	rw.Write([]byte(fmt.Sprintf("Expiry: %v\n", time.Unix(int64(*claims.Expiry), 0))))
	rw.Write([]byte(fmt.Sprintf("PreferredUsername: %s\n", claims.PreferredUsername)))

	rw.Write([]byte(`</pre><p><a href="/logout">Logout</a></body></html>`))
}
