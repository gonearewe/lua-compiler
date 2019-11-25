/*
APIs of a series of methods with the purpose of getting value from the table
*/
package state

import (
	. "github.com/gonearewe/lua-compiler/go/api"
)

// When you know exactly how much space to allocate,
// avoiding frequent dynamic allocation.
func (l *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	l.stack.push(t)
}

// When you don't know how much space you need, create an empty table
func (l *luaState) NewTable() {
	l.CreateTable(0, 0)
}

// Pop the key out of the stack, look for value in
// the table whose index in the stack is given by idx
// before push it into the stack.
func (l *luaState) GetTable(idx int) LuaType {
	t := l.stack.get(idx) // table
	k := l.stack.pop()    // key
	return l.getTable(t, k)
}

// INPUT: t: table, k: key
func (l *luaState) getTable(t, k luaValue) LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		l.stack.push(v)
		return typeOf(v)
	}

	panic("not a table !")
}

// Get and push the value who belongs to table whose index
// in the stack is given by idx and whose key(type of string) is given by k.
func (l *luaState) GetField(idx int, k string) LuaType {
	t := l.stack.get(idx)
	return l.getTable(t, k)
}

// Get and push the value who belongs to table whose index
// in the stack is given by idx and whose key(type of number) is given by k.
func (l *luaState) GetI(idx int, i int64) LuaType {
	t := l.stack.get(idx)
	return l.getTable(t, i)
}

// Get and push the value who belongs to Global Table and whose key is
// given by name
func (l *luaState) GetGlobal(name string) LuaType {
	t := l.registry.get(LUA_RIDX_GLOBALS)
	return l.getTable(t, name)
}
