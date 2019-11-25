package state

import (
	"github.com/gonearewe/lua-compiler/go/api"
)

type luaStack struct {
	/*
	   lua stack : | 1 | 2 | 3 | 4 | 5 |
	   slots     : | 0 | 1 | 2 | 3 | 4 |
	   if top == 3
	   			   |	   |top|	   |
	   			   |  valid    |invalid|
	   			   |     acceptable	   |
	*/
	slots []luaValue
	top   int // index of the top of the lua stack, notice that index starts with 1

	prev    *luaStack // use linked list to achieve function call-back stack
	closure *closure
	openuvs map[int]*upvalue
	varargs []luaValue
	pc      int

	state *luaState
}

func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0, // 0 means it is empty
		state: state,
	}
}

// check if the stack has enough space for n elements,
// if not, it will enlarge the stack to just enough to contain n elements
func (l *luaStack) check(n int) {
	free := len(l.slots) - l.top
	for i := free; i < n; i++ {
		l.slots = append(l.slots, nil)
	}
}

func (l *luaStack) push(val luaValue) {
	if l.top == len(l.slots) {
		panic("stack overflow !")
	}

	l.slots[l.top] = val
	l.top++
}

func (l *luaStack) pop() luaValue {
	if l.top < 1 {
		panic("stack underflow !")
	}

	l.top--
	val := l.slots[l.top]
	l.slots[l.top] = nil

	return val
}

// change the relevant index into absolute one
// it will not check whether the index is valid
// for example, if the stack's top is 5,
// given relevant index -2, returns absolute index 4,
// given relevant index 2, returns absolute index 2
func (l *luaStack) absIndex(idx int) int {
	if idx <= api.LUA_REGISTRYINDEX {
		return idx // support fake index
	}

	if idx >= 0 {
		return idx
	}

	return idx + l.top + 1
}

func (l *luaStack) isValid(idx int) bool {
	if idx == api.LUA_REGISTRYINDEX {
		return true
	}

	if idx < api.LUA_REGISTRYINDEX { // upvalues
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := l.closure

		return c != nil && uvIdx < len(c.upvals)
	}

	absIdx := l.absIndex(idx)
	return absIdx > 0 && absIdx <= l.top
}

// Get value from the stack by relevant index,
// returns the registry if idx is fake index of registry
// returns nil if the index is invalid.
func (l *luaStack) get(idx int) luaValue {
	if idx == api.LUA_REGISTRYINDEX {
		return l.state.registry
	}

	if idx < api.LUA_REGISTRYINDEX { // upvalues
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := l.closure
		if c == nil || uvIdx >= len(c.upvals) { // invalid index
			return nil
		}

		return *(c.upvals[uvIdx].val)
	}

	absIdx := l.absIndex(idx)

	if absIdx > 0 && absIdx <= l.top {
		return l.slots[absIdx-1]
	}

	return nil
}

func (l *luaStack) set(idx int, val luaValue) {
	if idx == api.LUA_REGISTRYINDEX {
		l.state.registry = val.(*luaTable)
		return
	}

	if idx < api.LUA_REGISTRYINDEX { // upvalues
		uvIdx := api.LUA_REGISTRYINDEX - idx - 1
		c := l.closure
		if c != nil && uvIdx < len(c.upvals) { // valid index
			*(c.upvals[uvIdx].val) = val
		}

		return
	}

	absIdx := l.absIndex(idx)

	if absIdx > 0 && absIdx <= l.top {
		l.slots[absIdx-1] = val
		return
	}

	panic("invalid index !")
}

// reverse values in range [from,to] upside down
// notice that from and to is the index of slot rather than stack
func (l *luaStack) reverse(from, to int) {
	slots := l.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

// Returns a slice of n(given) luaValue popped from the stack.
func (l *luaStack) popN(n int) []luaValue {
	// allocate first when we know how much space we need
	vals := make([]luaValue, n)

	for i := n - 1; i >= 0; i-- {
		vals[i] = l.pop()
	}

	return vals
}

// Push n(given) luaValue from the slice vals(given) into the stack,
// nil is pushed when vals(given) can't offer enough number of luaValue,
// redundant luaValue are abandoned when vals(given) offers too many.
func (l *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}

	for i := 0; i < n; i++ {
		if i < nVals {
			l.push(vals[i])
		} else {
			l.push(nil)
		}
	}
}
