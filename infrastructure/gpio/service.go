package gpio

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/core/enum"
	"github.com/Erexo/Ventana/core/utils"
	"github.com/racerxdl/go-mcp23017"
	"github.com/stianeikeland/go-rpio"
)

const (
	mcpBus          = 1
	checkInterval   = 100 * time.Millisecond
	timedPinTime    = 2 * time.Second // todo, move to entities
	defaultPinState = true
)

var (
	inactiveErr = errors.New("Pin Manager is no longer active")
)

type workerPin interface {
	ReadState() (bool, error)
	WriteState(state bool) error
	fmt.Stringer
}

type Service struct {
	pinPairs   map[domain.Pin]*pair
	openedMpcs map[uint8]*mcp23017.Device
	pinMux     sync.Mutex

	outputGroup *sync.WaitGroup
	isActive    bool
	gpioOpened  bool
}

func CreateService() *Service {
	return &Service{
		pinPairs:    make(map[domain.Pin]*pair),
		openedMpcs:  make(map[uint8]*mcp23017.Device),
		outputGroup: &sync.WaitGroup{},
		isActive:    true,
		gpioOpened:  false,
	}
}

type pair struct {
	outputPin          domain.Pin
	pairType           enum.PairType
	timedCancel        context.CancelFunc
	inputState         bool
	outputState        bool
	desiredOutputState bool
	terminated         bool
}

func (s *Service) IsPinRegistered(pins ...domain.Pin) error {
	s.pinMux.Lock()
	defer s.pinMux.Unlock()
	for k, v := range s.pinPairs {
		for _, p := range pins {
			if p == k || p == v.outputPin {
				return fmt.Errorf("Pin %v is already in use", p)
			}
		}
	}
	return nil
}

func (s *Service) RegisterPinPair(inputPin, outputPin domain.Pin, pairType enum.PairType) error {
	if !s.isActive {
		return inactiveErr
	}
	if err := s.IsPinRegistered(inputPin, outputPin); err != nil {
		return err
	}

	s.pinMux.Lock()
	defer s.pinMux.Unlock()
	err := s.registerPin(inputPin, enum.PinModeInput)
	err = utils.ConcatErrors(err, s.registerPin(outputPin, enum.PinModeOutput))
	if err != nil {
		return err
	}

	wi, err := s.createWorkerPin(inputPin)
	if err != nil {
		return err
	}
	wo, err := s.createWorkerPin(outputPin)
	if err != nil {
		return err
	}

	p := createPair(outputPin, pairType)
	switch pairType {
	case enum.PairTypeToggle:
		go s.togglePairWorker(wi, wo, p)
	case enum.PairTypeTimed:
		go s.timedPairWorker(wi, wo, p)
	default:
		log.Println("Unknown PairType:", pairType)
	}
	s.pinPairs[inputPin] = p
	return nil
}

func (s *Service) UnregisterPinPair(inputPin, outputPin domain.Pin) error {
	if !s.isActive {
		return inactiveErr
	}
	s.pinMux.Lock()
	defer s.pinMux.Unlock()
	return s.internalUnregisterPinPair(inputPin, outputPin)
}

func (s *Service) TogglePin(inputPin domain.Pin) error {
	if !s.isActive {
		return inactiveErr
	}
	s.pinMux.Lock()
	defer s.pinMux.Unlock()
	p, ok := s.pinPairs[inputPin]
	if !ok {
		return errors.New(fmt.Sprintf("Pin %v is not registered as input pin", inputPin))
	}
	switch p.pairType {
	case enum.PairTypeToggle:
		p.desiredOutputState = !p.desiredOutputState
	case enum.PairTypeTimed:
		p.desiredOutputState = !defaultPinState
		go func() {
			if p.timedCancel != nil {
				p.timedCancel()
			}
			ctx, f := context.WithCancel(context.Background())
			p.timedCancel = f
			time.Sleep(timedPinTime)
			select {
			case <-ctx.Done():
			default:
				p.desiredOutputState = defaultPinState
			}
		}()
	default:
		log.Println("Unknown PairType:", p.pairType)
	}
	return nil
}

func (s *Service) Close() error {
	if !s.isActive {
		return inactiveErr
	}
	s.pinMux.Lock()
	defer s.pinMux.Unlock()

	var ret error
	for in, out := range s.pinPairs {
		ret = utils.ConcatErrors(ret, s.internalUnregisterPinPair(in, out.outputPin))
	}
	s.isActive = false

	s.outputGroup.Wait()
	if s.gpioOpened {
		ret = utils.ConcatErrors(ret, rpio.Close())
		s.gpioOpened = false
	}

	for _, mpc := range s.openedMpcs {
		if err := mpc.Close(); err != nil {
			utils.ConcatErrors(ret, err)
		}
	}
	s.openedMpcs = make(map[uint8]*mcp23017.Device)
	return ret
}

