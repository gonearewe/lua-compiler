package codegen

import (
	. "github.com/gonearewe/lua-compiler/compiler/ast"
	. "github.com/gonearewe/lua-compiler/compiler/lexer"
)

// kind of operands
const (
	ARG_CONST = 1 // const index
	ARG_REG   = 2 // register index
	ARG_UPVAL = 4 // upvalue index
	ARG_RK    = ARG_REG | ARG_CONST
	ARG_RU    = ARG_REG | ARG_UPVAL
	ARG_RUK   = ARG_REG | ARG_UPVAL | ARG_CONST
)

// todo: rename to evalExp()?
func cgExp(fi *funcInfo, node Exp, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(exp.Line, a, n)
	case *FalseExp:
		fi.emitLoadBool(exp.Line, a, 0, 0)
	case *TrueExp:
		fi.emitLoadBool(exp.Line, a, 1, 0)
	case *IntegerExp:
		fi.emitLoadK(exp.Line, a, exp.Val)
	case *FloatExp:
		fi.emitLoadK(exp.Line, a, exp.Val)
	case *StringExp:
		fi.emitLoadK(exp.Line, a, exp.Str)
	case *ParensExp:
		cgExp(fi, exp.Exp, a, 1)
	case *VarargExp:
		cgVarargExp(fi, exp, a, n)
	case *FuncDefExp:
		cgFuncDefExp(fi, exp, a)
	case *TableConstructorExp:
		cgTableConstructorExp(fi, exp, a)
	case *UnopExp:
		cgUnopExp(fi, exp, a)
	case *BinopExp:
		cgBinopExp(fi, exp, a)
	case *ConcatExp:
		cgConcatExp(fi, exp, a)
	case *NameExp:
		cgNameExp(fi, exp, a)
	case *TableAccessExp:
		cgTableAccessExp(fi, exp, a)
	case *FuncCallExp:
		cgFuncCallExp(fi, exp, a, n)
	}
}

func cgVaragExp(fi *funcInfo, node *VarargExp, a, n int) {
	if !fi.isVararg {
		panic("concat uses '...' outside a vararg function")
	}

	fi.emitVararg(a, n)
}
