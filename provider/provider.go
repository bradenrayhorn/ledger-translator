package provider

import "net/http"

type Provider interface {
	Key() string
	Authenticate(w http.ResponseWriter, req *http.Request, jwtString string, userID string)
	Callback(w http.ResponseWriter, req *http.Request)
}
