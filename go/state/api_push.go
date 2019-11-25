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

// Wrap f(given,GoFunction) to closure and push it into the stack.
func (l *luaState) PushGoFunction(f api.GoFunction) {
	l.stack.push(newGoClosure(f, 0))
}

// Wrap f(given,GoFunction) to closure and push it into the stack
// after pop n luaValue as its upvalues.
func (l *luaState) PushGoClosure(f api.GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := l.stack.pop()
		closure.upvals[n-1] = &upvalue{&val}
	}

	l.stack.push(closure)
}

func (l *luaState) PushGlobalTable() {
	global := l.registry.get(api.LUA_RIDX_GLOBALS)
	l.stack.push(global)
}
