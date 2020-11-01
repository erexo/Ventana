package thermal

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Erexo/Ventana/core/dto"
	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/Erexo/Ventana/infrastructure/db"
	"github.com/yryz/ds18b20"
)

const blockSize = 100

type Service struct {
	thermometers    map[int64]*ThermalBlock
	thermometersMux sync.Mutex
}

func CreateService() *Service {
	return &Service{
		thermometers: make(map[int64]*ThermalBlock),
	}
}

func (s *Service) GetData(id int64, from, to entity.UnixTime) ([]dto.Point, error) {
	var ret []dto.Point
	fmt.Println(id, from, to)
	err := db.Select(&ret, "SELECT celsius, timestamp FROM thermaldata WHERE thermometerid=? AND timestamp>=? AND timestamp<=?", id, from, to)
	return ret, err
}

func (s *Service) Browse(filters dto.Filters) ([]*dto.Thermometer, error) {
	var ret []*dto.Thermometer
	if err := filters.Validate(dto.Thermometer{}); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT id, name, sensor FROM thermometer%s", filters.GetQuery())
	err := db.Select(&ret, query)
	if err != nil {
		return nil, err
	}

	s.thermometersMux.Lock()
	defer s.thermometersMux.Unlock()
	for _, t := range ret {
		if temp, ok := s.thermometers[t.Id]; ok {
			if p, err := temp.Last(); err == nil {
				r := entity.Temperature(math.Round(float64(p.Celsius)))
				t.Celsius = &r
			}
		}
	}
	return ret, nil
}

func (s *Service) Create(name, sensor string) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}
	if err := entity.ValidateEmpty(&sensor); err != nil {
		return fmt.Errorf("Sensor: %w", err)
	}
	r, err := db.Exec("INSERT INTO thermometer (name, sensor) VALUES (?, ?)", name, sensor)
	if err != nil {
		return err
	}
	id, _ := r.LastInsertId()

	log.Printf("Created thermometer '%d' with Name %s", id, name)
	return nil
}

func (s *Service) Update(id int64, name, sensor string) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}
	if err := entity.ValidateEmpty(&sensor); err != nil {
		return fmt.Errorf("Sensor: %w", err)
	}
	if _, err := db.Exec("UPDATE thermometer SET name=?, sensor=? WHERE id=?", name, sensor, id); err != nil {
		return err
	}
	log.Printf("Updated thermometer '%d'\n", id)
	return nil
}

func (s *Service) Delete(id int64) error {
	r, err := db.Exec("DELETE FROM thermometer WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, _ := r.RowsAffected()
	if rows < 1 {
		return fmt.Errorf("Thermometer '%d' does not exist", id)
	}
	log.Printf("Deleted thermometer '%d'", id)
	return nil
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
	s.thermometersMux.Lock()
	defer s.thermometersMux.Unlock()

	cfg := config.GetConfig()
	if cfg.GenerateRandomTemperature {
		var therms []entity.Thermometer
		if err := db.Select(&therms, "SELECT id, name, sensor FROM thermometer"); err != nil {
			log.Println("Retrieving thermometrs:", err)
			return
		}

		for _, therm := range therms {
			temp := float64(-10+rand.Intn(30)) + rand.Float64()
			s.addTemperature(therm.Id, temp)
		}
	} else {
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
					log.Printf("Retrieving thermometer '%s': %s\n", name, err)
				}
				continue
			}
			temp, err := ds18b20.Temperature(name)
			if err != nil {
				log.Printf("Unable to load sensor '%s': %v\n", name, err)
				continue
			}

			s.addTemperature(therm.Id, temp)
		}
	}
}

func (s *Service) addTemperature(id int64, temp float64) {
	now := time.Now().UTC().Unix()
	if _, err := db.Exec("INSERT INTO thermaldata (thermometerid, celsius, timestamp) VALUES (?, ?, ?)", id, temp, now); err != nil {
		log.Println("Thermometer temperature saving:", err)
	}
	block, ok := s.thermometers[id]
	if !ok {
		block = CreateThermalBlock(blockSize)
		s.thermometers[id] = block
	}
	block.Add(entity.Temperature(temp), entity.UnixTime(now))
}
