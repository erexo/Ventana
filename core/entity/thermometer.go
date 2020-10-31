package entity

type Thermometer struct {
	Id     int64  `db:"id"`
	Name   string `db:"name"`
	Sensor string `db:"sensor"`
}
