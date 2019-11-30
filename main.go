package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gonearewe/lua-compiler/api"
	"github.com/gonearewe/lua-compiler/compiler/lexer"
)

func main() {
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}

		testLexer(string(data), os.Args[1])
	}
}

func testLexer(chunk, chunkName string) {
	_lexer := lexer.NewLexer(chunk, chunkName)
	for {
		line, kind, token := _lexer.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n", line, kindToCategory(kind), token)
		if kind == lexer.TOKEN_EOF {
			break
		}
	}
}

func kindToCategory(kind int) string {
	switch {
	case kind < lexer.TOKEN_SEP_SEMI:
		return "other"
	case kind <= lexer.TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= lexer.TOKEN_OP_NOT:
		return "operator"
	case kind <= lexer.TOKEN_KW_WHILE:
		return "keyword"
	case kind == lexer.TOKEN_IDENTIFIER:
		return "identifier"
	case kind == lexer.TOKEN_NUMBER:
		return "number"
	case kind == lexer.TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}

// func main() {
// 	if len(os.Args) > 1 {
// 		data, err := ioutil.ReadFile(os.Args[1])
// 		if err != nil {
// 			panic(err)
// 		}

// 		ls := state.New()
// 		ls.Register("print", print)
// 		ls.Register("error", error)
// 		ls.Register("pcall", pCall)
// 		ls.Load(data, os.Args[1], "b")
// 		ls.Call(0, 0)
// 	}
// }

// func error(ls api.LuaState) int {
// 	return ls.Error()
// }

// func pCall(ls api.LuaState) int {
// 	nArgs := ls.GetTop() - 1
// 	status := ls.PCall(nArgs, -1, 0)
// 	ls.PushBoolean(status == api.LUA_OK)
// 	ls.Insert(1)

// 	return ls.GetTop()
// }

// func next(ls api.LuaState) int {
// 	ls.SetTop(2)
// 	if ls.Next(1) {
// 		return 2
// 	} else {
// 		ls.PushNil()
// 		return 1
// 	}
// }

// func pairs(l api.LuaState) int {
// 	l.PushGoFunction(next)
// 	l.PushValue(1)
// 	l.PushNil()

// 	return 3
// }

// func iPairs(l api.LuaState) int {
// 	l.PushGoFunction(_iPairsAux)
// 	l.PushValue(1)
// 	l.PushInteger(0)

// 	return 3
// }

// func _iPairsAux(l api.LuaState) int {
// 	i := l.ToInteger(2) + 1
// 	l.PushInteger(i)
// 	if l.GetI(1, i) == api.LUA_TNIL {
// 		return 1
// 	} else {
// 		return 2
// 	}
// }

// func getMetatable(ls api.LuaState) int {
// 	if !ls.GetMetatable(1) {
// 		ls.PushNil()
// 	}

// 	return 1
// }

// func setMetatable(ls api.LuaState) int {
// 	ls.SetMetatable(1)
// 	return 1
// }

func print(ls api.LuaState) int {
	nArgs := ls.GetTop()
	for i := 1; i <= nArgs; i++ {
		if ls.IsBoolean(i) {
			fmt.Printf("%t", ls.ToBoolean(i))
		} else if ls.IsString(i) {
			fmt.Print(ls.ToString(i))
		} else {
			fmt.Print(ls.TypeName(ls.Type(i)))
		}

		if i < nArgs {
			fmt.Print("\t")
		}
	}

	fmt.Println()

	return 0
}

/*
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

// Test Arith Functions
func Calculator() {
	ls := state.New(16, nil)
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)
	printStack(ls)

	ls.Arith(api.LUA_OPADD)
	printStack(ls)
	ls.Arith(api.LUA_OPBNOT)
	printStack(ls)
	ls.Len(2)
	printStack(ls)
	ls.Concat(3)
	printStack(ls)
	ls.PushBoolean(ls.Compare(1, 2, api.LUA_OPEQ))
	printStack(ls)
}

// Test Stack Functions
func TestStack() {
	ls := state.New(16, nil)

	ls.PushBoolean(true)
	printStack(ls)
	ls.PushInteger(10)
	printStack(ls)
	ls.PushNil()
	printStack(ls)
	ls.PushString("hello")
	printStack(ls)
	ls.PushValue(-4)
	printStack(ls)
	ls.Replace(3)
	printStack(ls)
	ls.SetTop(6)
	printStack(ls)
	ls.Remove(-3)
	printStack(ls)
	ls.SetTop(-5)
	printStack(ls)
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

/**
Read BinaryChunk Information
**/
/*
func list(f *binchunk.Prototype) {
	printHeader(f)
	printCode(f)
	printDetail(f)
	for _, p := range f.Protos {
		list(p)
	}
}

func printHeader(f *binchunk.Prototype) {
	funcType := "main"
	if f.LineDefined > 0 {
		funcType = "function"
	}

	varargFlag := ""
	if f.IsVararg > 0 {
		varargFlag = "+"
	}

	fmt.Printf("\n%s <%s:%d,%d> (%d instructions)\n",
		funcType, f.Source, f.LineDefined, f.LastLineDefined, len(f.Code))

	fmt.Printf("%d%s params, %d slots, %d upvalues, ",
		f.NumParams, varargFlag, f.MaxStackSize, len(f.Upvalues))

	fmt.Printf("%d locals, %d constants, %d functions\n",
		len(f.LocVars), len(f.Constants), len(f.Protos))
}

func printCode(f *binchunk.Prototype) {
	for pc, c := range f.Code {
		line := "-"
		if len(f.LineInfo) > 0 {
			line = fmt.Sprintf("%d", f.LineInfo[pc])
		}
		fmt.Printf("\t%d\t[%s]\t0x%08X\n", pc+1, line, c)
	}
}

func printDetail(f *binchunk.Prototype) {
	fmt.Printf("constants (%d):\n", len(f.Constants))
	for i, k := range f.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(k))
	}

	fmt.Printf("locals (%d):\n", len(f.LocVars))
	for i, locVar := range f.LocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n",
			i, locVar.VarName, locVar.StartPC+1, locVar.EndPC+1)
	}

	fmt.Printf("upvalues (%d):\n", len(f.Upvalues))
	for i, upval := range f.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n",
			i, upvalName(f, i), upval.Instack, upval.Idx)
	}
}

func constantToString(k interface{}) string {
	switch k.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case string:
		return fmt.Sprintf("%q", k)
	default:
		return "?"
	}
}

func upvalName(f *binchunk.Prototype, idx int) string {
	if len(f.UpvalueNames) > 0 {
		return f.UpvalueNames[idx]
	}
	return "-"
}
*/
