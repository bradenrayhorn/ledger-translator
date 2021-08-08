package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func CreateRouter(controller RouteController) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Logger)

	r.Get("/authenticate", controller.Authenticate)
	r.Get("/callback", controller.Callback)
	r.Get("/api/v1/providers", controller.GetProviders)
	r.Get("/health-check", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
	})

	return r
}
