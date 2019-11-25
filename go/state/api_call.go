package state

import (
	"github.com/gonearewe/lua-compiler/go/api"
	"github.com/gonearewe/lua-compiler/go/vm"

	"github.com/gonearewe/lua-compiler/go/binchunk"
)

func (l *luaState) Load(chunk []byte, chunkName, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	l.stack.push(c)

	if len(proto.Upvalues) > 0 { // set _ENV
		env := l.registry.get(api.LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}

	return 0
}

func (l *luaState) Call(nArgs, nResults int) {
	val := l.stack.get(-(nArgs + 1))
	if c, ok := val.(*closure); ok {
		if c.proto != nil {
			l.callLuaClosure(nArgs, nResults, c)
		} else {
			l.callGoClosure(nArgs, nResults, c)
		}
		// DEBUG
		// fmt.Printf(
		// 	"call %s<%d,%d>\n",
		// 	c.proto.Source,
		// 	c.proto.LineDefined,
		// 	c.proto.LastLineDefined,
		// )

	} else {
		panic("not function !")
	}
}

func (l *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize) // number of registers
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	newStack := newLuaStack(nRegs+api.LUA_MINSTACK, l)
	newStack.closure = c

	// pass parameters to the called function
	funcAndArgs := l.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	// call the function and run
	l.pushLuaStack(newStack)
	l.runLuaClosure()
	l.popLuaStack()

	// if the called function returns value
	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		l.stack.check(len(results))
		l.stack.pushN(results, nResults)
	}
}

func (l *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(l.Fetch())
		inst.Execute(l)

		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}

func (l *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+api.LUA_MINSTACK, l)
	newStack.closure = c

	args := l.stack.popN(nArgs)
	newStack.pushN(args, nArgs)
	l.stack.pop() // desert goClosure

	l.pushLuaStack(newStack) // call
	r := c.goFunc(l)         // execuate goFunc
	l.popLuaStack()          // return

	if nResults != 0 { // push return values if any
		results := newStack.popN(r)
		l.stack.check(len(results))
		l.stack.pushN(results, nResults)
	}
}
