package state

import (
	"github.com/gonearewe/lua-compiler/api"
)

// Pop value and then pop key, record this pair into
// the table whose index in the stack is given by idx.
func (l *luaState) SetTable(idx int) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	k := l.stack.pop()    // key
	l.setTable(t, k, v, false)
}

// Record k(key) and v(value) in the given t(table).
func (l *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}

	if !raw {
		if mf := getMetafield(t, "__newindex", l); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				l.setTable(x, k, v, false)
				return
			case *closure:
				l.stack.push(mf)
				l.stack.push(t)
				l.stack.push(k)
				l.stack.push(v)
				l.Call(3, 0)
				return
			}
		}
	}

	panic("index error !")
}

// Pop value and match it to key that is given by k(string),
// record this pair into the table whose index in the stack is given by idx.
func (l *luaState) SetField(idx int, k string) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	l.setTable(t, k, v, false)
}

func (self *luaState) RawSet(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, true)
}

func (self *luaState) RawSetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, true)
}

// Pop value and match it to key that is given by k(number), record this pair into
// the table whose index in the stack is given by idx.
func (l *luaState) SetI(idx int, k int64) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	l.setTable(t, k, v, false)
}

// Pop value and match it to key that is given by name,
// record this pair into the Global Table.
func (l *luaState) SetGlobal(name string) {
	t := l.registry.get(api.LUA_RIDX_GLOBALS)
	v := l.stack.pop()
	l.setTable(t, name, v, false)
}

// Register f(GoFunction) with the key given by name in the Global Table.
func (l *luaState) Register(name string, f api.GoFunction) {
	l.PushGoFunction(f)
	l.SetGlobal(name)
}

// Pop a value, if it is nil, delete the metatable of val of index idx(given),
// if it is a table, set it as the metatable of val of index idx(given).
func (l *luaState) SetMetatable(idx int) {
	val := l.stack.get(idx)
	mtVal := l.stack.pop()

	if mtVal == nil {
		setMetatable(val, nil, l)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, l)
	} else {
		panic("table expected !") // TODO
	}
}
