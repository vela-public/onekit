package taskit

import (
	"encoding/json"
	"github.com/vela-public/onekit/libkit"
	"time"
)

type TaskType interface {
	Key() string
	Hash() string
	//todo
}

type ReloadType interface {
	Start() error
	Close() error
	Reload() error
}

type TaskEntry struct {
	ID      int64  `json:"id"`
	Dialect bool   `json:"dialect"`
	Name    string `json:"name"`
	Chunk   []byte `json:"chunk"`
	Hash    string `json:"hash"`
}

type TaskDiffInfo struct {
	Removes []int64      `json:"removes"`
	Updates []*TaskEntry `json:"updates"`
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

type TaskView struct {
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
}

type TreeView struct {
	Tasks []*TaskView `json:"tasks"`
}

func (tv *TreeView) Text() []byte {
	text, _ := json.Marshal(tv)
	return text
}
