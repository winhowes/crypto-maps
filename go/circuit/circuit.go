package circuit

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
