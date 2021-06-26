package security

import (
	"fmt"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/logger"
)

var Log logger.Logger = logger.Default

type bearerTokenDecorator struct {
	base  http.RoundTripper
	token string
}

func BearerTokenAuthorization(base http.RoundTripper, token string) http.RoundTripper {
	return &bearerTokenDecorator{base, token}
}

func (d *bearerTokenDecorator) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.token))
	Log.Debug("Using Authorization: Bearer %s...", d.token[:min(11, len(d.token))])
	return d.base.RoundTrip(r)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
