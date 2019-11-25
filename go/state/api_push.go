package state

import (
	"github.com/gonearewe/lua-compiler/go/api"
)

func (l *luaState) PushNil() {
	l.stack.push(nil)
}

func (l *luaState) PushBoolean(b bool) {
	l.stack.push(b)
}

func (l *luaState) PushInteger(n int64) {
	l.stack.push(n)
}

func (l *luaState) PushNumber(n float64) {
	l.stack.push(n)
}

func (l *luaState) PushString(s string) {
	l.stack.push(s)
}

func (l *luaState) PushGoFunction(f api.GoFunction) {
	l.stack.push(newGoClosure(f))
}

func (l *luaState) PushGlobalTable() {
	global := l.registry.get(api.LUA_RIDX_GLOBALS)
	l.stack.push(global)
}
