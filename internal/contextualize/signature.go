package contextualize

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strconv"
	"strings"
)

func funcHasCtxParam(fn *ast.FuncDecl, info *types.Info) bool {
	if fn.Type.Params == nil {
		return false
	}
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			if name.Name == VarNameCtx {
				t := info.TypeOf(field.Type)
				if types.TypeString(t, func(p *types.Package) string { return p.Path() }) == "context.Context" {
					return true
				}
			}
		}
	}
	return false
}

func ensureFuncHasCtxParam(_ *token.FileSet, file *ast.File, fn *ast.FuncDecl) {
	// Add import if necessary
	ensureImport(file, "context")
	// Prepend parameter ctx context.Context
	ctxField := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(VarNameCtx)},
		Type:  &ast.SelectorExpr{X: ast.NewIdent("context"), Sel: ast.NewIdent("Context")},
	}
	if fn.Type.Params == nil {
		fn.Type.Params = &ast.FieldList{List: []*ast.Field{ctxField}}
	} else {
		fn.Type.Params.List = append([]*ast.Field{ctxField}, fn.Type.Params.List...)
	}
}

func ensureImport(file *ast.File, path string) {
	for _, imp := range file.Imports {
		if strings.Trim(imp.Path.Value, "\"") == path {
			return
		}
	}
	file.Imports = append(file.Imports, &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)}})
	// Maintain sorted order in AST.Imports for determinism
	sort.Slice(file.Imports, func(i, j int) bool {
		return strings.Trim(file.Imports[i].Path.Value, "\"") < strings.Trim(file.Imports[j].Path.Value, "\"")
	})
}
