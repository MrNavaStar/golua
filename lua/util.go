package lua

import (
	"fmt"
	"reflect"
)

// Registers a Go function as a global variable
func (L *State) Register(name string, f LuaGoFunction) {
	L.PushGoFunction(f)
	L.SetGlobal(name)
}

// Registers a map of go functions as a library that can be accessed using "require("name")"
func (L *State) RegisterLib(name string, funcs map[string]LuaGoFunction) {
	L.GetGlobal(name)
	found := L.IsTable(-1)
	if !found {
		L.Pop(1)
		L.CreateTable(0, len(funcs))
	}

	for fname, f := range funcs {
		L.PushGoFunction(f)
		L.SetField(-2, fname)
	}

	if !found {
		L.GetGlobal("package")
		L.GetField(-1, "loaded")
		L.PushValue(-3)
		L.SetField(-2, name)
		L.Pop(2)
	}
} 

func toSlice(arg interface{}) (out []interface{}, ok bool) {
	slice := reflect.ValueOf(arg)
    if slice.Kind() != reflect.Slice {
		return 
	}
    c := slice.Len()
    out = make([]interface{}, c)
    for i := 0; i < c; i++ {
        out[i] = slice.Index(i).Interface()
    }
    return out, true
}

// Pushes a representation of the given interface to the lua stack
func (L *State) PushGoInterface(value interface{}) {
	switch converted := value.(type) {
	case bool:
		L.PushBoolean(converted)
		return
	case int:
		L.PushInteger(int64(converted))
		return
	case float64:
		L.PushNumber(converted)
		return
	case string:
		L.PushString(converted)
		return
	case []byte:
		L.PushBytes(converted)
		return
	}

	slice, ok := toSlice(value)
	if ok {
		value = slice
	}

	switch converted := value.(type) {
 	case []interface{}:
		L.CreateTable(len(converted), 0)
		for i, item := range converted {
			L.PushGoInterface(item)
			L.RawSeti(-2, i+1)
		}
	case map[string]interface{}:
		L.CreateTable(0, len(converted))
		for key, item := range converted {
			L.PushGoInterface(item)
			L.SetField(-2, key)
		}
	}
}

func (L *State) PrintStack() {
	t := L.GetTop()
	fmt.Printf("~ | TOP\n")
	for i := t; i >= 1; i-- {
		if L.IsBoolean(i) {
			fmt.Printf("%d | BOOL : %t\n", i, L.ToBoolean(i))
			continue
		}
		if L.IsNumber(i) {
			fmt.Printf("%d | NUM  : %f\n", i, L.ToNumber(i))
			continue
		}
		if L.IsString(i) {
			fmt.Printf("%d | STR  : %s\n", i, L.ToString(i))
			continue
		}
		if L.IsTable(i) {
			fmt.Printf("%d | TBL  : Size:%d\n", i, L.ObjLen(i))
			continue
		}
		if L.IsFunction(i) || L.IsGoFunction(i) {
			fmt.Printf("%d | FUNC\n", i)
			continue
		}
		if L.IsUserdata(i) {
			fmt.Printf("%d | USR\n", i)
			continue
		}
		if L.IsLightUserdata(i) {
			fmt.Printf("%d | LUSR\n", i)
			continue
		}
		if L.IsNil(i) {
			fmt.Printf("%d | NIL\n", i)
		}
	}
	fmt.Printf("~ | BOTTOM\n")
}
