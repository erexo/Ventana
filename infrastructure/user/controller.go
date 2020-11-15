package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/core/dto"
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
	r.Post("/browse", c.browse)
	r.Post("/create", c.create)
	r.Patch("/update/password/{id}", c.updatePassword)
	r.Patch("/update/role/{id}", c.updateRole)
	r.Delete("/delete/{id}", c.delete)
}

func (c *Controller) GetUnauthorizedPrefix() string {
	return "/login"
}

func (c *Controller) UnauthorizedRoute(r chi.Router) {
	r.Post("/", c.login)
	r.Post("/create", c.create)
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
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		http.Error(w, err.Error(), status)
		return
	}
	retj, _ := json.Marshal(ret)
	w.WriteHeader(http.StatusOK)
	w.Write(retj)
}

// @Router /user/browse [post]
// @Param body body dto.Filters true "body"
// @Success 200 {array} dto.User
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

// @Router /user/create [post]
// @Param body body createDto true "body"
// @Success 200 {string} plain
// @Accept  json
// @Produce  plain
// @Security ApiKeyAuth
func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
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
	Role     domain.Role `json:"role"`
}

type updatePasswordDto struct {
	Password string `json:"password"`
}

type updateRoleDto struct {
	Role domain.Role `json:"role"`
}
