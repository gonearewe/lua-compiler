package state

// push the length of the string at given index into the luaStack
func (l *luaState) Len(idx int) {
	val := l.stack.get(idx)
	if s, ok := val.(string); ok {
		l.stack.push(int64(len(s)))
	} else if result, ok := callMetamethod(val, val, "__len", l); ok {
		l.stack.push(result)
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

			b := l.stack.pop()
			a := l.stack.pop()
			if result, ok := callMetamethod(a, b, "__concat", l); ok {
				l.stack.push(result)
				continue
			}

			panic("concatenation error !")
		}
	}
	// if n == 1, do nothing
}

// Api for iterator, the luaTable's index in the stack is idx, pop current key and
// push next key, return false if the luaTable is empty or iterating
// is over in which case nothing is pushed.
func (l *luaState) Next(idx int) bool {
	val := l.stack.get(idx)
	if t, ok := val.(*luaTable); ok {
		key := l.stack.pop()
		if nextKey := t.nextKey(key); nextKey != nil {
			l.stack.push(nextKey)
			l.stack.push(t.get(nextKey))

			return true
		}

		return false
	}

	panic("table expected !")
}
