package state

// Pop value and then pop key, record this pair into
// the table whose index in the stack is given by idx.
func (l *luaState) SetTable(idx int) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	k := l.stack.pop()    // key
	l.setTable(t, k, v)
}

// Record k(key) and v(value) in the given t(table).
func (l *luaState) setTable(t, k, v luaValue) {
	if tbl, ok := t.(*luaTable); ok {
		tbl.put(k, v)
		return
	}

	panic("not a table !")
}

// Pop value and match it to key that is given by k(string),
// record this pair into the table whose index in the stack is given by idx.
func (l *luaState) SetField(idx int, k string) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	l.setTable(t, k, v)
}

// Pop value and match it to key that is given by k(number), record this pair into
// the table whose index in the stack is given by idx.
func (l *luaState) SetI(idx int, k int64) {
	t := l.stack.get(idx) // table
	v := l.stack.pop()    // value
	l.setTable(t, k, v)
}
