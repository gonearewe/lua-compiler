package vm

import "github.com/gonearewe/lua-compiler/go/api"

type Instruction uint32

const MAXARG_Bx = 1<<18 - 1       // 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 131071

// extracts the operation code from the instruction
func (i Instruction) Opcode() int {
	return int(i & 0x3f)
}

// available in iABC mode
// extracts the operands from the instruction
func (i Instruction) ABC() (a, b, c int) {
	a = int(i >> 6 & 0xFF)
	b = int(i >> 14 & 0x1FF)
	c = int(i >> 23 & 0x1FF)
	return
}

// available in iABx mode
// extracts the operands from the instruction
func (i Instruction) ABx() (a, bx int) {
	a = int(i >> 6 & 0xFF)
	bx = int(i >> 14)
	return
}

// available in iAsBx mode
// extracts the operands from the instruction
// sbx is signed int while others are all unsigned
func (i Instruction) AsBx() (a, sbx int) {
	a, bx := i.ABx()
	return a, bx - MAXARG_sBx
}

// available in iAx mode
// extracts the operands from the instruction
func (i Instruction) Ax() int {
	return int(i >> 6)
}

func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}

func (self Instruction) Execute(vm api.LuaVM) {
	action := opcodes[self.Opcode()].action
	if action != nil {
		action(self, vm)
	} else {
		panic(self.OpName())
	}
}
