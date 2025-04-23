package web

import "github.com/vela-public/onekit/webkit"

type Config struct {
	Name    string         `lua:"name"`
	Cluster []string       `lua:"cluster"`
	Bind    string         `lua:"bind"`
	Router  *webkit.Router `lua:"-"`
}

func (cnf *Config) Must() {
	cnf.Router = webkit.NewRouter()
}
