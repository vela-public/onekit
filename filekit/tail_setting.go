package filekit

import (
	"fmt"
	"github.com/vela-public/onekit/todo"
)

type Setting struct {
	Name   string   `lua:"name"`
	Limit  int      `lua:"limit"`
	Thread int      `lua:"thread"`
	Buffer int      `lua:"buffer"`
	Wait   int      `lua:"wait"`
	Delim  byte     `lua:"delim"`
	Follow bool     `lua:"follow"`
	Target []string `lua:"target"`
	//Bucket   []string `lua:"bucket"`
	FastJSON bool     `lua:"fastjson"`
	Location SeekInfo `lua:"location"`
	Poll     int      `lua:"poll"`
}

func Default(name string) *Setting {
	return &Setting{
		Name:   todo.IF(name == "", "filekit.tail", name),
		Limit:  0,
		Delim:  '\n',
		Wait:   10, // 10s
		Thread: 64,
		Follow: true,
		Poll:   3,
		Buffer: 4096,
		//Bucket: []string{"SHM_FILE_RECORD", strings.ToUpper(name)},
	}
}

func (s *Setting) Bad() error {
	if s.Name == "" {
		return fmt.Errorf("not found name")
	}
	return nil
}
