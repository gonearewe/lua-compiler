package state

import "github.com/gonearewe/lua-compiler/go/api"

type luaState struct {
	stack    *luaStack
	registry *luaTable
}

// func New(stackSize int, proto *binchunk.Prototype) *luaState {
// 	return &luaState{
// 		stack: newLuaStack(stackSize),
// 		proto: proto,
// 		pc:    0,
// 	}
// }

func New() *luaState {
	registry := newLuaTable(0, 0)
	registry.put(api.LUA_RIDX_GLOBALS, newLuaTable(0, 0))

	ls := &luaState{
		registry: registry,
	}
	ls.pushLuaStack(newLuaStack(api.LUA_MINSTACK, ls))

	return ls
}

// Add a head node to the linked list.
func (l *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = l.stack
	l.stack = stack
}

// Delete the head node of the linked list.
func (l *luaState) popLuaStack() {
	stack := l.stack
	l.stack = stack.prev
	stack.prev = nil
}
