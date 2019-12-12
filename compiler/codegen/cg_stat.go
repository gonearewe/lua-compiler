package codegen

import (
	. "github.com/gonearewe/lua-compiler/compiler/ast"
	. "github.com/gonearewe/lua-compiler/compiler/lexer"
)

func cgStat(fi *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *BreakStat:
		cgBreakStat(fi, stat)
	case *DoStat:
		cgDoStat(fi, stat)
	case *WhileStat:
		cgWhileStat(fi, stat)
	case *RepeatStat:
		cgRepeatStat(fi, stat)
	case *IfStat:
		cgIfStat(fi, stat)
	case *ForNumStat:
		cgForNumStat(fi, stat)
	case *ForInStat:
		cgForInStat(fi, stat)
	case *AssignStat:
		cgAssignStat(fi, stat)
	case *LocalVarDeclStat:
		cgLocalVarDeclStat(fi, stat)
	case *LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *LabelStat, *GotoStat:
		panic("label and goto statements are not supported!")
	}
}

func cgLocalFuncDefStat(fi *funcInfo, node *LocalFuncDefStat) {
	r := fi.addLocVar(node.Name)
	cgFuncDefExp(fi, node.Exp, r)
}

func cgFuncCallStat(fi *funcInfo, node *FuncCallStat) {
	r := fi.allocRegs()
	cgFuncCallExp(fi, node, r, 0)
	fi.freeReg()
}

func cgBreakStat(fi *funcInfo, node *BreakStat) {
	pc := fi.emitJmp(0, 0)
	fi.addBreakJmp(pc)
}

func cgDoStat(fi *funcInfo, node *DoStat) {
	fi.enterScope(false)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.exitScope()
}

func (f *funcInfo) closeOpenUpvals() {
	a := f.getJmpArgA()
	if a > 0 {
		f.emitJmp(a, 0)
	}
}

func (f *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := f.maxRegs
	for _, locVar := range f.locNames {
		if locVar.scopeLv == f.scopeLv {
			for v := locVar; v != nil && v.scopeLv == f.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}

				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}

	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

func cgWhileStat(fi *funcInfo, node *WhileStat) {
	pcBeforeExp := fi.pc()
	r := fi.allocRegs()
	cgExp(fi, node.Exp, r, 1)
	fi.freeReg()
	fi.emitTest(r, 0)
	pcJmpToEnd := fi.emitJmp(0, 0)
	fi.enterScope(true)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.emitJmp(0, pcBeforeExp-fi.pc()-1)
	fi.exitScope()
	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

func cgRepeatStat(fi *funcInfo, node *RepeatStat) {
	fi.enterScope(true)

	pcBeforeBlock := fi.pc()
	cgBlock(fi, node.Block)
	r := fi.allocReg()
	cgExp(fi, node.Exp, r, 1) // it's also included in the scope, thus can be access from the block
	fi.freeReg()

	fi.emitTest(r, 0)
	fi.emitJmp(fi.getJmpArgA(), pcBeforeBlock-fi.pc()-1)
	fi.closeOpenUpvals()

	fi.exitScope()

}

func cgIfStat(fi *funcInfo, node *IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToEnds >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
			r := fi.allocReg()
			cgExp(fi, exp, r, 1)
			fi.freeReg()
			fi.emitTest(r, 0)
			pcJmpToNextExp = fi.emitJmp(0, 0)

			fi.enterScope(false)
			cgBlock(fi, node.Blocks[i])
			fi.closeOpenUpvals()
			fi.exitScope()

			if i < len(node.Exps)-1 {
				pcJmpToEnds[i] = fi.emitJmp(0, 0)
			} else {
				pcJmpToEnds[i] = pcJmpToNextExp
			}
		}

		for _, pc := range pcJmpToEnds {
			fi.fixSbx(pc, fi.pc()-pc)
		}
	}
}

func cgForNumStat(fi *funcInfo, node *ForNumStat) {
	fi.enterScope(true)
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{
		NameList: []string{"(for index)", "(for limit)", "(for step)"},
		ExpList:  []Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName)

	a := fi.usedRegs - 4
	pcForPre := fi.emitForPrep(a, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	pcForLoop := fi.emitForLoop(a, 0)

	fi.fixSbx(pcForLoop, pcForLoop-pcForPre-1)
	fi.fixSbx(pcForLoop, pcForPre-pcForLoop)

	fi.exitScope()
}

func cgForInStat(fi *funcInfo, node *ForInStat) {
	fi.enterScope(true)
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{
		NameList: []string{"(for generator)", "(for state)", "(for control)"},
		ExpList:  node.ExpList,
	})

	for _, name := range node.NameList {
		fi.addLocVar(name)
	}

	pcJmpToTFC := fi.emitJmp(0, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)

	rGenerator := fi.slotOfLocVar("(for generator)")
	fi.emitTForCall(rGenerator, len(node.NameList))
	fi.emitTForLoop(rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope()

}

func cgLocalVarDeclStat(fi *funcInfo, node *LocalVarDeclStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nNames := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nNames {
		for _, exp := range exps {
			a := fi.allocReg()
			cgExp(fi, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else {
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}

		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}

	fi.usedRegs = oldRegs
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
}

func cgAssignStat(fi *funcInfo, node *AssignStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)
	oldRegs := fi.usedRegs
	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			cgExp(fi, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			cgExp(fi, taExp.KeyExp, kRegs[i], 1)
		}
	}

	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else {
		multReg := false
		for i, exp := range exps {
			a := allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multReg = true
				n := nVars - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocReg(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}

		if !multReg {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}

	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(vRegs[i], b)
			} else { // global variable
				a := fi.indexOfUpval("_ENV")
				b := 0x100 + fi.indexOfConstant(varName)
				fi.emitSetTabUp(a, b, vRegs[i])
			}
		} else {
			fi.emitSetTable(tRegs[i], kRegs[i], vRegs[i])
		}
	}

	fi.usedRegs = oldRegs
}
