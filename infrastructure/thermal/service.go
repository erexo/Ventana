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
	"github.com/georgysavva/scany/sqlscan"
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
	err := db.Select(&ret, "SELECT celsius, timestamp FROM thermaldata WHERE thermometerid=? AND timestamp>=? AND timestamp<=?", id, from, to)
	return ret, err
}

func (s *Service) SaveOrder(userId int64, order []int64) error {
	tx, close, err := db.GetTransaction()
	if err != nil {
		return err
	}
	defer close()

	var currentOrder []struct {
		OrderId       int64 `db:"id"`
		ThermometerId int64 `db:"thermometerid"`
	}

	if err := sqlscan.Select(db.Ctx(), tx, &currentOrder, "SELECT id, thermometerid FROM thermometerorder WHERE userid=? ORDER BY id ASC", userId); db.IsError(err) {
		return err
	}

	lastId := int64(0)
	for i, thermometerId := range order {
		if i < len(currentOrder) {
			c := currentOrder[i]
			if thermometerId != c.ThermometerId {
				if _, err := tx.Exec("UPDATE thermometerorder SET thermometerid=? WHERE id=?", thermometerId, c.OrderId); err != nil {
					return err
				}
			}
			lastId = c.OrderId
			continue
		}
		if _, err := tx.Exec("INSERT INTO thermometerorder (userid, thermometerid) VALUES (?, ?)", userId, thermometerId); err != nil {
			return err
		}
	}
	if len(order) < len(currentOrder) {
		tx.Exec("DELETE FROM thermometerorder WHERE userid=? AND id>?", userId, lastId)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Updated Thermometer order for User '%d': %v\n", userId, order)
	return nil
}

func (s *Service) Browse(userId int64) ([]*dto.Thermometer, error) {
	var thermos []*dto.Thermometer
	query := fmt.Sprintf("SELECT id, name, sensor FROM thermometer ORDER BY id ASC")
	err := db.Select(&thermos, query)
	if err != nil {
		return nil, err
	}
	var order []int64
	if err := db.Select(&order, "SELECT thermometerid FROM thermometerorder WHERE userid=? ORDER BY id ASC", userId); err != nil {
		return nil, err
	}
	var ret []*dto.Thermometer
	if len(order) == 0 {
		ret = thermos
	} else {
		ret = make([]*dto.Thermometer, len(thermos))
		i := 0
		for _, id := range order {
			for j, thermo := range thermos {
				if thermo != nil && thermo.Id == id {
					ret[i] = thermo
					i++
					thermos[j] = nil
					break
				}
			}
		}
		for _, thermo := range thermos {
			if thermo != nil {
				ret[i] = thermo
				i++
			}
		}
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

	log.Printf("Created thermometer '%d' with Name %s\n", id, name)
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
	log.Printf("Deleted thermometer '%d'\n", id)
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
