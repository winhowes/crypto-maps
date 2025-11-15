package ges

// Context represents a dummy graded encoding context.
//
// The demos only need enough structure to pretend they are configuring a CLT13
// instance, so we keep the parameters but do not use them for anything beyond
// book-keeping.
type Context struct {
        params Params
}

// Element represents a dummy graded encoding element.
type Element struct{ v int }

// Params collects the knobs the demos like to tweak when instantiating the
// context.  They do not influence the behaviour of the toy implementation but
// keeping them around makes the samples look closer to their real counterparts.
type Params struct {
        Lambda int
        Kappa  int
        NZs    int
        Pows   []uint32
        Flags  int
}

// NewContext returns a new dummy context.
func NewContext(p Params) (*Context, error) {
        return &Context{params: p}, nil
}

// Free mirrors the lifecycle of a real CLT13 context.  There are no resources
// to release, so it is a no-op.
func (c *Context) Free() {}

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
