package state

func (l *luaState) Type(idx int) LuaState {
	if l.stack.isValid(idx) {
		val := l.stack.get(idx)
		return typeOf(val)
	}

	return LUA_TNONE
}
