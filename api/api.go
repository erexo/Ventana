package api

import (
	"net/http"

	"github.com/Erexo/Ventana/infrastructure/sunblind"
	"github.com/Erexo/Ventana/infrastructure/thermal"
	"github.com/Erexo/Ventana/infrastructure/user"
	"github.com/go-chi/chi"
)

type Controller interface {
	GetPrefix() string
	Route(r chi.Router)
}

func Run(as *user.Service, ss *sunblind.Service, ts *thermal.Service) error {
	r := chi.NewRouter()

	registerController(r, user.CreateController(as))
	registerController(r, sunblind.CreateController(ss))
	registerController(r, thermal.CreateController(ts))

	// todo port from env/config
	return http.ListenAndServe(":8081", r)
}

func registerController(r chi.Router, c Controller) {
	r.Route(c.GetPrefix(), c.Route)
}
