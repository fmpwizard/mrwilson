package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// createToken generates a crypto safe random string to use as auth token Bearer.
// When running on production mode, the app will sms the token to the `config.TokenTo` number
// when running on any other mode, we log the token to stdout
func createToken() (string, error) {
	c := 30
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("could not create random bytes, %s", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// checkToken is a middleware that checks for the right Bearer token, if not found, you get
// a 403 error, else, the handler will call the next function
func checkToken(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("Authorization")
		if strings.TrimSpace(t) == "Bearer "+token {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid or missing Bearer", 403)
		}
	})
}
