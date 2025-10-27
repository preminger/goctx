package contextualize

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestEnsureImport_AddsOnceAndSorts(t *testing.T) {
	f := &ast.File{}
	ensureImport(f, "context")
	ensureImport(f, "fmt")
	ensureImport(f, "context")
	// expect two imports: context and fmt, sorted lexicographically
	if len(f.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(f.Imports))
	}
	got0 := f.Imports[0].Path.Value
	got1 := f.Imports[1].Path.Value
	if (got0 != "\"context\"" || got1 != "\"fmt\"") && (got0 != "\"fmt\"" || got1 != "\"context\"") {
		t.Fatalf("unexpected imports order: %s, %s", got0, got1)
	}
}

func TestEnsureFuncHasCtxParam_AddsParam(t *testing.T) {
	fset := token.NewFileSet()
	f := &ast.File{}
	fn := &ast.FuncDecl{Type: &ast.FuncType{Params: &ast.FieldList{}}}
	ensureFuncHasCtxParam(fset, f, fn)
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		t.Fatalf("expected a ctx param to be added")
	}
	field := fn.Type.Params.List[0]
	if len(field.Names) == 0 || field.Names[0].Name != VarNameCtx {
		t.Fatalf("expected first param to be named ctx")
	}
}
