package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Erexo/Ventana/core/entity"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
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
	r.Post("/create", c.create)
	r.Patch("/update/password/{id}", c.updatePassword)
	r.Patch("/update/role/{id}", c.updateRole)
	r.Delete("/delete/{id}", c.delete)

	r.Post("/test", c.test)
}

func (c *Controller) GetUnauthorizedPrefix() string {
	return "/login"
}

func (c *Controller) UnauthorizedRoute(r chi.Router) {
	r.Post("/", c.login)
	r.Post("/create", c.create)
}

func (c *Controller) test(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Write([]byte(fmt.Sprintf("Role: %v", claims["role"])))
}

// @Router /login [post]
// @Param body body loginDto true "body"
// @Success 200 {object} user.LoginInfo
// @Accept  json
// @Produce  json
func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d loginDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret, err := c.s.Login(d.Username, d.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	retj, _ := json.Marshal(ret)
	w.WriteHeader(http.StatusOK)
	w.Write(retj)
}

// @Router /user/create [post]
// @Param body body createDto true "body"
// @Success 200 {string} plain
// @Accept  json
// @Produce  plain
// @Security ApiKeyAuth
func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var d createDto

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.Create(d.Username, d.Password, d.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /user/update/password/{id} [patch]
// @Param id path int true "path"
// @Param body body updatePasswordDto true "body"
// @Success 200 {string} plain
// @Security ApiKeyAuth
func (c *Controller) updatePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var d updatePasswordDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.UpdatePassword(id, d.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /user/update/role/{id} [patch]
// @Param id path int true "path"
// @Param body body updateRoleDto true "body"
// @Success 200 {string} plain
// @Security ApiKeyAuth
func (c *Controller) updateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var d updateRoleDto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.s.UpdateRole(id, d.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Router /user/delete/{id} [delete]
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

type loginDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createDto struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Role     entity.Role `json:"role"`
}

type updatePasswordDto struct {
	Password string `json:"password"`
}

type updateRoleDto struct {
	Role entity.Role `json:"role"`
}
