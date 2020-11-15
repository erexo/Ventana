package dto

import "github.com/Erexo/Ventana/core/domain"

type Sunblind struct {
	Id            int64      `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	InputDownPin  domain.Pin `json:"inputdownpin" db:"inputdownpin"`
	InputUpPin    domain.Pin `json:"inputuppin" db:"inputuppin"`
	OutputDownPin domain.Pin `json:"outputdownpin" db:"outputdownpin"`
	OutputUpPin   domain.Pin `json:"outputuppin" db:"outputuppin"`
}
