package entity

import "github.com/Erexo/Ventana/core/domain"

type Sunblind struct {
	Id            int // guid?
	Name          string
	InputDownPin  domain.Pin
	InputUpPin    domain.Pin
	OutputDownPin domain.Pin
	OutputUpPin   domain.Pin
}