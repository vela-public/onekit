package abi

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type Kind int

const (
	INT32 Kind = iota
	INT64
	FLOAT32
	FLOAT64
	BOOL
	TEXT
	STRUCT
)

type Attribute struct {
	Name   string
	Kind   Kind
	Size   int
	Align  int
	Offset int
	Memory struct {
		Cap int
		Len int
	}
	Nested *StructBuilder // 嵌套结构体定义
}

type StructBuilder struct {
	mu         sync.Mutex
	attributes []Attribute
	packed     bool
	finalized  bool
	size       int
	align      int
}

func (b *StructBuilder) String() string                 { return "" }
func (b *StructBuilder) Type() lua.LValueType           { return lua.LTObject }
func (b *StructBuilder) AssertFloat64() (float64, bool) { return 0, false }
func (b *StructBuilder) AssertString() (string, bool)   { return "", false }
func (b *StructBuilder) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(b.DefineL), true
}

func (b *StructBuilder) Hijack(fsm *lua.CallFrameFSM) bool { return false }

func (b *StructBuilder) DefineL(L *lua.LState) int {
	ds := lua.Unpack[string](L)
	for _, d := range ds {
		if err := b.Define(d); err != nil {
			L.RaiseError("%v", err)
			return 0
		}
	}
	return 0
}

func (b *StructBuilder) finalizeL(L *lua.LState) int {
	s, err := b.Finalize()
	if err != nil {
		L.RaiseError("%v", err)
		return 0
	}

	L.Push(s)
	return 1
}

func (b *StructBuilder) fillL(L *lua.LState) int {
	text := L.CheckString(1)

	s, err := b.FromTexT(lua.S2B(text))
	if err != nil {
		L.RaiseError("%v", err)
		return 0
	}
	L.Push(s)
	return 1
}

func (b *StructBuilder) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "size":
		return lua.LNumber(b.size)
	case "align":
		return lua.LNumber(b.align)
	case "packed":
		return lua.LBool(b.packed)
	case "final":
		return lua.NewFunction(b.finalizeL)
	case "fill":
		return lua.NewFunction(b.fillL)
	}
	return lua.LNil
}

func (b *StructBuilder) Define(descriptor string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.finalized {
		return errors.New("builder already finalized")
	}

	parts := strings.SplitN(descriptor, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid descriptor: %s", descriptor)
	}
	name, kind := parts[0], parts[1]
	if name == "" || kind == "" {
		return errors.New("name and type required")
	}

	for _, attr := range b.attributes {
		if attr.Name == name {
			return fmt.Errorf("field %q already defined", name)
		}
	}

	if strings.HasPrefix(kind, "text(") {
		sizeStr := strings.TrimSuffix(strings.TrimPrefix(kind, "text("), ")")
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("invalid text size: %v", err)
		}
		b.attributes = append(b.attributes, Attribute{
			Name:  name,
			Kind:  TEXT,
			Size:  size,
			Align: 1,
			Memory: struct {
				Cap int
				Len int
			}{Cap: size},
		})
		return nil
	}

	if strings.HasPrefix(kind, "struct{") {
		sub := NewStructBuilder(b.packed)
		body := strings.TrimSuffix(strings.TrimPrefix(kind, "struct{"), "}")
		fields := strings.Split(body, ",")
		for _, f := range fields {
			if err := sub.Define(strings.TrimSpace(f)); err != nil {
				return err
			}
		}
		if _, err := sub.Finalize(); err != nil {
			return err
		}
		b.attributes = append(b.attributes, Attribute{
			Name:   name,
			Kind:   STRUCT,
			Size:   sub.size,
			Align:  sub.align,
			Nested: sub,
		})
		return nil
	}

	size, align := 0, 0
	switch kind {
	case "int32":
		size, align = 4, 4
	case "float32":
		size, align = 4, 4
	case "float64":
		size, align = 8, 8
	case "int64":
		size, align = 8, 8
	case "bool":
		size, align = 1, 1
	default:
		return fmt.Errorf("unsupported type: %s", kind)
	}

	b.attributes = append(b.attributes, Attribute{
		Name:  name,
		Kind:  toKind(kind),
		Size:  size,
		Align: align,
	})
	return nil
}

