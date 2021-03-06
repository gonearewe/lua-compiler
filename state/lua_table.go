package state

import (
	"math"

	"github.com/gonearewe/lua-compiler/number"
)

// In lua, if the keys of the Table are continus int, the table is
// stored in memory as Array
type luaTable struct {
	metatable *luaTable
	arr       []luaValue
	_map      map[luaValue]luaValue
	// support iterator, this map is actually a linked list.
	// for 5-> 3-> 6-> 1, the map stores [nil]5, [5]3, [3]6, [6]1, [1]index4.
	keys    map[luaValue]luaValue
	changed bool
}

func newLuaTable(nArr, nRec int) *luaTable {
	t := &luaTable{}
	if nArr > 0 {
		t.arr = make([]luaValue, 0, nArr)
	}
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return t
}

func (l *luaTable) get(key luaValue) luaValue {
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		if idx >= 1 && idx <= int64(len(l.arr)) {
			return l.arr[idx-1]
		}
	}
	return l._map[key]
}

func _floatToInteger(key luaValue) luaValue {
	if f, ok := key.(float64); ok {
		if i, ok := number.FloatToInteger(f); ok {
			return i
		}
	}

	return key
}

func (l *luaTable) put(key, val luaValue) {
	if key == nil {
		panic("table index is nil !")
	}
	if f, ok := key.(float64); ok && math.IsNaN(f) {
		panic("table index is NaN !")
	}

	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok && idx >= 1 {
		arrlen := int64(len(l.arr))

		if idx <= arrlen {
			l.arr[idx-1] = val

			if idx == arrlen && val == nil {
				// A nil val means a hole, which leads to shrinking
				l._shrinkArray()
			}

			return
		}

		if idx == arrlen+1 {
			delete(l._map, key)

			if val != nil {
				l.arr = append(l.arr, val)
				l._expandArray()
			}

			return
		}

	}

	if val != nil {
		if l._map == nil {
			// initialize the map with cap of 8
			l._map = make(map[luaValue]luaValue, 8)
		}

		l._map[key] = val
	} else {
		delete(l._map, key)
	}

}

func (l *luaTable) _shrinkArray() {
	for i := len(l.arr) - 1; i >= 0; i-- {
		if l.arr[i] == nil {
			l.arr = l.arr[0:i]
		}
	}
}

// Expand the array in luaTable, may move value originally in map into array.
func (l *luaTable) _expandArray() {
	for idx := int64(len(l.arr)) + 1; true; idx++ {
		if val, found := l._map[idx]; found {
			delete(l._map, val)
			l.arr = append(l.arr, val)
		} else {
			break
		}
	}
}

func (l *luaTable) len() int {
	return len(l.arr)
}

func (l *luaTable) hasMetafield(fieldName string) bool {
	return l.metatable != nil && l.metatable.get(fieldName) != nil
}

// Method for iterator, receive one key and return next key.
func (l *luaTable) nextKey(key luaValue) luaValue {
	if l.keys == nil || key == nil {
		l.initKeys()
		l.changed = false
	}

	return l.keys[key]
}

// Method for iterator, range arr and _map to organise keys.
func (l *luaTable) initKeys() {
	l.keys = make(map[luaValue]luaValue)
	var key luaValue = nil

	for i, v := range l.arr {
		if v != nil {
			l.keys[key] = int64(i + 1)
			key = int64(i + 1)
		}
	}

	for k, v := range l._map {
		if v != nil {
			l.keys[key] = k
			key = k
		}
	}
}
