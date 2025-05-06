package treekit

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"io"
	"os"
	"strings"
)

func number(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func alphabet(ch byte) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}

	if ch >= 'A' && ch <= 'Z' {
		return true
	}

	return false
}

func Name(v string) error {
	if len(v) < 2 {
		return fmt.Errorf("too short name got:%s", v)
	}

	if !alphabet(v[0]) {
		return fmt.Errorf("first char must be a-z or A-Z got:%v", string(v[0]))
	}

	if strings.HasPrefix(v, "GET /") ||
		strings.HasPrefix(v, "POST /") ||
		strings.HasPrefix(v, "PUT /") ||
		strings.HasPrefix(v, "DELETE /") ||
		strings.HasPrefix(v, "HEAD /") ||
		strings.HasPrefix(v, "PATCH /") ||
		strings.HasPrefix(v, "OPTIONS /") {

		if offset := strings.IndexFunc(v, func(r rune) bool {
			return r == '/'
		}); offset != -1 {
			v = v[offset:]
		}
	}

	sz := len(v)
	for i := 1; i < sz; i++ {
		ch := v[i]
		switch {
		case alphabet(ch), number(ch):
			continue
		case ch == '_':
			continue
		case ch == '-':
			continue
		case ch == '/':
			continue
		default:
			return fmt.Errorf("not allowed char %v", string(ch))

		}
	}
	return nil
}

func Check[T any](L *lua.LState, pro *Process) (t T) {
	if pro.Nil() {
		L.RaiseError("not found processes data")
		return
	}

	dat, ok := pro.data.(T)
	if !ok {
		L.RaiseError("mismatch processes type must:%T got:%T", t, pro.data)
		return t
	}

	return dat
}

func Startup(L *lua.LState, process ProcessType, envs ...TreeEnvFunc) {
	env := &Env{}
	for _, fn := range envs {
		fn(env)
	}

	if env.ctx == nil {
		env.ctx = L.Context()
	}

	if env.lua == nil {
		env.lua = L
	}

	if env.err == nil {
		env.err = func(e error) {
			L.RaiseError("startup fail error %v", e)
		}
	}

	exdata := L.Exdata()

	switch dat := exdata.(type) {
	case *MicroService:
		dat.Startup(process, env)
	case *Task:
		dat.Startup(process, env)
	default:
		env.err(fmt.Errorf("lua.exdata must *MicroService or *TaskTree got:%T", exdata))
		L.RaiseError("lua.exdata must *MicroService or *TaskTree got:%T", exdata)
	}

}

func Start(L *lua.LState, process ProcessType, x func(e error)) {
	Startup(L, process, func(v *Env) {
		v.err = x
		v.lua = L
		v.ctx = L.Context()
	})
}

func Close(L *lua.LState, v ProcessType, x func(e error)) {
	exdata := L.Exdata()
	switch dat := exdata.(type) {
	case *MicroService:
		dat.Shutdown(v, x)
	case *Task:
		dat.Shutdown(v, x)
	default:
		x(fmt.Errorf("lua.exdata must *MicroService or *TaskTree got:%T", exdata))
	}
}

func Read(name string, path string) (*ServiceEntry, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	s := &ServiceEntry{
		Dialect: true,
		Name:    name,
	}

	st, err := fd.Stat()
	if err == nil {
		s.MTime = st.ModTime().Unix()
	}

	m5 := md5.New()
	buf := bytes.NewBuffer(nil)
	w := io.MultiWriter(m5, buf)
	_, err = io.Copy(w, fd)
	if err != nil && err != io.EOF {
		return s, err
	}

	s.Chunk = buf.Bytes()
	s.Hash = fmt.Sprintf("%x", m5.Sum(nil))
	return s, nil
}

func Load(s ...Script) ([]*ServiceEntry, error) {
	errs := errkit.New()
	var ss []*ServiceEntry
	for _, v := range s {
		if v.Name == "" {
			continue
		}
		if v.Path == "" {
			continue
		}

		entry, err := Read(v.Name, v.Path)
		if err != nil {
			errs.Try(v.Name, err)
			continue
		}
		ss = append(ss, entry)
	}
	return ss, errs.Wrap()
}
