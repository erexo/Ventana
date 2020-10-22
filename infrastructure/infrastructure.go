package infrastructure

import (
	"log"

	"github.com/Erexo/Ventana/api"
	"github.com/Erexo/Ventana/infrastructure/gpio"
	"github.com/Erexo/Ventana/infrastructure/sunblind"
	"github.com/Erexo/Ventana/infrastructure/thermal"
	"github.com/Erexo/Ventana/infrastructure/user"
)

var (
	pm *gpio.PinManager

	us *user.Service
	ss *sunblind.Service
	ts *thermal.Service
)

func Run() {
	pm = gpio.CreatePinManager()

	us = user.CreateService()
	ss = sunblind.CreateService(pm)
	if err := ss.Load(); err != nil {
		log.Println("SunblindService error:", err)
	}
	ts = thermal.CreateService()
	if err := ts.Load(); err != nil {
		log.Println("ThermalService error:", err)
	}

	// todo, add flag to run api
	if err := api.Run(us, ss, ts); err != nil {
		log.Println("Api error:", err)
	}
}

func Terminate() error {
	if pm == nil {
		return nil
	}
	return pm.Close()
}
