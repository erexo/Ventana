package thermal

import (
	"log"
	"time"

	"github.com/yryz/ds18b20"
)

const (
	checkInterval = 1000 * time.Millisecond
)

type Service struct {
	sensors map[string]float64
}

func CreateService() *Service {
	return &Service{
		sensors: make(map[string]float64),
	}
}

func (s *Service) Load() error {
	go func() {
		for {
			time.Sleep(checkInterval)
			s.updateSensors()
		}
	}()
	return nil
}

func (s *Service) updateSensors() {
	sensors := make(map[string]float64)
	names, err := ds18b20.Sensors()
	if err != nil {
		log.Println("Unable to discover sensors:", err)
	} else {
		for _, name := range names {
			t, err := ds18b20.Temperature(name)
			if err != nil {
				log.Printf("Unable to load sensor '%s': %v", name, err)
				continue
			}
			log.Printf("%s: %.2f°C", name, t)
			sensors[name] = t
		}
	}
	s.sensors = sensors
}
