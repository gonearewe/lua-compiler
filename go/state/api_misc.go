package state

func (l *luaState) Len(idx int) {
	val := l.stack.get(idx)
	if s, ok := val.(string); ok {
		l.stack.push(int64(len(s)))
	} else {
		panic("length error !")
	}
}
