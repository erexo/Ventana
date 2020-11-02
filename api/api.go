package api

import (
	"log"
	"net/http"

	"github.com/Erexo/Ventana/api/controller"
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

func Run(us *user.Service, ss *sunblind.Service, ts *thermal.Service) error {
	config := config.GetConfig()
	if !config.ApiAddr.Valid {
		return nil
	}

	r := chi.NewRouter()
	addCors(r)

	token := jwtauth.New("HS256", []byte(config.JwtToken), nil)
	registerController(r, us, token, user.CreateController(us))
	registerController(r, us, token, sunblind.CreateController(ss))
	registerController(r, us, token, thermal.CreateController(ts))

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

func registerController(r chi.Router, us *user.Service, token *jwtauth.JWTAuth, c Controller) {
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(token))
		r.Use(authenticate(us))
		r.Route(c.GetPrefix(), c.Route)
	})

	if uc, ok := c.(UnauthorizedController); ok {
		r.Group(func(r chi.Router) {
			r.Route(uc.GetUnauthorizedPrefix(), uc.UnauthorizedRoute)
		})
	}
}

func authenticate(us *user.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := controller.ReadClaims(w, r)
			if !ok {
				// unauthorization is performed inside ReadClaims
				return
			}
			if !us.Verify(claims.UserId, claims.Hash, claims.Role) {
				controller.Unauthorize(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// @Success 200 {string} plain
// @Router / [get]
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ventana"))
}
