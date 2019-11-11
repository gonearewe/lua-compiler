package state

// push the length of the string at given index into the luaStack
func (l *luaState) Len(idx int) {
	val := l.stack.get(idx)
	if s, ok := val.(string); ok {
		l.stack.push(int64(len(s)))
	} else if t, ok := val.(*luaTable); ok {
		l.stack.push(int64(t.len()))
	} else {
		panic("length error !")
	}
}

// pop n elements from the luaStack and converse them to string
// before concat them, push the result back
func (l *luaState) Concat(n int) {
	if n == 0 {
		l.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if l.IsString(-1) && l.IsString(-2) {
				s2 := l.ToString(-1)
				s1 := l.ToString(-2)

				l.stack.pop()
				l.stack.pop()
				l.stack.push(s1 + s2)

				continue
			}

			panic("concatenation error !")
		}
	}
	// if n == 1, do nothing
}
