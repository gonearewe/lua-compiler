package api

type LuaVM interface {
	LuaState
	PC() int          // return the current value of PC, only for debug
	AddPC(n int)      // modify PC, used to achieve JUMP instruction
	Fetch() uint32    // fetch current value of PC, and go to next instruction
	GetConst(idx int) // push requested const into the stack
	GetRK(rk int)     // push requested const or value into the stack
	RegisterCount() int
	LoadVararg(n int)
	LoadProto(idx int)
}
