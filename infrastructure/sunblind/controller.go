package sunblind

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Controller struct {
	s *Service
}

func CreateController(s *Service) *Controller {
	return &Controller{
		s: s,
	}
}

func (c *Controller) GetPrefix() string {
	return "/sunblind"
}

func (c *Controller) Route(r chi.Router) {
	r.Get("/get", c.get)
	r.Post("/create", c.create)
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
	c.s.ToggleSunblind(0)
	w.Write([]byte("hello"))
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {

}
