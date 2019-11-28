/*
APIs of a series of methods with the purpose of getting value from the table
*/
package state

import (
	. "github.com/gonearewe/lua-compiler/api"
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
	return l.getTable(t, k, false)
}

// INPUT: t: table, k: key, raw? true=> ignore metamehod, false=> do not ignore metamehod.
func (l *luaState) getTable(t, k luaValue, raw bool) LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			l.stack.push(v)
			return typeOf(v)
		}
	}

	if !raw {
		if mf := getMetafield(t, "__index", l); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return l.getTable(x, k, false)
			case *closure:
				l.stack.push(mf)
				l.stack.push(t)
				l.stack.push(k)
				l.Call(2, 1)
				v := l.stack.get(-1)
				return typeOf(v)
			}
		}
	}

	panic("index error !")
}

// Get and push the value who belongs to table whose index
// in the stack is given by idx and whose key(type of string) is given by k.
func (l *luaState) GetField(idx int, k string) LuaType {
	t := l.stack.get(idx)
	return l.getTable(t, k, false)
}

func (self *luaState) RawGet(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, true)
}

func (self *luaState) RawGetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, true)
}

// Get and push the value who belongs to table whose index
// in the stack is given by idx and whose key(type of number) is given by k.
func (l *luaState) GetI(idx int, i int64) LuaType {
	t := l.stack.get(idx)
	return l.getTable(t, i, false)
}

// Get and push the value who belongs to Global Table and whose key is
// given by name
func (l *luaState) GetGlobal(name string) LuaType {
	t := l.registry.get(LUA_RIDX_GLOBALS)
	return l.getTable(t, name, false)
}

// If val of index idx(given) has a metatable, push the metatable into the stack
// and return true, if not, return false without changing the stack.
func (l *luaState) GetMetatable(idx int) bool {
	val := l.stack.get(idx)

	if mt := getMetatable(val, l); mt != nil {
		l.stack.push(mt)
		return true
	} else {
		return false
	}
}
