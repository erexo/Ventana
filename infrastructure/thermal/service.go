package thermal

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/Erexo/Ventana/infrastructure/db"
	"github.com/yryz/ds18b20"
)

const blockSize = 100

type Service struct {
	thermometers map[int64]*ThermalBlock
}

func CreateService() *Service {
	return &Service{
		thermometers: make(map[int64]*ThermalBlock),
	}
}

func (s *Service) Load() error {
	go func() {
		updateInterval := config.GetConfig().ThermalUpdateInterval
		for {
			time.Sleep(time.Duration(updateInterval) * time.Millisecond)
			s.updateSensors()
		}
	}()
	return nil
}

func (s *Service) updateSensors() {
	now := time.Now()
	defer log.Println("Updated sensors in", time.Now().Sub(now))

	names, err := ds18b20.Sensors()
	if err != nil {
		log.Println("Unable to discover sensors:", err)
		return
	}

	for _, name := range names {
		var therm entity.Thermometer
		err := db.Get(&therm, "SELECT id, name, sensor FROM thermometer WHERE sensor LIKE ?", name)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Println("Thermometer parsing:", err)
			}
			continue
		}
		temp, err := ds18b20.Temperature(name)
		if err != nil {
			log.Printf("Unable to load sensor '%s': %v\n", name, err)
			continue
		}

		now := time.Now().UTC().Unix()
		if _, err := db.Exec("INSERT INTO thermaldata (thermometerid, celsius, timestamp) VALUES (?, ?, ?)", therm.Id, temp, now); err != nil {
			log.Println("Thermometer temperature saving:", err)
		}

		block, ok := s.thermometers[therm.Id]
		if !ok {
			block = CreateThermalBlock(blockSize)
			s.thermometers[therm.Id] = block
		}
		block.Add(entity.Temperature(temp), entity.UnixTime(now))
	}
}
