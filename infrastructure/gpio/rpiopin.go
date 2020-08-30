package gpio

import (
	"strconv"

	"github.com/stianeikeland/go-rpio"
)

type rpioPin uint8

func createRpioPin() rpioPin {
	return 2 //todo
}

func (p rpioPin) ReadState() (bool, error) {
	pin := rpio.Pin(p)
	return rpio.ReadPin(pin) == rpio.High, nil
}

func (p rpioPin) WriteState(state bool) error {
	pin := rpio.Pin(p)
	output := rpio.Low
	if state {
		output = rpio.High
	}
	rpio.WritePin(pin, output)
	return nil
}

func (p rpioPin) String() string {
	return strconv.Itoa(int(p))
}
