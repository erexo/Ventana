package sunblind

import (
	"net/http"
	"strconv"

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
	r.Get("/toggle/{id}/{dir}", c.toggle)
	r.Post("/create", c.create)
}

func (c *Controller) toggle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	dir := chi.URLParam(r, "dir")
	c.s.ToggleSunblind(id, dir == "down")
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {

}
