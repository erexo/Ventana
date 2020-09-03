package sunblind

import (
	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/core/entity"
	"github.com/Erexo/Ventana/core/enum"
	"github.com/Erexo/Ventana/core/utils"
	"github.com/Erexo/Ventana/infrastructure/gpio"
)

type Service struct {
	pm *gpio.PinManager
}

func CreateService(pm *gpio.PinManager) *Service {
	return &Service{
		pm: pm,
	}
}

func (s *Service) GetSunblind(id int) *entity.Sunblind {
	// todo, load from db
	idp, _ := domain.CreateMcpPin(0, 0)
	iup, _ := domain.CreateMcpPin(0, 1)
	odp, _ := domain.CreateMcpPin(0, 2)
	oup, _ := domain.CreateMcpPin(0, 3)
	return &entity.Sunblind{
		Id:            1,
		Name:          "Test",
		InputDownPin:  idp,
		InputUpPin:    iup,
		OutputDownPin: odp,
		OutputUpPin:   oup,
	}
	//
}

func (s *Service) ToggleSunblind(id int) error {
	sunblind := s.GetSunblind(id)
	if sunblind == nil {
		return utils.NotFoundError(id, "sunblind")
	}
	s.pm.TogglePin(sunblind.InputDownPin)
	s.pm.TogglePin(sunblind.InputUpPin)
	return nil
}

func (s *Service) BrowseSunblinds() []*entity.Sunblind {
	// todo, load from db
	return []*entity.Sunblind{
		s.GetSunblind(0),
	}
	//
}

func (s *Service) Load() error {
	for _, sb := range s.BrowseSunblinds() {
		s.pm.AddPinPair(sb.InputDownPin, sb.OutputDownPin, enum.PairTypeToggle)
		s.pm.AddPinPair(sb.InputUpPin, sb.OutputUpPin, enum.PairTypeTimed)
	}
	return nil
}
