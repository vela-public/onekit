package syntax

import (
	"github.com/vela-public/onekit/grep/syntax/ast"
	"github.com/vela-public/onekit/grep/syntax/lexer"
)

func Parse(s string) (*ast.Node, error) {
	return ast.Parse(lexer.NewLexer(s))
}

func Special(b byte) bool {
	return lexer.Special(b)
}
