package codegen

import (
	. "github.com/gonearewe/lua-compiler/compiler/ast"
	. "github.com/gonearewe/lua-compiler/compiler/lexer"
	. "github.com/gonearewe/lua-compiler/vm"
)

var arithAndBitwiseBinops = map[int]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

type funcInfo struct {
	insts []uint32 // corresponded instructions in binary chunk

	constants map[interface{}]int // key is the constant's value and val is it's index in the constant list
	usedRegs  int
	maxRegs   int

	scopeLv  int
	locVars  []*locVarInfo          // all declared local variables in order
	locNames map[string]*locVarInfo // current valid relationship between variable's name and the actual variable

	breaks [][]int // maintain addresses of `break` jmp

	parent   *funcInfo
	upvalues map[string]upvalInfo

	subFuncs  []*funcInfo
	numParams int
	isVararg  bool
}

// In lua, variable's name is just a label, a rather different thing from variable itself.
// Thus, we maintain a linked list for a local variable.
type locVarInfo struct {
	prev     *locVarInfo // to construct a linked list
	name     string      // name of the variable
	scopeLv  int         // level of scope
	slot     int         // correspond index in the registers
	captured bool        // whether it's captured by a closure
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
	}
}

func (f *funcInfo) newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
	}
}

// Return index of k in the constant list if found, if not,
// add it to the constant list and return its index.
func (f *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := f.constants[k]; found {
		return idx
	}

	idx := len(f.constants)
	f.constants[k] = idx

	return idx
}

func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}
	if self.usedRegs > self.maxRegs {
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1
}

func (self *funcInfo) freeReg() {
	if self.usedRegs <= 0 {
		panic("usedRegs <= 0 !")
	}
	self.usedRegs--
}

func (f *funcInfo) allocRegs() int {
	f.usedRegs++

	if f.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}

	if f.usedRegs > f.maxRegs {
		f.maxRegs = f.usedRegs
	}

	return f.usedRegs - 1
}

func (f *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		f.freeReg()
	}
}

func (f *funcInfo) enterScope(breakable bool) {
	f.scopeLv++

	if breakable { // a loop scope
		f.breaks = append(f.breaks, []int{})
	} else {
		f.breaks = append(f.breaks, nil)
	}
}

func (f *funcInfo) exitScope() {
	pendingBreakJmps := f.breaks[len(f.breaks)-1]
	f.breaks = f.breaks[:len(f.breaks-1)]
	a := f.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := f.pc() - pc
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP
		f.insts[pc] = uint32(i)
	}

	f.scopeLv--
	for _, locVar := range f.locNames {
		if locVar.scopeLv > f.scopeLv {
			f.removeLocVar(locVar)
		}
	}
}

func (f *funcInfo) addLocVar(name string) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    f.locNames[name],
		scopeLv: f.scopeLv,
		slot:    f.allocReg(),
	}

	f.locVars = append(f.locVars, newVar)
	f.locNames[name] = newVar

	return newVar.slot
}

