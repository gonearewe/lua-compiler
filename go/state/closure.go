package state

import (
	"github.com/gonearewe/lua-compiler/go/binchunk"
)

type closure struct {
	proto *binchunk.Prototype
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	return &closure{
		proto: proto,
	}
}
