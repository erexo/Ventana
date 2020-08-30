package gpio

import (
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

type PinManager struct {
	pinPairs   map[domain.Pin]domain.Pin
	openedMpcs map[uint8]*mcp23017.Device

	outputGroup *sync.WaitGroup
	isActive    bool
	gpioOpened  bool
}

func CreatePinManager() *PinManager {
	return &PinManager{
		pinPairs:    make(map[domain.Pin]domain.Pin),
		openedMpcs:  make(map[uint8]*mcp23017.Device),
		outputGroup: &sync.WaitGroup{},
		isActive:    true,
		gpioOpened:  false,
	}
}

func (pm *PinManager) AddPinPair(inputPin, outputPin domain.Pin) error {
	if !pm.isActive {
		return inactiveErr
	}
	for k, v := range pm.pinPairs {
		if inputPin == k || inputPin == v {
			return errors.New(fmt.Sprintf("Pin %v is already in use", inputPin))
		}
		if outputPin == k || outputPin == v {
			return errors.New(fmt.Sprintf("Pin %v is already in use", outputPin))
		}
	}

	err := pm.registerPin(inputPin, enum.PinModeInput)
	err = utils.ConcatErrors(err, pm.registerPin(outputPin, enum.PinModeOutput))
	if err != nil {
		return err
	}

	wi, err := pm.createWorkerPin(inputPin)
	if err != nil {
		return err
	}
	wo, err := pm.createWorkerPin(outputPin)
	if err != nil {
		return err
	}

	go func(group *sync.WaitGroup) {
		group.Add(1)
		for true {
			time.Sleep(checkInterval)
			if !pm.isActive {
				if err = wo.WriteState(defaultPinState); err != nil {
					log.Println("Pin", wo, "inactive write error:", err)
				}
				group.Done()
				break
			}
			v, err := wi.ReadState()
			if err != nil {
				log.Println("Pin", wi, "read error:", err)
			}
			if err = wo.WriteState(v); err != nil {
				log.Println("Pin", wo, "write error:", err)
			}
		}
	}(pm.outputGroup)

	pm.pinPairs[inputPin] = outputPin
	return nil
}

func (pm *PinManager) RemovePinPair(inputPin, outputPin domain.Pin) error {
	if !pm.isActive {
		return inactiveErr
	}
	v, ok := pm.pinPairs[inputPin]
	if !ok {
		return errors.New(fmt.Sprintf("Pin %v is not registered as input pin", inputPin))
	}
	if v != outputPin {
		return errors.New(fmt.Sprintf("Pin %v is not an output pin assigned to input pin %v", outputPin, inputPin))
	}

	// todo, halt gofuncs
	delete(pm.pinPairs, inputPin)
	return nil
}

func (pm *PinManager) Close() error {
	if !pm.isActive {
		return inactiveErr
	}

	var ret error
	for in, out := range pm.pinPairs {
		ret = utils.ConcatErrors(ret, pm.RemovePinPair(in, out))
	}
	pm.isActive = false

	pm.outputGroup.Wait()
	if pm.gpioOpened {
		ret = utils.ConcatErrors(ret, rpio.Close())
		pm.gpioOpened = false
	}

	for _, mpc := range pm.openedMpcs {
		if err := mpc.Close(); err != nil {
			utils.ConcatErrors(ret, err)
		}
	}
	pm.openedMpcs = make(map[uint8]*mcp23017.Device)
	return ret
}

func (pm *PinManager) registerPin(pin domain.Pin, mode enum.PinMode) error {
	if pin.IsMcpPin() {
		mcpNum, err := pin.GetMcpNum()
		if err != nil {
			return err
		}
		mcp, ok := pm.openedMpcs[mcpNum]
		if !ok {
			mcp, err = mcp23017.Open(mcpBus, mcpNum)
			if err != nil {
				return err
			}
			pm.openedMpcs[mcpNum] = mcp
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
		if !pm.gpioOpened {
			if err := rpio.Open(); err != nil {
				return err
			}
			pm.gpioOpened = true
		}
		rp := rpio.Pin(pin.GetPinIndex())
		rpio.PinMode(rp, mode.GetRpioMode())
	}
	return nil
}

func (pm *PinManager) createWorkerPin(pin domain.Pin) (workerPin, error) {
	if pin.IsMcpPin() {
		mcpNum, err := pin.GetMcpNum()
		if err != nil {
			return nil, err
		}
		mcp, ok := pm.openedMpcs[mcpNum]
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
