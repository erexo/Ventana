package dto

import "github.com/Erexo/Ventana/core/entity"

type Thermometer struct {
	Id      int64               `json:"id" db:"id"`
	Name    string              `json:"name" db:"name"`
	Sensor  string              `json:"sensor" db:"sensor"`
	Celsius *entity.Temperature `json:"celsius" db:"-"`
}
