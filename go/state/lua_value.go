package state

import (
	. "github.com/gonearewe/lua-compiler/go/api"
	"github.com/gonearewe/lua-compiler/go/number"
)

// import _ "../api"

type luaValue interface{}

func typeOf(val luaValue) LuaType {
	switch val.(type) {
	case nil:
		return LUA_TNIL
	case bool:
		return LUA_TBOOLEAN
	case int64:
		return LUA_TNUMBER
	case float64:
		return LUA_TNUMBER
	case string:
		return LUA_TSTRING
	case *luaTable:
		return LUA_TTABLE
	default:
		panic("TODO !")
	}
}

// in lua, when val is not boolean, if not nil, then it's true
func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}

func convertToFloat(val luaValue) (float64, bool) {
	switch x := val.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case string:
		return number.ParserFloat(x)
	default:
		return 0, false
	}
}

func convertToInteger(val luaValue) (int64, bool) {
	switch x := val.(type) {
	case int64:
		return x, true
	case float64:
		return number.FloatToInteger(x)
	case string:
		return _stringToInteger(x)
	default:
		return 0, false
	}
}

// if string can not be conversed to integer directly,
// it will be conversed to float before finally to integer
func _stringToInteger(s string) (int64, bool) {
	if i, ok := number.ParserInteger(s); ok {
		return i, true
	}
	if f, ok := number.ParserFloat(s); ok {
		return number.FloatToInteger(f)
	}

	return 0, false
}
