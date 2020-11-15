package entity

import "github.com/Erexo/Ventana/core/domain"

type Sunblind struct {
	Id            int64      `db:"id"`
	Name          string     `db:"name"`
	InputDownPin  domain.Pin `db:"inputdownpin"`
	InputUpPin    domain.Pin `db:"inputuppin"`
	OutputDownPin domain.Pin `db:"outputdownpin"`
	OutputUpPin   domain.Pin `db:"outputuppin"`
}