func NewStructBuilder(packed bool) *StructBuilder {
	return &StructBuilder{packed: packed}
}

type StructInstance struct {
	layout *StructBuilder
	buffer []byte
}

func (s *StructInstance) String() string {
	return lua.B2S(s.buffer)
}

func (s *StructInstance) Type() lua.LValueType                   { return lua.LTObject }
func (s *StructInstance) AssertFloat64() (float64, bool)         { return 0, false }
func (s *StructInstance) AssertString() (string, bool)           { return s.String(), true }
func (s *StructInstance) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (s *StructInstance) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (s *StructInstance) Index(L *lua.LState, key string) lua.LValue {
	for _, attr := range s.layout.attributes {
		if attr.Name != key {
			continue
		}

		switch attr.Kind {
		case INT32:
			return lua.LInt(int32(binary.LittleEndian.Uint32(s.buffer[attr.Offset:])))
		case INT64:
			return lua.LInt(int64(binary.LittleEndian.Uint64(s.buffer[attr.Offset:])))
		case BOOL:
			return lua.LBool(s.buffer[attr.Offset] != 0)
		case TEXT:
			return lua.LString(BytesToCleanString(s.buffer[attr.Offset : attr.Offset+attr.Memory.Len]))
		case STRUCT:
			sub := &StructInstance{layout: attr.Nested, buffer: s.buffer[attr.Offset : attr.Offset+attr.Size]}
			return sub
		default:
			return lua.LNil
		}
	}

	return lua.LNil
}

