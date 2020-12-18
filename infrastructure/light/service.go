package light

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/core/dto"
	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/core/enum"
	"github.com/Erexo/Ventana/infrastructure/db"
	"github.com/Erexo/Ventana/infrastructure/gpio"
	"github.com/georgysavva/scany/sqlscan"
)

type Service struct {
	gs *gpio.Service
}

func CreateService(pm *gpio.Service) *Service {
	return &Service{
		gs: pm,
	}
}

func (s *Service) SaveOrder(userId int64, order []int64) error {
	tx, close, err := db.GetTransaction()
	if err != nil {
		return err
	}
	defer close()

	var currentOrder []struct {
		OrderId int64 `db:"id"`
		LightId int64 `db:"lightid"`
	}

	if err := sqlscan.Select(db.Ctx(), tx, &currentOrder, "SELECT id, lightid FROM lightorder WHERE userid=? ORDER BY id ASC", userId); db.IsError(err) {
		return err
	}

	lastId := int64(0)
	for i, lightid := range order {
		if i < len(currentOrder) {
			c := currentOrder[i]
			if lightid != c.LightId {
				if _, err := tx.Exec("UPDATE lightorder SET lightid=? WHERE id=?", lightid, c.OrderId); err != nil {
					return err
				}
			}
			lastId = c.OrderId
			continue
		}
		if _, err := tx.Exec("INSERT INTO lightorder (userid, lightid) VALUES (?, ?)", userId, lightid); err != nil {
			return err
		}
	}
	if len(order) < len(currentOrder) {
		tx.Exec("DELETE FROM lightorder WHERE userid=? AND id>?", userId, lastId)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Updated Light order for User '%d': %v\n", userId, order)
	return nil
}

func (s *Service) Browse(userId int64) ([]*dto.Light, error) {
	var lights []*entity.Light
	err := db.Select(&lights, "SELECT id, name, inputpin, outputpin FROM light ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	var order []int64
	if err := db.Select(&order, "SELECT lightid FROM lightorder WHERE userid=? ORDER BY id ASC", userId); err != nil {
		return nil, err
	}
	ret := make([]*dto.Light, len(lights))
	i := 0
	for _, id := range order {
		for j, light := range lights {
			if light != nil && light.Id == id {
				ret[i] = s.getLight(light)
				i++
				lights[j] = nil
				break
			}
		}
	}
	for _, light := range lights {
		if light != nil {
			ret[i] = s.getLight(light)
			i++
		}
	}
	return ret, nil
}

func (s *Service) Create(name string, inputPin, outputPin domain.Pin) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}
	if err := s.gs.IsPinRegistered(inputPin, outputPin); err != nil {
		return err
	}

	r, err := db.Exec("INSERT INTO light (name, inputpin, outputpin) VALUES (?, ?, ?)", name, inputPin, outputPin)
	if err != nil {
		return err
	}
	id, _ := r.LastInsertId()

	if err := s.gs.RegisterPinPair(inputPin, outputPin, enum.PairTypeToggle); err != nil {
		return err
	}

	log.Printf("Created Light '%d' with Name %s\n", id, name)
	return nil
}

func (s *Service) Update(id int64, name string, inputPin, outputPin domain.Pin) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}

	light, err := getData(id)
	if err != nil {
		return err
	}
	var newPins []domain.Pin
	for _, p := range []domain.Pin{inputPin, outputPin} {
		if !light.ContainsPin(p) {
			newPins = append(newPins, p)
		}
	}
	if len(newPins) > 0 {
		if err := s.gs.IsPinRegistered(newPins...); err != nil {
			return err
		}
	}

	if _, err := db.Exec("UPDATE light SET name=?, inputpin=?, outputpin=? WHERE id=?", name, inputPin, outputPin, id); err != nil {
		return err
	}

	if len(newPins) > 0 {
		if err := s.gs.UnregisterPinPair(light.InputPin, light.OutputPin); err != nil {
			return err
		}
		if err := s.gs.RegisterPinPair(inputPin, outputPin, enum.PairTypeToggle); err != nil {
			return err
		}
	}

	log.Printf("Updated Light '%d'\n", id)
	return nil
}

func (s *Service) Delete(id int64) error {
	light, err := getData(id)
	if err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM light WHERE id=?", id); err != nil {
		return err
	}

	if err := s.gs.UnregisterPinPair(light.InputPin, light.OutputPin); err != nil {
		return err
	}

	log.Printf("Deleted Light '%d'\n", id)
	return nil
}

func (s *Service) Toggle(id int64) (error, bool) {
	var pin domain.Pin
	if err := db.Get(&pin, "SELECT inputpin FROM light WHERE id=?", id); err != nil {
		return err, false
	}
	err := s.gs.TogglePin(pin)
	state := s.gs.GetPinState(pin)
	return err, state
}

func (s *Service) Load() error {
	var lights []*loadData
	err := db.Select(&lights, "SELECT inputpin, outputpin FROM light")
	if err != nil {
		return err
	}
	for _, light := range lights {
		if err := s.gs.RegisterPinPair(light.InputPin, light.OutputPin, enum.PairTypeToggle); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getLight(light *entity.Light) *dto.Light {
	ret := dto.Light{
		Id:        light.Id,
		Name:      light.Name,
		InputPin:  light.InputPin,
		OutputPin: light.OutputPin,
	}
	ret.State = s.gs.GetPinState(ret.InputPin)
	return &ret
}

func getData(id int64) (loadData, error) {
	var light loadData
	if err := db.Get(&light, "SELECT inputpin, outputpin FROM light WHERE id=?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return loadData{}, fmt.Errorf("Light '%d' does not exist", id)
		}
		return loadData{}, err
	}
	return light, nil
}

type loadData struct {
	InputPin  domain.Pin `db:"inputdownpin"`
	OutputPin domain.Pin `db:"outputdownpin"`
}

func (l loadData) ContainsPin(pin domain.Pin) bool {
	switch pin {
	case l.InputPin,
		l.OutputPin:
		return true
	default:
		return false
	}
}
