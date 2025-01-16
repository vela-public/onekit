package ssckit

import "github.com/vela-public/onekit/zapkit"

type Config struct {
	Protect bool           `json:"protect"`
	Logger  *zapkit.Config `json:"logger"`

	Node struct {
		DNS    string `json:"dns"`
		Prefix string `json:"prefix"`
	} `json:"node"`

	Extend []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"extend"`
}

func DefaultConfig() *Config {
	return &Config{}
}
