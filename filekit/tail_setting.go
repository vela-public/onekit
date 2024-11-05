package filekit

import (
	"fmt"
	"strings"
)

type Setting struct {
	Name     string   `lua:"name"`
	Limit    int      `lua:"limit"`
	Thread   int      `lua:"thread"`
	Buffer   int      `lua:"buffer"`
	Wait     int      `lua:"wait"`
	Delim    byte     `lua:"delim"`
	Follow   bool     `lua:"follow"`
	Target   []string `lua:"target"`
	Bucket   []string `lua:"bucket"`
	FastJSON bool     `lua:"fastjson"`
	Poll     int      `lua:"poll"`
}

func Default(name string) *Setting {
	return &Setting{
		Name:   name,
		Limit:  0,
		Delim:  '\n',
		Wait:   30, // 30s
		Thread: 64,
		Follow: true,
		Poll:   10,
		Bucket: []string{"SHM_FILE_RECORD", strings.ToUpper(name)},
	}
}

func (s *Setting) Bad() error {
	if s.Name == "" {
		return fmt.Errorf("not found name")
	}
	if len(s.Bucket) == 0 {
		return fmt.Errorf("not found bucket")
	}

	return nil
}
