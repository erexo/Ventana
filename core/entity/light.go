package entity

import "github.com/Erexo/Ventana/core/domain"

type Light struct {
	Id        int64      `db:"id"`
	Name      string     `db:"name"`
	InputPin  domain.Pin `db:"inputpin"`
	OutputPin domain.Pin `db:"outputpin"`
}
