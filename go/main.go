package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gonearewe/lua-compiler/go/api"

	"github.com/gonearewe/lua-compiler/go/binchunk"
	"github.com/gonearewe/lua-compiler/go/state"
	"github.com/gonearewe/lua-compiler/go/vm"
)

func main() {
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		proto := binchunk.Undump(data)
		luaMain(proto)
	}
}

func luaMain(proto *binchunk.Prototype) {
	nRegs := int(proto.MaxStackSize)
	ls := state.New(nRegs+8, proto)
	ls.SetTop(nRegs)

	for {
		pc := ls.PC()
		inst := vm.Instruction(ls.Fetch())

		if inst.Opcode() != vm.OP_RETURN {
			inst.Execute(ls)
			fmt.Printf("[%02d] %s", pc+1, inst.OpName())
			printStack(ls)
		} else {
			break
		}
	}
}

func printStack(ls api.LuaState) {
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case api.LUA_TBOOLEAN:
			fmt.Printf("[%t]", ls.ToBoolean(i))
		case api.LUA_TNUMBER:
			fmt.Printf("[%g]", ls.ToNumber(i))
		case api.LUA_TSTRING:
			fmt.Printf("[%q]", ls.ToString(i))
		default:
			fmt.Printf("[%s]", ls.TypeName(i))
		}

	}
	fmt.Println()
}
