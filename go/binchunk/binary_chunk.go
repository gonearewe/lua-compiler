package binchunk

// some constants for header struct
const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

// some constants for Constants in Prototype struct
const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header
	sizeUpvalues byte
	mainFunc     *Prototype
}

type header struct {
	signature       [4]byte // magic number 0x1B4C7561, for identifying chunk file
	version         byte    // for version x.y.z, it is x*16+y
	format          byte    // default value is 0
	luacData        [6]byte // for verifying file, "\x19\x93\r\n\x1a\n"
	cintSize        byte
	sizetSize       byte
	instructionSize byte
	luaIntegerSize  byte
	luaNumberSize   byte
	luacInt         int64   // stores 0x5678, used to determine Big-endian or Little-endian
	luacNum         float32 // stores 370.5, used to determine number format(usually IEEE 754)
}

type Prototype struct {
	Source          string   // where source comes form, only not empty in main function
	LineDefined     uint32   // where this prototype starts in source file(line index), main function always starts with 0
	LastLineDefined uint32   // where this prototype ends in source file(line index)
	NumParams       byte     // useless if IsVararg is 1(true)
	IsVararg        byte     // whether this prototype accepts parametres of variable numbers, 0 for false and 1 for true
	MaxStackSize    byte     // how many virtual registers this function needs at least, stack is used to virtualize register
	Code            []uint32 // list of instructions
	Constants       []interface{}
	Upvalues        []Upvalue
	Protos          []*Prototype // list of sub-functions
	LineInfo        []uint32     // list of line indexes, mapped to Code(the instruction list)
	LocVars         []LocVar     // list of local variables
	UpvalueNames    []string     // mapped to Upvalues
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32 // index of the starting instruction
	EndPC   uint32 // index of ending instruction
}

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 &&
		string(data[:4]) == LUA_SIGNATURE
}

func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()
	reader.readByte() // size_upvalues
	return reader.readProto("")
}
