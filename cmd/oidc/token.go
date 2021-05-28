package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
)

func SetToken(rw http.ResponseWriter, token *oauth2.Token) error {
	buf, err := json.Marshal(token)
	if err != nil {
		return err
	}

	http.SetCookie(rw, &http.Cookie{Name: "token", Value: base64.StdEncoding.EncodeToString(buf), Path: "/", HttpOnly: true})

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
