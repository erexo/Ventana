package entity

import (
	"fmt"
	"time"
)

type Temperature float32

func (t Temperature) String() string {
	return fmt.Sprintf("%.2f°C", t)
}

type UnixTime int64

func (t UnixTime) Time() time.Time {
	return time.Unix(int64(t), 0)
}

func (t UnixTime) String() string {
	return t.Time().String()
}

type ThermalData struct {
	Id            int64       `db:"id"`
	ThermometerId int64       `db:"thermometerid"`
	Celsius       Temperature `db:"celsius"`
	Timestamp     UnixTime    `db:"timestamp"`
}
