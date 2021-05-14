package provider

import "net/http"

type Provider interface {
	Key() string
	Authenticate(w http.ResponseWriter, req *http.Request)
	Callback(w http.ResponseWriter, req *http.Request)
}
