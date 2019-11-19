package state

type luaState struct {
	stack *luaStack
}

// func New(stackSize int, proto *binchunk.Prototype) *luaState {
// 	return &luaState{
// 		stack: newLuaStack(stackSize),
// 		proto: proto,
// 		pc:    0,
// 	}
// }

func New() *luaState {
	return &luaState{
		stack: newLuaStack(20),
	}
}

// Add a head node to the linked list.
func (l *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = l.stack
	l.stack = stack
}

// Delete the head node of the linked list.
func (l *luaState) popLuaStack() {
	stack := l.stack
	l.stack = stack.prev
	stack.prev = nil
}