func (s *Service) registerPin(pin domain.Pin, mode enum.PinMode) error {
	if pin.IsMcpPin() {
		mcpNum, err := pin.GetMcpNum()
		if err != nil {
			return err
		}
		mcp, ok := s.openedMpcs[mcpNum]
		if !ok {
			mcp, err = mcp23017.Open(mcpBus, mcpNum)
			if err != nil {
				return err
			}
			s.openedMpcs[mcpNum] = mcp
		}
		index := pin.GetPinIndex()
		err = mcp.PinMode(index, mode.GetMcpMode())
		if err != nil {
			return err
		}
		if mode == enum.PinModeInput {
			err = mcp.SetPullUp(index, true)
			if err != nil {
				return err
			}
		}
	} else {
		if !s.gpioOpened {
			if err := rpio.Open(); err != nil {
				return err
			}
			s.gpioOpened = true
		}
		rp := rpio.Pin(pin.GetPinIndex())
		rpio.PinMode(rp, mode.GetRpioMode())
	}
	return nil
}

func (s *Service) internalUnregisterPinPair(inputPin, outputPin domain.Pin) error {
	v, ok := s.pinPairs[inputPin]
	if !ok {
		return errors.New(fmt.Sprintf("Pin %v is not registered as input pin", inputPin))
	}
	if v.outputPin != outputPin {
		return errors.New(fmt.Sprintf("Pin %v is not an output pin assigned to input pin %v", outputPin, inputPin))
	}

	v.terminated = true
	delete(s.pinPairs, inputPin)
	return nil
}

func (s *Service) createWorkerPin(pin domain.Pin) (workerPin, error) {
	if pin.IsMcpPin() {
		mcpNum, err := pin.GetMcpNum()
		if err != nil {
			return nil, err
		}
		mcp, ok := s.openedMpcs[mcpNum]
		if !ok {
			return nil, errors.New(fmt.Sprintf("Mcp %v is not opened", mcpNum))
		}

		return mcpPin{
			pinIndex: pin.GetPinIndex(),
			mcpIndex: mcpNum,
			mcp:      mcp,
		}, nil
	}
	return rpioPin(pin.GetPinIndex()), nil
}

func (s *Service) togglePairWorker(wi, wo workerPin, p *pair) {
	s.outputGroup.Add(1)
	defer s.outputGroup.Done()
	var err error
	for true {
		time.Sleep(checkInterval)
		if !s.isActive || p.terminated {
			if err = wo.WriteState(defaultPinState); err != nil {
				log.Println("Pin", wo, "inactive write error:", err)
			}
			break
		}
		if p.desiredOutputState == p.outputState {
			v, err := wi.ReadState()
			if err != nil {
				log.Println("Pin", wi, "read error:", err)
				continue
			} else if v == p.inputState {
				continue
			}
			p.inputState = v
			p.desiredOutputState = !p.outputState
		}
		if err = wo.WriteState(p.desiredOutputState); err != nil {
			log.Println("Pin", wo, "write error:", err)
		} else {
			p.outputState = p.desiredOutputState
		}
	}
}

func (s *Service) timedPairWorker(wi, wo workerPin, p *pair) {
	s.outputGroup.Add(1)
	defer s.outputGroup.Done()
	var err error
	for true {
		time.Sleep(checkInterval)
		if !s.isActive || p.terminated {
			if err = wo.WriteState(defaultPinState); err != nil {
				log.Println("Pin", wo, "inactive write error:", err)
			}
			break
		}
		if p.desiredOutputState != defaultPinState {
			if p.outputState == defaultPinState {
				if err = wo.WriteState(!p.outputState); err != nil {
					log.Println("Pin", wo, "write error:", err)
				} else {
					p.outputState = !p.outputState
				}
			}
		} else {
			v, err := wi.ReadState()
			if err != nil {
				log.Println("Pin", wi, "read error:", err)
				continue
			}
			p.inputState = v
			if p.inputState != p.outputState {
				if err = wo.WriteState(p.inputState); err != nil {
					log.Println("Pin", wo, "write error:", err)
				} else {
					p.outputState = p.inputState
				}
			}
		}
	}
}

func createPair(outputPin domain.Pin, pairType enum.PairType) *pair {
	return &pair{
		outputPin:          outputPin,
		pairType:           pairType,
		inputState:         defaultPinState,
		outputState:        defaultPinState,
		desiredOutputState: defaultPinState,
		terminated:         false,
	}
}
