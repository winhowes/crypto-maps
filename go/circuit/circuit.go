package circuit

import "fmt"

type GateType int

const (
	GateInput GateType = iota
	GateAnd
	GateOr
	GateNot
)

type Gate struct {
	Type GateType
	In1  int // index of first input wire (or -1)
	In2  int // index of second input wire (or -1)
}

type Circuit struct {
	NumInputs  int
	Gates      []Gate
	OutputGate int
}

// Example: (a AND b)
func ExampleAndCircuit() *Circuit {
	// inputs are implicit: 0,1 are input wires
	// gate 0: AND(a,b)
	return &Circuit{
		NumInputs: 2,
		Gates: []Gate{
			{Type: GateAnd, In1: 0, In2: 1}, // gate index 0
		},
		OutputGate: 0,
	}
}

// Eval evaluates the circuit in plain Go for sanity checking.
func Eval(c *Circuit, in []byte) (byte, error) {
	if c == nil {
		return 0, fmt.Errorf("nil circuit")
	}
	if len(in) != c.NumInputs {
		return 0, fmt.Errorf("input length mismatch: got %d want %d", len(in), c.NumInputs)
	}
	if c.OutputGate < 0 || c.OutputGate >= len(c.Gates) {
		return 0, fmt.Errorf("invalid output gate index %d", c.OutputGate)
	}

	totalWires := c.NumInputs + len(c.Gates)
	wires := make([]byte, totalWires)
	copy(wires, in)

	for i, g := range c.Gates {
		outIdx := c.NumInputs + i
		// helper to load a wire that has already been computed
		loadWire := func(idx int) (byte, error) {
			if idx < 0 || idx >= c.NumInputs+i {
				return 0, fmt.Errorf("gate %d references invalid wire %d", i, idx)
			}
			return wires[idx], nil
		}

		switch g.Type {
		case GateInput:
			val, err := loadWire(g.In1)
			if err != nil {
				return 0, err
			}
			wires[outIdx] = val
		case GateAnd:
			a, err := loadWire(g.In1)
			if err != nil {
				return 0, err
			}
			b, err := loadWire(g.In2)
			if err != nil {
				return 0, err
			}
			wires[outIdx] = a & b
		case GateOr:
			a, err := loadWire(g.In1)
			if err != nil {
				return 0, err
			}
			b, err := loadWire(g.In2)
			if err != nil {
				return 0, err
			}
			wires[outIdx] = a | b
		case GateNot:
			a, err := loadWire(g.In1)
			if err != nil {
				return 0, err
			}
			if a == 0 {
				wires[outIdx] = 1
			} else {
				wires[outIdx] = 0
			}
		default:
			return 0, fmt.Errorf("gate %d has unknown type %d", i, g.Type)
		}
	}

	return wires[c.NumInputs+c.OutputGate], nil
}
