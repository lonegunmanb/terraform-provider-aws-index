package pkg

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func parseSource(source string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, "", source, parser.ParseComments)
}
