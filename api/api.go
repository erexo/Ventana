package api

import (
	"log"
	"net/http"

	_ "github.com/Erexo/Ventana/docs"
	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/Erexo/Ventana/infrastructure/sunblind"
	"github.com/Erexo/Ventana/infrastructure/thermal"
	"github.com/Erexo/Ventana/infrastructure/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Controller interface {
	GetPrefix() string
	Route(r chi.Router)
}

type UnauthorizedController interface {
	GetUnauthorizedPrefix() string
	UnauthorizedRoute(r chi.Router)
}

func Run(as *user.Service, ss *sunblind.Service, ts *thermal.Service) error {
	config := config.GetConfig()
	if !config.ApiAddr.Valid {
		return nil
	}

	r := chi.NewRouter()
	addCors(r)

	token := jwtauth.New("HS256", []byte(config.JwtToken), nil)
	registerController(r, token, user.CreateController(as))
	registerController(r, token, sunblind.CreateController(ss))
	registerController(r, token, thermal.CreateController(ts))

	r.Get("/", home)

	if config.UseSwagger {
		r.Mount("/swagger", httpSwagger.WrapHandler)
	}
	log.Printf("Initializing API '%s'\n", config.ApiAddr.String)
	return http.ListenAndServe(config.ApiAddr.String, r)
}

func addCors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Referer"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func registerController(r chi.Router, token *jwtauth.JWTAuth, c Controller) {
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token))
		r.Use(jwtauth.Authenticator)
		r.Route(c.GetPrefix(), c.Route)
	})

	if uc, ok := c.(UnauthorizedController); ok {
		r.Group(func(r chi.Router) {
			r.Route(uc.GetUnauthorizedPrefix(), uc.UnauthorizedRoute)
		})
	}
}

// @Success 200 {string} plain
// @Router / [get]
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ventana"))
}
