package goctx

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureImport_AddsOnceAndSorts(t *testing.T) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "f.go", "package p\n", parser.ParseComments)
	require.NoError(t, err)

	ensureImport(fset, astFile, "context")
	ensureImport(fset, astFile, "fmt")
	ensureImport(fset, astFile, "context")

	// expect two imports: context and fmt, sorted lexicographically
	assert.Len(t, astFile.Imports, 2)

	imports := []string{astFile.Imports[0].Path.Value, astFile.Imports[1].Path.Value}
	slices.Sort(imports)
	assert.Equal(t, []string{"\"context\"", "\"fmt\""}, imports)
}

func TestEnsureFuncHasCtxParam_AddsParam(t *testing.T) {
	fset := token.NewFileSet()
	f := &ast.File{}
	fn := &ast.FuncDecl{Type: &ast.FuncType{Params: &ast.FieldList{}}}
	// info can be nil; function should still add a ctx param conservatively
	ensureFuncHasCtxParam(fset, f, fn, nil, false)
	assert.NotNil(t, fn.Type.Params)
	assert.NotEmpty(t, fn.Type.Params.List)

	field := fn.Type.Params.List[0]
	assert.NotEmpty(t, field.Names)
	assert.Equal(t, VarNameCtx, field.Names[0].Name)
}
