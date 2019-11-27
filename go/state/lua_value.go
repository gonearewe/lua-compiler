package state

import (
	"fmt"

	. "github.com/gonearewe/lua-compiler/go/api"
	"github.com/gonearewe/lua-compiler/go/number"
)

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
	case *closure:
		return LUA_TFUNCTION
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

// Set metatable for given luaValue, every luaTable contains a metatable
// and for other luaValue, each type shares one metatable in the registry.
func setMetatable(val luaValue, mt *luaTable, ls *luaState) {
	if t, ok := val.(*luaTable); ok {
		t.metatable = mt
		return
	}

	key := fmt.Sprintf("_MT%d", typeOf(val))
	ls.registry.put(key, mt)
}

// Get metatable for given luaValue, every luaTable contains a metatable
// and for other luaValue, each type shares one metatable in the registry.
func getMetatable(val luaValue, ls *luaState) *luaTable {
	if t, ok := val.(*luaTable); ok {
		return t.metatable
	}

	key := fmt.Sprintf("_MT%d", typeOf(val))
	if mt := ls.registry.get(key); mt != nil {
		return mt.(*luaTable)
	}

	return nil
}

func callMetamethod(a, b luaValue, mmName string, ls *luaState) (luaValue, bool) {
	var mm luaValue
	if mm = getMetafield(a, mmName, ls); mm == nil {
		if mm = getMetafield(b, mmName, ls); mm == nil {
			return nil, false
		}
	}

	ls.stack.check(4)
	ls.stack.push(mm)
	ls.stack.push(a)
	ls.stack.push(b)
	ls.Call(2, 1)

	return ls.stack.pop(), true
}

func getMetafield(val luaValue, fieldName string, ls *luaState) luaValue {
	if mt := getMetatable(val, ls); mt != nil {
		return mt.get(fieldName)
	}

	return nil
}
