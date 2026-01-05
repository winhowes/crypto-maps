package ges

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/clt13/src -I${SRCDIR}/../../third_party/libaesrand -O2 -fopenmp -Wno-error=deprecated-declarations
#cgo LDFLAGS: -lgmp -lcrypto -lm -fopenmp
#include <stdlib.h>
#include "clt13.h"

int go_clt_encode_level1(clt_elem_t *rop, const clt_state_t *state, unsigned long long value, const int *ix);
clt_state_t *go_clt_state_new(const clt_params_t *params, size_t flags);
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// Context represents a CLT13 graded encoding context backed by the C implementation.
type Context struct {
	params Params
	state  *C.clt_state_t
	pp     *C.clt_pp_t
}

// Element represents a graded encoding element backed by the C implementation.
type Element struct {
	ctx  *Context
	elem *C.clt_elem_t
}

// Params collects the knobs the demos like to tweak when instantiating the
// context.
type Params struct {
	Lambda int
	Kappa  int
	NZs    int
	Pows   []uint32
	Flags  int
}

// NewContext instantiates a CLT13 context using the provided parameters.
func NewContext(p Params) (*Context, error) {
	if p.NZs <= 0 {
		return nil, errors.New("nzs must be positive")
	}
	if len(p.Pows) != p.NZs {
		return nil, fmt.Errorf("expected %d pows entries, got %d", p.NZs, len(p.Pows))
	}

	cPows, freeFn, err := makeCIntArray(p.Pows)
	if err != nil {
		return nil, err
	}
	defer freeFn()

	var params C.clt_params_t
	params.lambda = C.size_t(p.Lambda)
	params.kappa = C.size_t(p.Kappa)
	params.nzs = C.size_t(p.NZs)
	params.pows = cPows

	state := C.go_clt_state_new(&params, C.size_t(p.Flags))
	if state == nil {
		return nil, errors.New("clt_state_new failed")
	}

	pp := C.clt_pp_new(state)
	if pp == nil {
		C.clt_state_free(state)
		return nil, errors.New("clt_pp_new failed")
	}

	ctx := &Context{params: p, state: state, pp: pp}
	runtime.SetFinalizer(ctx, (*Context).finalize)
	return ctx, nil
}

func (c *Context) finalize() {
	c.Free()
}

// Free releases the underlying C resources.
func (c *Context) Free() {
	if c == nil {
		return
	}
	if c.pp != nil {
		C.clt_pp_free(c.pp)
		c.pp = nil
	}
	if c.state != nil {
		C.clt_state_free(c.state)
		c.state = nil
	}
}

// EncodeLevel1 returns an Element containing the provided value at level 1.
func (c *Context) EncodeLevel1(value uint64, index []uint32) (*Element, error) {
	if c == nil || c.state == nil || c.pp == nil {
		return nil, errors.New("context not initialized")
	}
	if len(index) != c.params.NZs {
		return nil, fmt.Errorf("expected %d index entries, got %d", c.params.NZs, len(index))
	}

	elem, err := newElement(c)
	if err != nil {
		return nil, err
	}

	ix := make([]C.int, len(index))
	for i, v := range index {
		ix[i] = C.int(v)
	}
	var ixPtr *C.int
	if len(ix) > 0 {
		ixPtr = (*C.int)(unsafe.Pointer(&ix[0]))
	}

	if rc := C.go_clt_encode_level1(elem.elem, c.state, C.ulonglong(value), ixPtr); rc != C.CLT_OK {
		elem.Free()
		return nil, fmt.Errorf("clt_encode failed: %d", int(rc))
	}
	return elem, nil
}

func newElement(ctx *Context) (*Element, error) {
	ptr := C.clt_elem_new()
	if ptr == nil {
		return nil, errors.New("clt_elem_new failed")
	}
	elem := &Element{ctx: ctx, elem: ptr}
	runtime.SetFinalizer(elem, (*Element).finalize)
	return elem, nil
}

func (e *Element) finalize() {
	e.Free()
}

// Free releases the underlying C allocation for the element.
func (e *Element) Free() {
	if e == nil || e.elem == nil {
		return
	}
	C.clt_elem_free(e.elem)
	e.elem = nil
}

// Add returns a new Element representing the sum of two elements.
func (e *Element) Add(other *Element) (*Element, error) {
	if err := ensureSameContext(e, other); err != nil {
		return nil, err
	}
	result, err := newElement(e.ctx)
	if err != nil {
		return nil, err
	}
	if rc := C.clt_elem_add(result.elem, e.ctx.pp, e.elem, other.elem); rc != C.CLT_OK {
		result.Free()
		return nil, fmt.Errorf("clt_elem_add failed: %d", int(rc))
	}
	return result, nil
}

// Mul returns a new Element representing the product of two elements.
func (e *Element) Mul(other *Element) (*Element, error) {
	if err := ensureSameContext(e, other); err != nil {
		return nil, err
	}
	result, err := newElement(e.ctx)
	if err != nil {
		return nil, err
	}
	if rc := C.clt_elem_mul(result.elem, e.ctx.pp, e.elem, other.elem); rc != C.CLT_OK {
		result.Free()
		return nil, fmt.Errorf("clt_elem_mul failed: %d", int(rc))
	}
	return result, nil
}

// IsZero reports whether the element is zero.
func (e *Element) IsZero() (bool, error) {
	if e == nil || e.elem == nil {
		return false, errors.New("element not initialized")
	}
	if e.ctx == nil || e.ctx.pp == nil {
		return false, errors.New("element context not available")
	}
	res := C.clt_is_zero(e.elem, e.ctx.pp)
	if res == C.CLT_ERR {
		return false, errors.New("clt_is_zero failed")
	}
	return res != 0, nil
}

func ensureSameContext(a, b *Element) error {
	if a == nil || b == nil {
		return errors.New("elements must be non-nil")
	}
	if a.elem == nil || b.elem == nil {
		return errors.New("elements must be initialized")
	}
	if a.ctx == nil || b.ctx == nil || a.ctx != b.ctx {
		return errors.New("elements belong to different contexts")
	}
	return nil
}

func makeCIntArray(values []uint32) (*C.int, func(), error) {
	if len(values) == 0 {
		return nil, func() {}, nil
	}
	size := C.size_t(len(values)) * C.size_t(unsafe.Sizeof(C.int(0)))
	ptr := C.malloc(size)
	if ptr == nil {
		return nil, func() {}, errors.New("malloc failed")
	}
	slice := (*[1 << 30]C.int)(ptr)[:len(values):len(values)]
	for i, v := range values {
		slice[i] = C.int(v)
	}
	freeFn := func() {
		C.free(ptr)
	}
	return (*C.int)(ptr), freeFn, nil
}
