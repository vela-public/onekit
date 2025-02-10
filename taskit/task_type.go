package taskit

import "github.com/vela-public/onekit/libkit"

type TaskType interface {
	Key() string
	Hash() string
	Metadata() libkit.DataKV[string, any]
	View() *TaskView
}
