package dto

import (
	"github.com/Erexo/Ventana/core/domain"
)

type Light struct {
	Id        int64      `json:"id"`
	Name      string     `json:"name"`
	InputPin  domain.Pin `json:"inputpin"`
	OutputPin domain.Pin `json:"outputpin"`
	State  bool       `json:"position"`
}
