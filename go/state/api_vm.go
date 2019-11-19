package state

func (l *luaState) PC() int {
	return l.stack.pc
}

func (l *luaState) AddPC(n int) {
	l.stack.pc += n
}

func (l *luaState) Fetch() uint32 {
	i := l.stack.closure.proto.Code[l.pc]
	l.stack.pc++
	return i
}

// Push a constant into the stack whose index in the
// constant list is given by idx
func (l *luaState) GetConst(idx int) {
	c := l.stack.closure.proto.Constants[idx]
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
