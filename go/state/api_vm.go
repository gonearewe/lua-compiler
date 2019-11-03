package state

func (l *luaState) PC() int {
	return l.pc
}

func (l *luaState) AddPC(n int) {
	l.pc += n
}

func (l *luaState) Fetch() uint32 {
	i := l.proto.Code[l.pc]
	l.pc++
	return i
}

func (l *luaState) GetConst(idx int) {
	c := l.proto.Constants[idx]
	l.stack.push(c)
}

func (l *luaState) GetRK(rk int) {
	if rk > 0xff { // constant
		l.GetConst(rk & 0xff)
	} else { // register
		l.PushValue(rk + 1)
		// add 1 because the index of luaStack starts with 1
	}
}