// Remove a local variable by freeing variable's register and collecting the name if it's used outside this scope.
func (f *funcInfo) removeLocVar(locVar *locVarInfo) {
	f.freeReg()

	if locVar.prev == nil {
		delete(f.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		f.removeLocVar(locVar.prev)
	} else {
		f.locNames[locVar.name] = locVar.prev
	}
}

// Return name's bound index in the registers, return -1 if not found.
func (f *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := f.locNames[name]; found {
		return locVar.slot
	}

	return -1
}

func (f *funcInfo) addBreakJmp(pc int) {
	for i := f.scopeLv; i >= 0; i-- {
		if f.breaks[i] != nil {
			// add jmp address for `break`
			f.breaks[i] = append(f.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop !")
}

// Return the index of upval bound with given name,
// try to bind if not bound, return -1 if failed to bind.
func (f *funcInfo) indexOfUpval(name string) int {
	if upval, ok := f.upvalues[name]; ok {
		return upval.index
	}

	if f.parent != nil {
		if locVar, found := f.parent.locNames[name]; found {
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{locVar.slot}
			locVar.captured = true

			return idx
		}

		if uvIdx := f.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{-1, uvIdx, idx}

			return idx
		}
	}

	return -1
}

/* instructions generation */

func (f *funcInfo) pc() int {
	return len(f.insts) - 1
}

func (self *funcInfo) fixSbx(pc, sBx int) {
	i := self.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+MAXARG_sBx)<<14 // reset sBx
	self.insts[pc] = i
}

func (self *funcInfo) emitABC(opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
}

func (self *funcInfo) emitABx(opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
}

func (self *funcInfo) emitAsBx(opcode, a, b int) {
	i := (b+MAXARG_sBx)<<14 | a<<6 | opcode
	self.insts = append(self.insts, uint32(i))
}

func (self *funcInfo) emitAx(opcode, ax int) {
	i := ax<<6 | opcode
	self.insts = append(self.insts, uint32(i))
}

// r[a] = r[b]
func (self *funcInfo) emitMove(a, b int) {
	self.emitABC(OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (self *funcInfo) emitLoadNil(a, n int) {
	self.emitABC(OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (self *funcInfo) emitLoadBool(a, b, c int) {
	self.emitABC(OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (self *funcInfo) emitLoadK(a int, k interface{}) {
	idx := self.indexOfConstant(k)
	if idx < (1 << 18) {
		self.emitABx(OP_LOADK, a, idx)
	} else {
		self.emitABx(OP_LOADKX, a, 0)
		self.emitAx(OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (self *funcInfo) emitVararg(a, n int) {
	self.emitABC(OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (self *funcInfo) emitClosure(a, bx int) {
	self.emitABx(OP_CLOSURE, a, bx)
}

// r[a] = {}
func (self *funcInfo) emitNewTable(a, nArr, nRec int) {
	self.emitABC(OP_NEWTABLE,
		a, Int2fb(nArr), Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (self *funcInfo) emitSetList(a, b, c int) {
	self.emitABC(OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (self *funcInfo) emitGetTable(a, b, c int) {
	self.emitABC(OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTable(a, b, c int) {
	self.emitABC(OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (self *funcInfo) emitGetUpval(a, b int) {
	self.emitABC(OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (self *funcInfo) emitSetUpval(a, b int) {
	self.emitABC(OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (self *funcInfo) emitGetTabUp(a, b, c int) {
	self.emitABC(OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (self *funcInfo) emitSetTabUp(a, b, c int) {
	self.emitABC(OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (self *funcInfo) emitCall(a, nArgs, nRet int) {
	self.emitABC(OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (self *funcInfo) emitTailCall(a, nArgs int) {
	self.emitABC(OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (self *funcInfo) emitReturn(a, n int) {
	self.emitABC(OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (self *funcInfo) emitSelf(a, b, c int) {
	self.emitABC(OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (self *funcInfo) emitJmp(a, sBx int) int {
	self.emitAsBx(OP_JMP, a, sBx)
	return len(self.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (self *funcInfo) emitTest(a, c int) {
	self.emitABC(OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (self *funcInfo) emitTestSet(a, b, c int) {
	self.emitABC(OP_TESTSET, a, b, c)
}

func (self *funcInfo) emitForPrep(a, sBx int) int {
	self.emitAsBx(OP_FORPREP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitForLoop(a, sBx int) int {
	self.emitAsBx(OP_FORLOOP, a, sBx)
	return len(self.insts) - 1
}

func (self *funcInfo) emitTForCall(a, c int) {
	self.emitABC(OP_TFORCALL, a, 0, c)
}

func (self *funcInfo) emitTForLoop(a, sBx int) {
	self.emitAsBx(OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (self *funcInfo) emitUnaryOp(op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		self.emitABC(OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		self.emitABC(OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		self.emitABC(OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		self.emitABC(OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (self *funcInfo) emitBinaryOp(op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		self.emitABC(opcode, a, b, c)
	} else {
		switch op {
		case TOKEN_OP_EQ:
			self.emitABC(OP_EQ, 1, b, c)
		case TOKEN_OP_NE:
			self.emitABC(OP_EQ, 0, b, c)
		case TOKEN_OP_LT:
			self.emitABC(OP_LT, 1, b, c)
		case TOKEN_OP_GT:
			self.emitABC(OP_LT, 1, c, b)
		case TOKEN_OP_LE:
			self.emitABC(OP_LE, 1, b, c)
		case TOKEN_OP_GE:
			self.emitABC(OP_LE, 1, c, b)
		}
		self.emitJmp(0, 1)
		self.emitLoadBool(a, 0, 1)
		self.emitLoadBool(a, 1, 0)
	}
}
