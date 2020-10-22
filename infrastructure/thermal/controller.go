package thermal

import "github.com/go-chi/chi"

type Controller struct {
	s *Service
}

func CreateController(s *Service) *Controller {
	return &Controller{
		s: s,
	}
}

func (c *Controller) GetPrefix() string {
	return "/thermal"
}

func (c *Controller) Route(r chi.Router) {
	// todo
}
