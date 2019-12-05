package codegen

type funcInfo struct {
	constants map[interface{}]int // key is the constant's value and val is it's index in the constant list
	usedRegs  int
	maxRegs   int

	scopeLv  int
	locVars  []*locVarInfo          // all declared local variables in order
	locNames map[string]*locVarInfo // current valid relationship between variable's name and the actual variable

	breaks [][]int // maintain addresses of `break` jmp

	parent   *funcInfo
	upvalues map[string]upvalInfo
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
