package state

type luaStack struct {
	/*
	   lua stack : | 1 | 2 | 3 | 4 | 5 |
	   slots     : | 0 | 1 | 2 | 3 | 4 |
	   if top == 3
	   			   |	   |top|	   |
	   			   |  valid    |invalid|
	   			   |     acceptable	   |
	*/
	slots []luaValue
	top   int // index of the top of the lua stack, notice that index starts with 1
}

func newLuaStack(size int) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0, // 0 means it is empty
	}
}

// check if the stack has enough space for n elements,
// if not, it will enlarge the stack to just enough to contain n elements
func (l *luaStack) check(n int) {
	free := len(l.slots) - l.top
	for i := free; i < n; i++ {
		l.slots = append(l.slots, nil)
	}
}

func (l *luaStack) push(val luaValue) {
	if l.top == len(l.slots) {
		panic("stack overflow !")
	}

	l.slots[l.top] = val
	l.top++
}

func (l *luaStack) pop() luaValue {
	if l.top < 1 {
		panic("stack underflow !")
	}

	l.top--
	val := l.slots[l.top]
	l.slots[l.top] = nil

	return val
}

// change the relevant index into absolute one
// it will not check whether the index is valid
// for example, if the stack's top is 5,
// given relevant index -2, returns absolute index 4,
// given relevant index 2, returns absolute index 2
func (l *luaStack) absIndex(idx int) int {
	if idx >= 0 {
		return idx
	}

	return idx + l.top + 1
}

func (l *luaStack) isValid(idx int) bool {
	absIdx := l.absIndex(idx)
	return absIdx > 0 && absIdx <= l.top
}

// get value from the stack by relevant index
// returns nil if the index is invalid
func (l *luaStack) get(idx int) luaValue {
	absIdx := l.absIndex(idx)

	if absIdx > 0 && absIdx <= l.top {
		return l.slots[absIdx-1]
	}

	return nil
}

func (l *luaStack) set(idx int, val luaValue) {
	absIdx := l.absIndex(idx)

	if absIdx > 0 && absIdx <= l.top {
		l.slots[absIdx-1] = val
		return
	}

	panic("invalid index !")
}

// reverse values in range [from,to] upside down
// notice that from and to is the index of slot rather than stack
func (l *luaStack) reverse(from, to int) {
	slots := l.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
