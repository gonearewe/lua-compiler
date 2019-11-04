package vm

import . "github.com/gonearewe/lua-compiler/go/api"

func loadNil(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	vm.PushNil() // just for copy
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i) // copy nil to requested position
	}
	vm.Pop(1) // we don't actually need a nil on the top of the stack
}

func loadBool(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.PushBoolean(b != 0)
	vm.Replace(a)
	if c != 0 {
		vm.AddPC(1)
	}
}
