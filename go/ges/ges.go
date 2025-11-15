package ges

/*
#cgo CFLAGS: -I../../third_party/clt13/include
#cgo LDFLAGS: -L../../third_party/clt13 -lclt13 -lgmp
#include "clt13.h"

// You may need to adjust include paths / library names depending on the repo.
*/
import "C"
import (
    "errors"
    "unsafe"
)

type Context struct {
    state *C.struct_clt_state
    pp    *C.struct_clt_pp
}

type Element struct {
    ctx *Context
    v   *C.struct_clt_elem
}

type Params struct {
    Lambda uint32
    Kappa  uint32
    NZs    uint32
    Pows   []uint32 // top-level index vector
    Flags  uint32
}

func NewContext(p Params) (*Context, error) {
    ctx := &Context{}
    // allocate and init C structs; this will call clt_state_init, clt_pp_init, etc.
    // Pseudocode-ish:
    //
    // C.clt_state_init(&ctx.state, C.uint(p.Lambda), C.uint(p.Kappa), C.uint(p.NZs), ... )
    //
    // You’ll likely follow the examples in clt13/tests.
    //
    return ctx, nil
}

func (c *Context) Free() {
    if c.state != nil {
        C.clt_state_clear(c.state)
        c.state = nil
    }
    if c.pp != nil {
        C.clt_pp_clear(c.pp)
        c.pp = nil
    }
}

func (c *Context) EncodeLevel1(value uint64, index []uint32) (*Element, error) {
    if len(index) != int(c.pp.nzs) { // example; adjust based on struct fields
        return nil, errors.New("bad index length")
    }
    // Convert index slice to C array
    idx := make([]C.uint, len(index))
    for i, v := range index {
        idx[i] = C.uint(v)
    }

    elem := &Element{ctx: c, v: C.clt_elem_new(c.state)}
    rv := C.clt_encode(c.state, c.pp, elem.v, C.ulong(value), (*C.uint)(unsafe.Pointer(&idx[0])))
    if rv != 0 {
        return nil, errors.New("clt_encode failed")
    }
    return elem, nil
}

func (e *Element) Add(other *Element) (*Element, error) {
    if e.ctx != other.ctx {
        return nil, errors.New("different contexts")
    }
    out := &Element{ctx: e.ctx, v: C.clt_elem_new(e.ctx.state)}
    rv := C.clt_elem_add(e.ctx.state, out.v, e.v, other.v)
    if rv != 0 {
        return nil, errors.New("clt_elem_add failed")
    }
    return out, nil
}

func (e *Element) Mul(other *Element) (*Element, error) {
    if e.ctx != other.ctx {
        return nil, errors.New("different contexts")
    }
    out := &Element{ctx: e.ctx, v: C.clt_elem_new(e.ctx.state)}
    rv := C.clt_elem_mul(e.ctx.state, out.v, e.v, other.v)
    if rv != 0 {
        return nil, errors.New("clt_elem_mul failed")
    }
    return out, nil
}

func (e *Element) IsZero() (bool, error) {
    rv := C.clt_zero_test(e.ctx.state, e.ctx.pp, e.v)
    if rv < 0 {
        return false, errors.New("clt_zero_test error")
    }
    return rv == 1, nil
}

func (e *Element) Free() {
    if e.v != nil {
        C.clt_elem_clear(e.ctx.state, e.v)
        e.v = nil
    }
}
