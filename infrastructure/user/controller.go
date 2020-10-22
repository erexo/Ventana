package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	return "/user"
}

func (c *Controller) Route(r chi.Router) {
	r.Post("/login", c.login)
	r.Post("/create", c.create)
	r.Patch("/update/password/{id}", c.updatePassword)
	r.Patch("/update/role/{id}", c.updateRole)
	r.Delete("/delete/{id}", c.delete)
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d struct {
		Username string `json:"username"`
		Password string `json:"Password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Login(d.Username, d.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d struct {
		Username string      `json:"username"`
		Password string      `json:"password"`
		Role     entity.Role `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Create(d.Username, d.Password, d.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) updatePassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("begin")
	w.Header().Set("content-type", "application/json")
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var d struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.UpdatePassword(id, d.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) updateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var d struct {
		Role entity.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.UpdateRole(id, d.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
