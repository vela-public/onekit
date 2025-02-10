package libkit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"reflect"
	"testing"
	"unsafe"
)

type MyStruct struct {
	Field1 int
	Field2 string
}

func TestA(t *testing.T) {

	s := MyStruct{
		Field1: 42,
		Field2: "Hello",
	}

	ptr := unsafe.Pointer(&s)

	// 获取结构体字段的内存地址（uintptr）
	field1Addr := (*int)(unsafe.Pointer(uintptr(ptr) + unsafe.Offsetof(s.Field1)))
	field2Addr := (*string)(unsafe.Pointer(uintptr(ptr) + unsafe.Offsetof(s.Field2)))
	fmt.Printf("Field1 address: %v\n", *field1Addr)
	fmt.Printf("Field2 address: %v\n", *field2Addr)
}

func TestB(t *testing.T) {
	// 定义结构体的类型
	field1 := reflect.StructField{
		Name: "Name",
		Type: reflect.TypeOf(""), // 字符串类型
		Tag:  reflect.StructTag(`json:"name"`),
	}

	field2 := reflect.StructField{
		Name: "Age",
		Type: reflect.TypeOf(0), // 整数类型
		Tag:  reflect.StructTag(`json:"age"`),
	}

	field3 := reflect.StructField{
		Name: "Call",
		Type: reflect.TypeOf(func(*lua.LState) int { return 0 }), // 整数类型
		Tag:  reflect.StructTag(`json:"-"`),
	}

	// 动态构造结构体类型
	structType := reflect.StructOf([]reflect.StructField{field1, field2, field3})

	// 创建一个新的结构体实例
	structValue := reflect.New(structType).Elem()

	// 设置字段值
	structValue.FieldByName("Name").SetString("John Doe")
	structValue.FieldByName("Age").SetInt(30)
	structValue.FieldByName("Call").Set(reflect.ValueOf(func(L *lua.LState) int {
		print(L.CheckString(1))
		L.Push(lua.S2L("app"))
		return 1
	}))

	// 打印结构体的值
	fmt.Println(structValue)
}
