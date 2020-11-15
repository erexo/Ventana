package dto

import "github.com/Erexo/Ventana/core/entity"

type Point struct {
	Celsius   entity.Temperature `json:"celsius" db:"celsius"`
	Timestamp entity.UnixTime    `json:"timestamp" db:"timestamp"`
}

func CreatePoint(temp entity.Temperature, time entity.UnixTime) Point {
	return Point{
		Celsius:   temp,
		Timestamp: time,
	}
}
