package sunblind

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
		OrderId    int64 `db:"id"`
		SunblindId int64 `db:"sunblindid"`
	}

	if err := sqlscan.Select(db.Ctx(), tx, &currentOrder, "SELECT id, sunblindid FROM sunblindorder WHERE userid=? ORDER BY id ASC", userId); db.IsError(err) {
		return err
	}

	lastId := int64(0)
	for i, sunblindId := range order {
		if i < len(currentOrder) {
			c := currentOrder[i]
			if sunblindId != c.SunblindId {
				if _, err := tx.Exec("UPDATE sunblindorder SET sunblindid=? WHERE id=?", sunblindId, c.OrderId); err != nil {
					return err
				}
			}
			lastId = c.OrderId
			continue
		}
		if _, err := tx.Exec("INSERT INTO sunblindorder (userid, sunblindid) VALUES (?, ?)", userId, sunblindId); err != nil {
			return err
		}
	}
	if len(order) < len(currentOrder) {
		tx.Exec("DELETE FROM sunblindorder WHERE userid=? AND id>?", userId, lastId)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Updated Sunblind order for User '%d': %v\n", userId, order)
	return nil
}

func (s *Service) Browse(userId int64) ([]*dto.Sunblind, error) {
	var sunblinds []*dto.Sunblind
	err := db.Select(&sunblinds, "SELECT id, name, inputdownpin, inputuppin, outputdownpin, outputuppin FROM sunblind ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	var order []int64
	if err := db.Select(&order, "SELECT sunblindid FROM sunblindorder WHERE userid=? ORDER BY id ASC", userId); err != nil {
		return nil, err
	}
	var ret []*dto.Sunblind
	if len(order) == 0 {
		ret = sunblinds
	} else {
		ret = make([]*dto.Sunblind, len(sunblinds))
		i := 0
		for _, id := range order {
			for j, sunblind := range sunblinds {
				if sunblind != nil && sunblind.Id == id {
					ret[i] = sunblind
					i++
					sunblinds[j] = nil
					break
				}
			}
		}
		for _, sunblind := range sunblinds {
			if sunblind != nil {
				ret[i] = sunblind
				i++
			}
		}
	}
	return ret, nil
}

func (s *Service) Create(name string, inputDownPin, inputUpPin, outputDownPin, outputUpPin domain.Pin) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}
	if err := s.gs.IsPinRegistered(inputDownPin, inputUpPin, outputDownPin, outputUpPin); err != nil {
		return err
	}

	r, err := db.Exec("INSERT INTO sunblind (name, inputdownpin, inputuppin, outputdownpin, outputuppin) VALUES (?, ?, ?, ?, ?)", name, inputDownPin, inputUpPin, outputDownPin, outputUpPin)
	if err != nil {
		return err
	}
	id, _ := r.LastInsertId()

	if err := s.gs.RegisterPinPair(inputDownPin, outputDownPin, enum.PairTypeTimed); err != nil {
		return err
	}
	if err := s.gs.RegisterPinPair(inputUpPin, outputUpPin, enum.PairTypeTimed); err != nil {
		return err
	}

	log.Printf("Created sunblind '%d' with Name %s\n", id, name)
	return nil
}

func (s *Service) Update(id int64, name string, inputDownPin, inputUpPin, outputDownPin, outputUpPin domain.Pin) error {
	if err := entity.ValidateName(&name); err != nil {
		return fmt.Errorf("Name: %w", err)
	}

	sunblind, err := getData(id)
	if err != nil {
		return err
	}
	var newPins []domain.Pin
	for _, p := range []domain.Pin{inputDownPin, inputUpPin, outputDownPin, outputUpPin} {
		if !sunblind.ContainsPin(p) {
			newPins = append(newPins, p)
		}
	}
	if len(newPins) > 0 {
		if err := s.gs.IsPinRegistered(newPins...); err != nil {
			return err
		}
	}

	if _, err := db.Exec("UPDATE sunblind SET name=?, inputdownpin=?, inputuppin=?, outputdownpin=?, outputuppin=? WHERE id=?", name, inputDownPin, inputUpPin, outputDownPin, outputUpPin, id); err != nil {
		return err
	}

	changeDown := inputDownPin != sunblind.InputDownPin || outputDownPin != sunblind.OutputDownPin
	changeUp := inputUpPin != sunblind.InputUpPin || outputUpPin != sunblind.OutputUpPin
	if changeDown {
		if err := s.gs.UnregisterPinPair(sunblind.InputDownPin, sunblind.OutputDownPin); err != nil {
			return err
		}
	}
	if changeUp {
		if err := s.gs.UnregisterPinPair(sunblind.InputUpPin, sunblind.OutputUpPin); err != nil {
			return err
		}
	}
	if changeDown {
		if err := s.gs.RegisterPinPair(inputDownPin, outputDownPin, enum.PairTypeTimed); err != nil {
			return err
		}
	}
	if changeUp {
		if err := s.gs.RegisterPinPair(inputUpPin, outputUpPin, enum.PairTypeTimed); err != nil {
			return err
		}
	}

	log.Printf("Updated sunblind '%d'\n", id)
	return nil
}

func (s *Service) Delete(id int64) error {
	sunblind, err := getData(id)
	if err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM sunblind WHERE id=?", id); err != nil {
		return err
	}

	if err := s.gs.UnregisterPinPair(sunblind.InputDownPin, sunblind.OutputDownPin); err != nil {
		return err
	}
	if err := s.gs.UnregisterPinPair(sunblind.InputUpPin, sunblind.OutputUpPin); err != nil {
		return err
	}

	log.Printf("Deleted sunblind '%d'\n", id)
	return nil
}

func (s *Service) Toggle(id int64, down bool) error {
	var query string
	if down {
		query = "SELECT inputdownpin FROM sunblind WHERE id=?"
	} else {
		query = "SELECT inputuppin FROM sunblind WHERE id=?"
	}
	var pin domain.Pin
	if err := db.Get(&pin, query, id); err != nil {
		return err
	}
	s.gs.TogglePin(pin)
	return nil
}

func (s *Service) Load() error {
	var sunblinds []*loadData
	err := db.Select(&sunblinds, "SELECT inputdownpin, inputuppin, outputdownpin, outputuppin FROM sunblind")
	if err != nil {
		return err
	}
	for _, sb := range sunblinds {
		if err := s.gs.RegisterPinPair(sb.InputDownPin, sb.OutputDownPin, enum.PairTypeTimed); err != nil {
			return err
		}
		if err := s.gs.RegisterPinPair(sb.InputUpPin, sb.OutputUpPin, enum.PairTypeTimed); err != nil {
			return err
		}
	}
	return nil
}

func getData(id int64) (loadData, error) {
	var sunblind loadData
	if err := db.Get(&sunblind, "SELECT inputdownpin, inputuppin, outputdownpin, outputuppin FROM sunblind WHERE id=?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return loadData{}, fmt.Errorf("Sunblind '%d' does not exist", id)
		}
		return loadData{}, err
	}
	return sunblind, nil
}

type loadData struct {
	InputDownPin  domain.Pin `db:"inputdownpin"`
	InputUpPin    domain.Pin `db:"inputuppin"`
	OutputDownPin domain.Pin `db:"outputdownpin"`
	OutputUpPin   domain.Pin `db:"outputuppin"`
}

func (l loadData) ContainsPin(pin domain.Pin) bool {
	switch pin {
	case l.InputDownPin,
		l.InputUpPin,
		l.OutputDownPin,
		l.OutputUpPin:
		return true
	default:
		return false
	}
}
