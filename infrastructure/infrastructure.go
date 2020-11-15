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
	gs *gpio.Service
	us *user.Service
	ss *sunblind.Service
	ts *thermal.Service
)

func Run() {
	gs = gpio.CreateService()
	us = user.CreateService()
	ss = sunblind.CreateService(gs)
	if err := ss.Load(); err != nil {
		log.Println("SunblindService error:", err)
	} else {
		log.Println("Loaded sunblind service")
	}
	ts = thermal.CreateService()
	if err := ts.Load(); err != nil {
		log.Println("ThermalService error:", err)
	} else {
		log.Println("Loaded thermal service")
	}

	// todo, add flag to run api
	if err := api.Run(us, ss, ts); err != nil {
		log.Println("Api error:", err)
	}
}

func Terminate() error {
	if gs == nil {
		return nil
	}
	return gs.Close()
}
