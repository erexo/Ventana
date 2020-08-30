package enum

import (
	"github.com/racerxdl/go-mcp23017"
	"github.com/stianeikeland/go-rpio"
)

const (
	PinModeInput PinMode = iota
	PinModeOutput
)

type PinMode uint8

func (p PinMode) GetRpioMode() rpio.Mode {
	if p == PinModeInput {
		return rpio.Input
	}
	return rpio.Output
}

func (p PinMode) GetMcpMode() mcp23017.PinMode {
	if p == PinModeInput {
		return mcp23017.INPUT
	}
	return mcp23017.OUTPUT
}
