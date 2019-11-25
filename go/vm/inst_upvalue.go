package vm

import "github.com/gonearewe/lua-compiler/go/api"

func getUpval(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(api.LuaUpvalueIndex(b), a)
}

func setUpval(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(a, api.LuaUpvalueIndex(b))
}

// Replace register of index a with value whose key's(in registers or constants list)
// index is c, and the map is a table whose index in the upvalues is b.
func getTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.GetRK(c)
	vm.GetTable(api.LuaUpvalueIndex(b))
	vm.Replace(a)
}

func setTabUp(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(api.LuaUpvalueIndex(a))
}
