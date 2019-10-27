package state

func (l *luaState) GetTop() int {
	return l.stack.top
}

func (l *luaState) AbsIndex(idx int) int {
	return l.stack.absIndex(idx)
}

func (l *luaState) CheckStack(n int) bool {
	l.stack.check(n)

	return true
}

// pop n elements from the stack
func (l *luaState) Pop(n int) {
	for i := 0; i < n; i++ {
		l.stack.pop()
	}
}

// copy value in fromIdx to toIdx
func (l *luaState) Copy(fromIdx, toIdx int) {
	val := l.stack.get(fromIdx)
	l.stack.set(toIdx, val)
}

// push a copy of value in index idx into the stack
func (l *luaState) PushValue(idx int) {
	val := l.stack.get(idx)
	l.stack.push(val)
}

// pop the top element and replace value of index idx with it
func (l *luaState) Replace(idx int) {
	val := l.stack.pop()
	l.stack.set(idx, val)
}

// pop the top element and insert it before index idx,
// then of course, its new index is idx
func (l *luaState) Insert(idx int) {
	l.Rotate(idx, 1)
}

// remove value of index idx, then of course,
// the old idx+1 will become the new idx
func (l *luaState) Remove(idx int) {
	l.Rotate(idx, -1)
	l.Pop(1)
}

// consider index range [idx,top] as a circle, move each element n steps towards top
// if n < 0, it means the other direction
// for example, for stack [a,b,c,d,e,f] (f is the top)
// Rotate(2,1) gives [a,f,b,c,d,e]
// Rotate(2,-1) gives [a,c,d,e,f,b]
func (l *luaState) Rotate(idx, n int) {
	t := l.stack.top - 1           // high slot index
	p := l.stack.absIndex(idx) - 1 //low slot index
	var m int

	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}

	l.stack.reverse(p, m)
	l.stack.reverse(m+1, t)
	l.stack.reverse(p, t)
}

// set the top, pop the redundancy or push nil to achieve this
func (l *luaState) SetTop(idx int) {
	newTop := l.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow !")
	}

	n := l.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			l.stack.pop()
		}
	} else if n < 0 {
		for i := o; i > n; i-- {
			l.stack.push(nil)
		}
	}
}
