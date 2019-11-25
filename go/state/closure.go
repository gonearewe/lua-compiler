package state

import (
	"github.com/gonearewe/lua-compiler/go/api"
	"github.com/gonearewe/lua-compiler/go/binchunk"
)

type closure struct {
	proto  *binchunk.Prototype // lua closure
	goFunc api.GoFunction      // go closure
	upvals []*upvalue
}

type upvalue struct {
	val *luaValue
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	c := &closure{proto: proto}
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}

	return c
}

func newGoClosure(f api.GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}

	return c
}
