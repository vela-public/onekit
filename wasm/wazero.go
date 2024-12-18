package wasm

import (
	"context"
	"fmt"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type Module struct {
	mutex   sync.Mutex
	offset  uint32
	ctx     context.Context
	cfg     wazero.RuntimeConfig
	file    string
	mod     api.Module
	runtime wazero.Runtime
	buffer  api.Function
	error   *pipe.Chain
}

func (m *Module) NoError(e error) {
	if m.error.Len() == 0 {
		fmt.Printf("%v\n", e)
		return
	}
	m.error.Invoke(e.Error())
}

func (m *Module) Buffer(data string) (ptr uint32, sz uint32, err error) {
	sz = uint32(len(data))
	ptr = atomic.AddUint32(&m.offset, 1)
	mem := m.mod.Memory()
	if !mem.Write(ptr, []byte(data)) {
		return 0, 0, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d",
			ptr, sz, m.mod.Memory().Size())
	}
	atomic.AddUint32(&m.offset, sz)
	return
}

func (m *Module) String() string                         { return fmt.Sprintf("wasm://%s", m.file) }
func (m *Module) Type() lua.LValueType                   { return lua.LTObject }
func (m *Module) AssertFloat64() (float64, bool)         { return 0, false }
func (m *Module) AssertString() (string, bool)           { return "", false }
func (m *Module) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (m *Module) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (m *Module) show(ctx context.Context, mod api.Module, ptr uint32, size uint32) {
	buff, ok := mod.Memory().Read(ptr, size)
	if !ok {
		m.NoError(fmt.Errorf("show: memory overflow"))
	} else {
		m.NoError(fmt.Errorf(string(buff)))
	}
}

func NewModule(ctx context.Context, file string, args ...string) (*Module, error) {
	m := &Module{
		ctx:   ctx,
		file:  file,
		error: pipe.NewChain(),
	}

	cfg := wazero.NewRuntimeConfig()
	cfg.WithMemoryLimitPages(2)
	r := wazero.NewRuntimeWithConfig(ctx, cfg)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := wazero.NewModuleConfig()
	r.NewHostModuleBuilder("env").NewFunctionBuilder().WithFunc(m.show).Export("show").Instantiate(ctx)
	mod, err := r.InstantiateWithConfig(ctx, data, config.WithArgs(args...))
	if err != nil {
		return nil, err
	}

	m.mod = mod
	m.runtime = r
	return m, nil
}

/*
index
@key  fnc_text() //result: text
@key  fnc_json() //result: json
@key  fnc()      //result: number
*/

func (m *Module) Index(L *lua.LState, key string) lua.LValue {
	var typ ResultType
	switch {
	case strings.HasSuffix(key, "_text"):
		key = strings.TrimSuffix(key, "_text")
		typ = Text
	case strings.HasSuffix(key, "_json"):
		key = strings.TrimSuffix(key, "_json")
		typ = Json
	case strings.HasSuffix(key, "_bin"):
		key = strings.TrimSuffix(key, "_bin")
		typ = Bin
	case strings.HasSuffix(key, "_num"):
		key = strings.TrimSuffix(key, "_num")
		typ = Number
	default:
		typ = Number
	}

	fn := m.mod.ExportedFunction(key)
	return &Function{
		name:   key,
		parent: m,
		api:    fn,
		typ:    typ,
		handle: pipe.NewChain(),
	}
}