func (s *StructInstance) NewIndex(L *lua.LState, key string, val lua.LValue) {
	for _, attr := range s.layout.attributes {
		if attr.Name != key {
			continue
		}

		switch attr.Kind {
		case INT32:
			i32, ok := lua.Must[int32](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			err := s.SetInt32(key, i32)
			if err != nil {
				L.RaiseError("set field %q failed: %v", key, err)
				return
			}
		case INT64:
			i64, ok := lua.Must[int64](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			err := s.SetInt64(key, i64)
			if err != nil {
				L.RaiseError("set field %q failed: %v", key, err)
				return
			}
		case FLOAT32:
			f32, ok := lua.Must[float32](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			bits := math.Float32bits(f32)
			binary.LittleEndian.PutUint32(s.buffer[attr.Offset:], bits)
			return
		case FLOAT64:
			f64, ok := lua.Must[float64](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			bits := math.Float64bits(f64)
			binary.LittleEndian.PutUint64(s.buffer[attr.Offset:], bits)
			return

		case BOOL:
			b, ok := lua.Must[bool](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			err := s.SetBool(key, b)
			if err != nil {
				L.RaiseError("set field %q failed: %v", key, err)
				return
			}

		case TEXT:
			text, ok := lua.Must[string](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			err := s.SetText(key, text)
			if err != nil {
				L.RaiseError("set field %q failed: %v", key, err)
				return
			}

		case STRUCT:
			sub, ok := lua.Must[*StructInstance](val)
			if !ok {
				L.RaiseError("invalid value for field %q", key)
				return
			}
			err := s.SetStruct(key, sub)
			if err != nil {
				L.RaiseError("set field %q failed: %v", key, err)
				return
			}
		default:
		}
	}
}

func toKind(t string) Kind {
	switch t {
	case "int32":
		return INT32
	case "int64":
		return INT64
	case "bool":
		return BOOL
	case "text":
		return TEXT
	case "struct":
		return STRUCT
	case "float32":
		return FLOAT32
	case "float64":
		return FLOAT64
	default:
		return -1
	}
}

func (b *StructBuilder) FromTexT(raw []byte) (*StructInstance, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.finalized {
		if len(raw) != b.size {
			return nil, errors.New("mismatch builder size and raw size")
		}
		return &StructInstance{layout: b, buffer: raw}, nil
	}

	offset := 0
	maxAlign := 1
	for i := range b.attributes {
		a := &b.attributes[i]
		if !b.packed {
			offset = alignUp(offset, a.Align)
		}
		a.Offset = offset
		offset += a.Size
		if a.Align > maxAlign {
			maxAlign = a.Align
		}
	}
	if !b.packed {
		offset = alignUp(offset, maxAlign)
	}
	b.size = offset
	b.align = maxAlign
	b.finalized = true
	return &StructInstance{layout: b, buffer: make([]byte, offset)}, nil
}

func (b *StructBuilder) Finalize() (*StructInstance, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.finalized {
		return &StructInstance{layout: b, buffer: make([]byte, b.size)}, nil
	}

	offset := 0
	maxAlign := 1
	for i := range b.attributes {
		a := &b.attributes[i]
		if !b.packed {
			offset = alignUp(offset, a.Align)
		}
		a.Offset = offset
		offset += a.Size
		if a.Align > maxAlign {
			maxAlign = a.Align
		}
	}
	if !b.packed {
		offset = alignUp(offset, maxAlign)
	}
	b.size = offset
	b.align = maxAlign
	b.finalized = true
	return &StructInstance{layout: b, buffer: make([]byte, offset)}, nil
}

func (s *StructInstance) SetInt32(name string, val int32) error {
	attr, err := s.field(name, INT32)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(s.buffer[attr.Offset:], uint32(val))
	return nil
}

func (s *StructInstance) SetInt64(name string, val int64) error {
	attr, err := s.field(name, INT64)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(s.buffer[attr.Offset:], uint64(val))
	return nil
}

func (s *StructInstance) SetFloat32(name string, float322 float32) error {
	attr, err := s.field(name, FLOAT32)
	if err != nil {
		return err
	}
	bits := math.Float32bits(float322)
	binary.LittleEndian.PutUint32(s.buffer[attr.Offset:], bits)
	return nil
}

func (s *StructInstance) SetFloat64(name string, float642 float64) error {
	attr, err := s.field(name, FLOAT64)
	if err != nil {
		return err
	}
	bits := math.Float64bits(float642)
	binary.LittleEndian.PutUint64(s.buffer[attr.Offset:], bits)
	return nil
}

func (s *StructInstance) SetBool(name string, val bool) error {
	attr, err := s.field(name, BOOL)
	if err != nil {
		return err
	}
	if val {
		s.buffer[attr.Offset] = 1
	} else {
		s.buffer[attr.Offset] = 0
	}
	return nil
}

func (s *StructInstance) SetText(name, val string) error {
	attr, err := s.field(name, TEXT)
	if err != nil {
		return err
	}

	data := cast.S2B(val)
	sz := len(data)
	if sz >= attr.Size {
		sz = attr.Size - 1 // 预留 null terminator
		data = data[:sz]
	}

	// 拷贝字符串内容
	copy(s.buffer[attr.Offset:], data)

	// 写入 null terminator
	s.buffer[attr.Offset+sz] = 0

	// 清理多余区域（可选）
	for i := attr.Offset + sz + 1; i < attr.Offset+attr.Size; i++ {
		s.buffer[i] = 0
	}

	// 包括 '\0' 的长度
	attr.Memory.Len = sz + 1
	return nil
}

func (s *StructInstance) SetStruct(name string, sub *StructInstance) error {
	attr, err := s.field(name, STRUCT)
	if err != nil {
		return err
	}
	if attr.Nested != sub.layout {
		return fmt.Errorf("struct layout mismatch for field %q", name)
	}
	if attr.Size != len(sub.buffer) {
		return fmt.Errorf("struct size mismatch for field %q", name)
	}
	copy(s.buffer[attr.Offset:], sub.buffer)
	return nil
}

func (s *StructInstance) GetBool(name string) (bool, error) {
	attr, err := s.field(name, BOOL)
	if err != nil {
		return false, err
	}
	return s.buffer[attr.Offset] != 0, nil
}

func (s *StructInstance) GetInt32(name string) (int32, error) {
	attr, err := s.field(name, INT32)
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(s.buffer[attr.Offset:])), nil
}

func (s *StructInstance) GetText(name string) (string, error) {
	attr, err := s.field(name, TEXT)
	if err != nil {
		return "", err
	}
	raw := s.buffer[attr.Offset : attr.Offset+attr.Memory.Len]
	return cast.B2S(raw), nil
}

func (s *StructInstance) GetStruct(name string) (*StructInstance, error) {
	attr, err := s.field(name, STRUCT)
	if err != nil {
		return nil, err
	}
	return &StructInstance{layout: attr.Nested, buffer: s.buffer[attr.Offset : attr.Offset+attr.Size]}, nil
}

func (s *StructInstance) GetBytes(name string) ([]byte, error) {
	for _, attr := range s.layout.attributes {
		if attr.Name == name {
			return s.buffer[attr.Offset : attr.Offset+attr.Size], nil
		}
	}

	return nil, fmt.Errorf("field %q not found", name)
}

func (s *StructInstance) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	for _, attr := range s.layout.attributes {
		switch attr.Kind {
		case INT32:
			m[attr.Name] = int32(binary.LittleEndian.Uint32(s.buffer[attr.Offset:]))
		case INT64:
			m[attr.Name] = int64(binary.LittleEndian.Uint64(s.buffer[attr.Offset:]))
		case BOOL:
			m[attr.Name] = s.buffer[attr.Offset] != 0
		case TEXT:
			raw := s.buffer[attr.Offset : attr.Offset+attr.Size]
			m[attr.Name] = BytesToCleanString(raw)
		case STRUCT:
			sub := &StructInstance{layout: attr.Nested, buffer: s.buffer[attr.Offset : attr.Offset+attr.Size]}
			m[attr.Name] = sub.ToMap()
		default:
			//todo
		}
	}
	return m
}

func (s *StructInstance) ToJSON() ([]byte, error) {
	return json.Marshal(s.ToMap())
}

func (s *StructInstance) FromJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for _, attr := range s.layout.attributes {
		val, ok := m[attr.Name]
		if !ok {
			continue
		}
		switch attr.Kind {
		case INT32:
			if f, ok := val.(float64); ok {
				_ = s.SetInt32(attr.Name, int32(f))
			}
		case TEXT:
			if str, ok := val.(string); ok {
				_ = s.SetText(attr.Name, str)
			}
		case STRUCT:
			if subMap, ok := val.(map[string]interface{}); ok {
				sub := &StructInstance{layout: attr.Nested, buffer: s.buffer[attr.Offset : attr.Offset+attr.Size]}
				b, _ := json.Marshal(subMap)
				_ = sub.FromJSON(b)
			}
		}
	}
	return nil
}

func (s *StructInstance) CastTo(target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}
	v = v.Elem()
	if int(v.Type().Size()) != len(s.buffer) {
		return fmt.Errorf("size mismatch: struct=%d, target=%d", len(s.buffer), v.Type().Size())
	}
	ptr := unsafe.Pointer(v.UnsafeAddr())
	mem := unsafe.Slice((*byte)(ptr), len(s.buffer))
	copy(mem, s.buffer)
	return nil
}

func (s *StructInstance) field(name string, expect Kind) (*Attribute, error) {
	for i := range s.layout.attributes {
		if s.layout.attributes[i].Name == name {
			if s.layout.attributes[i].Kind != expect {
				return nil, fmt.Errorf("field %q type mismatch", name)
			}
			return &s.layout.attributes[i], nil
		}
	}
	return nil, fmt.Errorf("field %q not found", name)
}

func alignUp(x, align int) int {
	return (x + align - 1) & ^(align - 1)
}
