package main

import "gopkg.in/square/go-jose.v2/jwt"

type Claims struct {
	jwt.Claims
	PreferredUsername string `json:"preferred_username,omitempty"`
}
