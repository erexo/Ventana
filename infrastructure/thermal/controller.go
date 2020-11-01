package thermal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Erexo/Ventana/core/dto"
	"github.com/Erexo/Ventana/core/entity"
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
	return "/thermal"
}

func (c *Controller) Route(r chi.Router) {
	r.Post("/browse", c.browse)
	r.Post("/data", c.data)
	r.Post("/create", c.create)
	r.Patch("/update/{id}", c.update)
	r.Delete("/delete/{id}", c.delete)
}

// @Router /thermal/browse [post]
// @Param body body dto.Filters true "body"
// @Success 200 {array} dto.Thermometer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
func (c *Controller) browse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d dto.Filters
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret, err := c.s.Browse(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	retj, _ := json.Marshal(ret)
	w.WriteHeader(http.StatusOK)
	w.Write(retj)
}

// @Router /thermal/data [post]
// @Param body body dataDto true "body"
// @Success 200 {array} dto.Point
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
func (c *Controller) data(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d dataDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret, err := c.s.GetData(d.ThermometerId, d.From, d.To)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	retj, _ := json.Marshal(ret)
	w.WriteHeader(http.StatusOK)
	w.Write(retj)
}

// @Router /thermal/create [post]
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
	if err := c.s.Create(d.Name, d.Sensor); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /thermal/update/{id} [patch]
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
	if err := c.s.Update(id, d.Name, d.Sensor); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /thermal/delete/{id} [delete]
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

type dataDto struct {
	ThermometerId int64           `json:"thermometerid"`
	From          entity.UnixTime `json:"from"`
	To            entity.UnixTime `json:"to"`
}

type saveDto struct {
	Name   string `json:"name"`
	Sensor string `json:"sensor"`
}
