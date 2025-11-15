package ges

// Context represents a dummy graded encoding context.
type Context struct{}

// Element represents a dummy graded encoding element.
type Element struct{ v int }

// Params is an empty placeholder for future parameters.
type Params struct{}

// NewContext returns a new dummy context.
func NewContext(p Params) (*Context, error) {
	return &Context{}, nil
}

// EncodeLevel1 returns an Element containing the provided value.
func (c *Context) EncodeLevel1(value uint64, index []uint32) (*Element, error) {
	return &Element{v: int(value)}, nil
}

// Add returns a new Element representing the sum of two elements.
func (e *Element) Add(other *Element) (*Element, error) {
	return &Element{v: e.v + other.v}, nil
}

// Mul returns a new Element representing the product of two elements.
func (e *Element) Mul(other *Element) (*Element, error) {
	return &Element{v: e.v * other.v}, nil
}

// IsZero reports whether the element is zero.
func (e *Element) IsZero() (bool, error) {
	return e.v == 0, nil
}
