package sunblind

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Erexo/Ventana/api/controller"
	"github.com/Erexo/Ventana/core/domain"
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
	r.Post("/order", c.order)
	r.Post("/browse", c.browse)
	r.Post("/create", c.create)
	r.Patch("/update/{id}", c.update)
	r.Delete("/delete/{id}", c.delete)
	r.Post("/toggle/{id}/{dir}", c.toggle)
}

// @Router /api/sunblind/order [post]
// @Param body body []int64 true "body"
// @Success 200 {string} plain
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
func (c *Controller) order(w http.ResponseWriter, r *http.Request) {
	claims, ok := controller.ReadClaims(w, r)
	if !ok {
		return
	}

	var d []int64
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.s.SaveOrder(claims.UserId, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /api/sunblind/browse [post]
// @Success 200 {array} dto.Sunblind
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
func (c *Controller) browse(w http.ResponseWriter, r *http.Request) {
	claims, ok := controller.ReadClaims(w, r)
	if !ok {
		return
	}

	w.Header().Set("content-type", "application/json")
	ret, err := c.s.Browse(claims.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	retj, _ := json.Marshal(ret)
	w.WriteHeader(http.StatusOK)
	w.Write(retj)
}

// @Router /api/sunblind/create [post]
// @Param body body saveDto true "body"
// @Success 200 {string} plain
// @Accept  json
// @Produce  plain
// @Security ApiKeyAuth
func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	var d saveDto

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Create(d.Name, d.InputDownPin, d.InputUpPin, d.OutputDownPin, d.OutputUpPin); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /api/sunblind/update/{id} [patch]
// @Param id path int true "path"
// @Param body body saveDto true "body"
// @Success 200 {string} plain
// @Accept  json
// @Produce  plain
// @Security ApiKeyAuth
func (c *Controller) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var d saveDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Update(id, d.Name, d.InputDownPin, d.InputUpPin, d.OutputDownPin, d.OutputUpPin); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /api/sunblind/delete/{id} [delete]
// @Param id path int true "path"
// @Success 200 {string} plain
// @Security ApiKeyAuth
func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// @Router /api/sunblind/toggle/{id}/{dir} [post]
// @Param id path int true "path"
// @Param dir path string true "path"
// @Success 200 {string} plain
// @Security ApiKeyAuth
func (c *Controller) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dir := chi.URLParam(r, "dir")
	i, err := strconv.Atoi(dir)
	if err != nil {
		log.Printf("Unable to read toggle direction '%s' for sunblind '%d'", dir, id)
	} else if err := c.s.Toggle(id, i == 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type saveDto struct {
	Name          string     `json:"name"`
	InputDownPin  domain.Pin `json:"inputdownpin"`
	InputUpPin    domain.Pin `json:"inputuppin"`
	OutputDownPin domain.Pin `json:"outputdownpin"`
	OutputUpPin   domain.Pin `json:"outputuppin"`
}
