package thermal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Erexo/Ventana/api/controller"
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
	r.Post("/order", c.order)
	r.Post("/browse", c.browse)
	r.Post("/data", c.data)
	r.Post("/create", c.create)
	r.Patch("/update/{id}", c.update)
	r.Delete("/delete/{id}", c.delete)
}

// @Router /api/thermal/order [post]
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

// @Router /api/thermal/browse [post]
// @Success 200 {array} dto.Thermometer
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

// @Router /api/thermal/data [post]
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

// @Router /api/thermal/create [post]
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

// @Router /api/thermal/update/{id} [patch]
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

// @Router /api/thermal/delete/{id} [delete]
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
