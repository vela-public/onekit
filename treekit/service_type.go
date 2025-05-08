package treekit

import (
	"encoding/json"
	"github.com/vela-public/onekit/libkit"
	"time"
)

type ReloadType interface {
	ProcessType
	Reload(*Env) error
}

type Script struct {
	Name string
	Path string
}

type ServiceEntry struct {
	ID      int64  `json:"id"`
	Dialect bool   `json:"dialect"`
	Name    string `json:"name"`
	Chunk   []byte `json:"chunk"`
	Hash    string `json:"hash"`
	MTime   int64  `json:"mtime"`
}

type Diff struct {
	Nothing []*ServiceView  `json:"nothing"`
	Removes []*ServiceEntry `json:"removes"`
	Updates []*ServiceEntry `json:"updates"`
}

func (d Diff) NotChange() bool {
	return d.Change() == 0
}

func (d Diff) Change() int {
	return len(d.Updates) + len(d.Removes)
}

func (d Diff) UpdateNames() []string {
	names := make([]string, len(d.Updates))
	for i, entry := range d.Updates {
		names[i] = entry.Name
	}
	return names
}

func (d Diff) RemoveNames() []string {
	names := make([]string, len(d.Removes))
	for i, entry := range d.Removes {
		names[i] = entry.Name
	}
	return names
}

type ServiceDiffInfo struct {
	Removes []int64         `json:"removes"`
	Updates []*ServiceEntry `json:"updates"`
}

type Runner struct {
	Name     string                     `json:"name"`
	Type     string                     `json:"type"`
	Status   string                     `json:"status"`
	CodeVM   string                     `json:"code_vm"`
	Private  bool                       `json:"private"`
	Cause    string                     `json:"cause"`
	Metadata libkit.DataKV[string, any] `json:"metadata"`
}
type LuaText struct {
	Name  string
	Text  string
	MTime time.Time
}

type ServiceView struct {
	ID      int64     `json:"id"`
	Dialect bool      `json:"dialect"`
	Name    string    `json:"name"`
	Link    string    `json:"link"`
	Status  string    `json:"status"`
	Hash    string    `json:"hash"`
	From    string    `json:"from"`
	Uptime  time.Time `json:"uptime"`
	Failed  bool      `json:"failed"`
	Cause   string    `json:"cause"`
	Runners []*Runner `json:"runners"`
	MTime   int64     `json:"mtime"`
}

type TreeView struct {
	Services []*ServiceView `json:"tasks"`
}

func (tv *TreeView) Text() []byte {
	text, _ := json.Marshal(tv)
	return text
}
