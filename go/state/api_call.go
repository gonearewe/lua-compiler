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

	return api.LUA_OK
}

// Call LuaClosure, GoClosure or Metamethod with nArgs(number of args)
// and nResults(number of requesting return values).
// EXAMPLE1: for stack[5,2,fun,45,13], Call(2,1) calls fun(45,13) requesting one return value,
// after that, the stack is [5,2,<returnValue 1>].
// EXAMPLE2: for stack[5,2,val,45,13], val is not a closure but has a metamehod mf,
// Call(2,1) calls mf(val,45,13) requesting one return value,
// after that, the stack is [5,2,<returnValue 1>].
func (l *luaState) Call(nArgs, nResults int) {
	val := l.stack.get(-(nArgs + 1))

	c, ok := val.(*closure)
	if !ok { // support metamethod
		if mf := getMetafield(val, "__call", l); mf != nil {
			if c, ok = mf.(*closure); ok { // ok can be modified here
				l.stack.push(val)
				l.Insert(-(nArgs + 2))
				nArgs += 1
				// for stack[5,2,val,45,13], val is not a closure but has a metamehod mf,
				// now the stack is [5,2,mf,val,45,13] and args include val, 45 and 13
			}
		}
	}

	if ok { // call closure or metamethod
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

// Call a function that is able to throw an exception, offer exception
// catching and handling support, refer to Call() for details of basic function calling.
func (l *luaState) PCall(nArgs, nResults int, msgh int) (status int) {
	caller := l.stack
	status = api.LUA_ERRRUN

	// catch error
	defer func() {
		if err := recover(); err != nil {
			// VM recovered, but luaStack remains where exception occurs
			for l.stack != caller {
				l.popLuaStack() // roll back to safe luaStack where pcall() is waiting.
			}

			l.stack.push(err)
		}
	}()

	l.Call(nArgs, nResults)
	status = api.LUA_OK // no exceptions, defer func not excuated

	return
}

func (l *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize) // number of registers
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	newStack := newLuaStack(nRegs+api.LUA_MINSTACK, l)
	newStack.closure = c

	// pass parameters to the called function
	funcAndArgs := l.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams) // funcAndArgs[0] is the closure
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
