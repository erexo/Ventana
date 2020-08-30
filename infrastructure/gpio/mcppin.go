package gpio

import (
	"fmt"

	"github.com/racerxdl/go-mcp23017"
)

type mcpPin struct {
	pinIndex uint8
	mcpIndex uint8
	mcp      *mcp23017.Device
}

func (p mcpPin) ReadState() (bool, error) {
	level, err := p.mcp.DigitalRead(p.pinIndex)
	if err != nil {
		return false, err
	}
	return bool(level), nil
}

func (p mcpPin) WriteState(state bool) error {
	return p.mcp.DigitalWrite(p.pinIndex, mcp23017.PinLevel(state))
}

func (p mcpPin) String() string {
	return fmt.Sprintf("[%d]%d", p.mcpIndex, p.pinIndex)
}
