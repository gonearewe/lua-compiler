package state

import (
	"fmt"

	. "github.com/gonearewe/lua-compiler/api"
)

func (self *luaState) RawLen(idx int) uint {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return uint(len(x))
	case *luaTable:
		return uint(x.len())
	default:
		return 0
	}
}

func (l *luaState) Type(idx int) LuaType {
	if l.stack.isValid(idx) {
		val := l.stack.get(idx)
		return typeOf(val)
	}

	return LUA_TNONE
}

func (l *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value"
	case LUA_TNIL:
		return "nil"
	case LUA_TBOOLEAN:
		return "boolean"
	case LUA_TNUMBER:
		return "number"
	case LUA_TSTRING:
		return "string"
	case LUA_TTABLE:
		return "table"
	case LUA_TFUNCTION:
		return "function"
	case LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}

/**************************
following methods check the type of index idx and return bool
**************************/

func (l *luaState) IsNone(idx int) bool {
	return l.Type(idx) == LUA_TNONE
}

func (l *luaState) IsNil(idx int) bool {
	return l.Type(idx) == LUA_TNIL
}

func (l *luaState) IsNoneOrNil(idx int) bool {
	t := l.Type(idx)
	return t == LUA_TNONE || t == LUA_TNIL
}

func (l *luaState) IsBoolean(idx int) bool {
	return l.Type(idx) == LUA_TBOOLEAN
}

// check if the type of index idx is string or number
func (l *luaState) IsString(idx int) bool {
	t := l.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

// check if the type of index idx is number or can be conversed to number
func (l *luaState) IsNumber(idx int) bool {
	_, ok := l.ToNumberX(idx)
	return ok
}

func (l *luaState) IsInteger(idx int) bool {
	val := l.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

/**************************
following methods return the conversed form of the type of index idx to another type
but only ToString() and ToStringX() actually modify the stack
**************************/

func (l *luaState) ToBoolean(idx int) bool {
	val := l.stack.get(idx)
	return convertToBoolean(val)
}

func (l *luaState) ToNumber(idx int) float64 {
	n, _ := l.ToNumberX(idx)
	return n
}

// ToNumberX will tell you if the conversion is successful by returning bool
func (l *luaState) ToNumberX(idx int) (float64, bool) {
	val := l.stack.get(idx)
	return convertToFloat(val)
}

func (l *luaState) ToInteger(idx int) int64 {
	n, _ := l.ToIntegerX(idx)
	return n
}

// ToIntegerX will tell you if the conversion is successful by returning bool
func (l *luaState) ToIntegerX(idx int) (int64, bool) {
	val := l.stack.get(idx)
	return convertToInteger(val)
}

func (l *luaState) ToString(idx int) string {
	s, _ := l.ToStringX(idx)
	return s
}

// ToStringX will tell you if the conversion is successful by returning bool
// notice this method will change the stack,
// if the value of index idx can be conversed to string,it will be set to its string form,
// if not, returns ("", false) and stack stays unchanged
func (l *luaState) ToStringX(idx int) (string, bool) {
	val := l.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		l.stack.set(idx, s) //
		return s, true
	default:
		return "", false
	}
}

func (l *luaState) IsGoFunction(idx int) bool {
	val := l.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc != nil
	}

	return false
}

// Converse the luaValue whose index in the stack is given by idx
// to GoFunction and return it, return nil when facing conversion failure;
// This method doesn't change the stack.
func (l *luaState) ToGoFunction(idx int) GoFunction {
	val := l.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}

	return nil
}
