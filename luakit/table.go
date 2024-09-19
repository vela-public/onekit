package luakit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"reflect"
)

func IntOr(tab *lua.LTable, key string, d int) int {
	lv := tab.RawGetString(key)

	switch lv.Type() {
	case lua.LTNumber:
		return int(lv.(lua.LNumber))
	case lua.LTInt:
		return int(lv.(lua.LInt))
	default:
		return d
	}
}

func StringOr(tab *lua.LTable, key string, d string) string {
	lv := tab.RawGetString(key)
	if lv == lua.LNil {
		return d
	}

	return lv.String()
}

func Copier(L *lua.LState, field reflect.Value, val lua.LValue) error {
	typ := field.Type()

	kind := field.Type().Kind()
	switch kind {
	case reflect.String:
		field.SetString(val.String())
		return nil
	case reflect.Int:
		switch val.Type() {
		case lua.LTNumber:
			field.SetInt(int64(val.(lua.LNumber)))
		case lua.LTInt:
			field.SetInt(int64(val.(lua.LInt)))
		case lua.LTInt64:
			field.SetInt(int64(val.(lua.LInt64)))
		case lua.LTUint:
			field.SetInt(int64(val.(lua.LUint)))
		case lua.LTUint64:
			field.SetInt(int64(val.(lua.LUint64)))
		default:
			return fmt.Errorf("type mismatch for fied:%s must %s got:%s", typ.Name, field.Type().Name(), val.Type().String())
		}
		return nil

	case reflect.Bool:
		switch val.Type() {
		case lua.LTBool:
			field.SetBool(bool(val.(lua.LBool)))
		case lua.LTNumber:
			field.SetBool(int64(val.(lua.LNumber)) != 0)
		default:
			return fmt.Errorf("type mismatch for fied:%s must %s got:%s", typ.Name, field.Type().Name(), val.Type().String())
		}
		return nil

	case reflect.Struct:
		if val.Type() != lua.LTTable {
			return fmt.Errorf("type mismatch for fied:%s must %s got:%s", typ.Name, field.Type().Name(), val.Type().String())
		}
		return TableTo(L, val.(*lua.LTable), field.Interface())

	case reflect.Slice:
		switch val.Type() {
		case lua.LTTable:
			a := val.(*lua.LTable).Array()
			sa := reflect.MakeSlice(typ, len(a), len(a))
			for i, elem := range a {
				if err := Copier(L, sa.Index(i), elem); err != nil {
					return err
				}
			}

			field.Set(sa)
			return nil
		case lua.LTInt:
			field.Set(reflect.ValueOf([]int{lua.IsInt(val)}))
			return nil
		case lua.LTString:
			field.Set(reflect.ValueOf([]string{val.String()}))
			return nil
		case lua.LTBool:
			field.Set(reflect.ValueOf([]bool{bool(val.(lua.LBool))}))
			return nil

		default:
			return fmt.Errorf("type mismatch for fied:%s must %s got:%s", typ.Name, field.Type().Name(), val.Type().String())

		}
	default:
		return fmt.Errorf("type mismatch for fied:%s must %s got:%s", typ.Name, field.Type().Name(), val.Type().String())
	}
}

func TableTo(L *lua.LState, tab *lua.LTable, v any) error {
	vo := reflect.ValueOf(v)
	if vo.Kind() != reflect.Ptr || vo.IsNil() {
		return fmt.Errorf("must be non-nil pointer to %T", v)
	}

	dst := vo.Elem()
	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("must be struct pointer to %T", v)
	}

	vt := dst.Type()
	for i := 0; i < dst.NumField(); i++ {
		typ := vt.Field(i)
		tag := typ.Tag.Get("lua")
		if len(tag) == 0 {
			tag = typ.Name
		}

		if tag == "-" {
			continue
		}

		val := tab.RawGetString(tag)
		if val == lua.LNil {
			continue
		}

		field := dst.Field(i)
		if !field.CanSet() {
			continue
		}

		if e := Copier(L, field, val); e != nil {
			return e
		}

	}

	return nil
}
