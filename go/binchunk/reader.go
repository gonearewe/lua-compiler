package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte
}

func (r *reader) readByte() byte {
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

/********************************
following methods read raw data from the data stream wrapped in reader
offset will increase automatically
********************************/

// returns n bytes from the data stream wrapped in reader
func (r *reader) readBytes(n uint) []byte {
	bytes := r.data[:n]
	r.data = r.data[n:]
	return bytes
}

func (r *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(r.data)
	r.data = r.data[4:]
	return i
}

func (r *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(r.data)
	r.data = r.data[8:]
	return i
}

func (r *reader) readLuaInteger() int64 {
	return int64(r.readUint64())
}

func (r *reader) readLuaNumber() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *reader) readString() string {
	size := uint(r.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(r.readUint64())
	}
	bytes := r.readBytes(size - 1)
	return string(bytes)
}

/********************************
following methods serve for readProto(), extracting useful information
offset will increase automatically
********************************/

// return the list of instructions
func (r *reader) readCode() []uint32 {
	code := make([]uint32, r.readUint64())
	for i := range code {
		code[i] = r.readUint32()
	}
	return code
}

func (r *reader) readConstant() interface{} {
	switch r.readByte() {
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return r.readByte() != 0
	case TAG_INTEGER:
		return r.readLuaInteger()
	case TAG_NUMBER:
		return r.readLuaNumber()
	case TAG_SHORT_STR:
		return r.readString()
	case TAG_LONG_STR:
		return r.readString()
	default:
		panic("corrupted !")
	}
}

// return the list of constants
func (r *reader) readConstants() []interface{} {
	constants := make([]interface{}, r.readUint32())
	for i := range constants {
		constants[i] = r.readConstant()
	}
	return constants
}

func (r *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, r.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: r.readByte(),
			Idx:     r.readByte(),
		}
	}

	return upvalues
}

// param parentSource indicates where the source of the chunk file comes from
// chunk file only save this information for main function,
// thus it requires to be passed to every subfunctions
// read sub-functions recursively
func (r *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, r.readUint32())
	for i := range protos {
		protos[i] = r.readProto(parentSource)
	}

	return protos
}

// returns the list of line indexes
func (r *reader) readLineInfo() []uint32 {
	lineinfos := make([]uint32, r.readUint32())
	for i := range lineinfos {
		lineinfos[i] = r.readUint32()
	}

	return lineinfos
}

func (r *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, r.readUint32())
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: r.readString(),
			StartPC: r.readUint32(),
			EndPC:   r.readUint32(),
		}
	}

	return locVars
}

func (r *reader) readUpvalueNames() []string {
	names := make([]string, r.readUint32())
	for i := range names {
		names[i] = r.readString()
	}

	return names
}

/********************************
following methods are ready for use
they offer api for reader
********************************/

// check the header of the chunk file, panic if anything goes wrong
func (r *reader) checkHeader() {
	if string(r.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk !")
	} else if r.readByte() != LUAC_VERSION {
		panic("version mismatch !")
	} else if r.readByte() != LUAC_FORMAT {
		panic("format mismatch !")
	} else if string(r.readBytes(6)) != LUAC_DATA {
		panic("corrupted !")
	} else if r.readByte() != CINT_SIZE {
		panic("int size mismatch !")
	} else if r.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch !")
	} else if r.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch !")
	} else if r.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch !")
	} else if r.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch !")
	} else if r.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch !")
	} else if r.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch !")
	}
}

func (r *reader) readProto(parentSource string) *Prototype {
	src := r.readString()
	if src == "" {
		src = parentSource
	}
	return &Prototype{
		Source:          src,
		LineDefined:     r.readUint32(),
		LastLineDefined: r.readUint32(),
		NumParams:       r.readByte(),
		IsVararg:        r.readByte(),
		MaxStackSize:    r.readByte(),
		Code:            r.readCode(),
		Constants:       r.readConstants(),
	}
}
