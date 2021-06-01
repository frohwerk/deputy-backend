package security

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var zero = jwt.NumericDate(0)

func SetToken(rw http.ResponseWriter, token *oauth2.Token) error {
	buf, err := json.Marshal(token)
	if err != nil {
		return err
	}

	claims, err := ParseClaims(token.RefreshToken)
	if err != nil {
		return err
	}

	if claims.Expiry == nil {
		claims.Expiry = &zero
	}

	segments := strings.Split(token.AccessToken, ".")
	if len(segments) != 3 {
		fmt.Fprintln(os.Stderr, "Expected 3 segments, but found", len(segments))
		return fmt.Errorf("invalid jwt: expected exactly 3 segments: header, claims and signature")
	}

	http.SetCookie(rw, &http.Cookie{Name: "claims", Value: segments[1], Path: "/", Expires: claims.Expiry.Time()})
	http.SetCookie(rw, &http.Cookie{Name: "whatever", Value: "whatever", Path: "/", Expires: claims.Expiry.Time()})
	http.SetCookie(rw, &http.Cookie{Name: "token", Value: base64.StdEncoding.EncodeToString(buf), Path: "/", Secure: true, HttpOnly: true})

	return nil
}

func GetToken(r *http.Request) (*oauth2.Token, error) {
	tc, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}

	buf, err := base64.StdEncoding.DecodeString(tc.Value)
	if err != nil {
		return nil, err
	}

	oauthToken := new(oauth2.Token)
	if err := json.Unmarshal(buf, oauthToken); err != nil {
		return nil, err
	}

	return oauthToken, nil
}
