package keycloak

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"

	"github.com/frohwerk/deputy-backend/cmd/oidc/internal/security"
	"golang.org/x/oauth2"
)

func NewCallbackHandler(config *oauth2.Config) http.Handler {
	return &callback{config, "/"}
}

type callback struct {
	config      *oauth2.Config
	RedirectUri string
}

func (c *callback) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := httptrace.WithClientTrace(r.Context(), &httptrace.ClientTrace{
		GetConn:              func(hostPort string) { fmt.Printf("GetConn(%s)\n", hostPort) },
		GotFirstResponseByte: func() { fmt.Println("Got first response byte") },
		TLSHandshakeStart:    func() { fmt.Println("TLS Handshake start") },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("TLS Handshake done")
			}
		},
		WroteHeaderField: func(key string, value []string) { fmt.Printf("Wrote header field %v: %v \n", key, value) },
		WroteRequest:     func(wri httptrace.WroteRequestInfo) { fmt.Println("Wrote request:", wri.Err) },
		WroteHeaders:     func() { fmt.Println("Wrote headers") },
	})

	code := r.URL.Query().Get("code")

	// rw.Write([]byte(code))
	// if true {
	// 	return
	// }

	oauthToken, err := c.config.Exchange(ctx, code)
	if err != nil {
		rw.Write([]byte(fmt.Sprintln(err)))
		return
	}

	if err := security.SetToken(rw, oauthToken); err != nil {
		rw.Write([]byte(fmt.Sprintln(err)))
		return
	}

	http.Redirect(rw, r, c.RedirectUri, http.StatusFound)
}
