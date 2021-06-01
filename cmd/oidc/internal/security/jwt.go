package security

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/square/go-jose.v2/jwt"
)

type Claims struct {
	jwt.Claims
	PreferredUsername string `json:"preferred_username,omitempty"`
}

func ParseClaims(token string) (*jwt.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid parameter: jwts should have at least two segments")
	}

	claims := new(jwt.Claims)
	decoder := json.NewDecoder(base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(parts[1])))
	if err := decoder.Decode(claims); err != nil {
		return nil, err
	}

	return claims, nil
}
