package domain

import (
	"errors"
	"math"
)

const (
	McpNumbers       = uint8(8)
	McpPins          = uint8(16)
	firstDeviceIndex = McpNumbers * McpPins
)

type Pin uint8

func CreatePin(pinIndex uint8) (Pin, error) {
	if pinIndex > math.MaxUint8-firstDeviceIndex {
		return 0, errors.New("Invalid pin index")
	}
	return Pin(firstDeviceIndex + pinIndex), nil
}

func CreateMcpPin(mcpNumber, pinIndex uint8) (Pin, error) {
	if mcpNumber >= McpNumbers {
		return 0, errors.New("Invalid mcp number")
	}
	if pinIndex >= McpPins {
		return 0, errors.New("Invalid pin index")
	}
	return Pin(mcpNumber*McpPins + pinIndex), nil
}

func (p Pin) IsMcpPin() bool {
	return uint8(p) < firstDeviceIndex
}

func (p Pin) GetMcpNum() (uint8, error) {
	if p.IsMcpPin() {
		return uint8(p) / McpPins, nil
	}
	return 0, errors.New("Pin is not an MCP pin")
}

func (p Pin) GetPinIndex() uint8 {
	if p.IsMcpPin() {
		return uint8(p) % McpPins
	}
	return uint8(p) - firstDeviceIndex
}
