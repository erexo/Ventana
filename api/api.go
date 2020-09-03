package api

import (
	"net/http"

	"github.com/Erexo/Ventana/infrastructure/sunblind"
	"github.com/go-chi/chi"
)

type Controller interface {
	GetPrefix() string
	Route(r chi.Router)
}

func Run(ss *sunblind.Service) error {
	r := chi.NewRouter()

	registerController(r, sunblind.CreateController(ss))

	// todo port from env/config
	return http.ListenAndServe(":8081", r)
}

func registerController(r chi.Router, c Controller) {
	r.Route(c.GetPrefix(), c.Route)
}
