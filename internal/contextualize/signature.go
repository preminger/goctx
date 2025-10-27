package contextualize

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// ensureFuncHasCtxParam ensures the function has a usable context.Context parameter.
// Behavior:
// - If there's a parameter of type context.Context named "_", rename it to "ctx" when renameBlank is true.
// - If there's any parameter of type context.Context with a different usable name, do nothing.
// - Otherwise, add a new first parameter "ctx context.Context".
// Returns true if the signature was modified.
func ensureFuncHasCtxParam(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, info *types.Info, renameBlank bool) bool {
	// Fast paths and guards
	if fn == nil || fn.Type == nil {
		return false
	}
	params := fn.Type.Params
	if params == nil {
		// No params at all: we will add one below
		return addCtxParamAsFirst(fset, file, fn)
	}

	// If a parameter explicitly named ctx already exists, do not add another
	if hasParamNamedCtx(params) {
		return false
	}

	// Look for any context.Context parameter in existing list
	for _, field := range params.List {
		if field == nil || field.Type == nil || !isContextType(info, field.Type) {
			continue
		}
		// Found a context parameter
		if len(field.Names) > 0 {
			// It's named
			if field.Names[0].Name == "_" && renameBlank {
				field.Names[0].Name = VarNameCtx
				field.Names[0].NamePos = token.NoPos
				ensureImport(fset, file, "context")
				return true
			}
			// Already named (or rename not requested): nothing to do
			return false
		}
		// Unnamed parameter of the right type: we can't reference it; conservatively add a named one in front.
		return addCtxParamAsFirst(fset, file, fn)
	}

	// No suitable existing parameter found: add a new one.

	return addCtxParamAsFirst(fset, file, fn)
}

// hasParamNamedCtx reports whether any parameter is named ctx.
func hasParamNamedCtx(params *ast.FieldList) bool {
	if params == nil {
		return false
	}
	for _, field := range params.List {
		for _, nm := range field.Names {
			if nm != nil && nm.Name == VarNameCtx {
				return true
			}
		}
	}

	return false
}

// isContextType reports whether expr denotes context.Context, using types.Info when available
// and falling back to a direct AST selector check.
func isContextType(info *types.Info, expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	if info != nil {
		if t := info.TypeOf(expr); t != nil {
			if types.TypeString(t, func(p *types.Package) string { return p.Path() }) == ContextContext {
				return true
			}
		}
	}
	// Fallback: check for selector expression context.Context
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if xid, ok := sel.X.(*ast.Ident); ok && xid.Name == "context" && sel.Sel != nil && sel.Sel.Name == "Context" {
			return true
		}
	}

	return false
}

// addCtxParamAsFirst inserts a new first parameter "ctx context.Context" and normalizes positions.
func addCtxParamAsFirst(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl) bool {
	ensureImport(fset, file, "context")
	ctxField := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(VarNameCtx)},
		Type:  &ast.SelectorExpr{X: ast.NewIdent("context"), Sel: ast.NewIdent("Context")},
	}
	if fn.Type.Params == nil {
		fn.Type.Params = &ast.FieldList{List: []*ast.Field{ctxField}}
	} else {
		fn.Type.Params.List = append([]*ast.Field{ctxField}, fn.Type.Params.List...)
	}
	// Reset positions to let the formatter render without spurious trailing commas
	fn.Type.Params.Opening = token.NoPos
	fn.Type.Params.Closing = token.NoPos
	for _, fld := range fn.Type.Params.List {
		for _, nm := range fld.Names {
			nm.NamePos = token.NoPos
		}
	}

	return true
}

func ensureImport(fset *token.FileSet, file *ast.File, path string) {
	if file == nil {
		return
	}
	// If already present, nothing to do.
	for _, imp := range file.Imports {
		if imp.Path != nil && strings.Trim(imp.Path.Value, "\"") == path {
			return
		}
	}
	// Snapshot old end line for each existing import path to preserve trailing comment association.
	oldLineByPath := map[string]int{}
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, s := range gd.Specs {
				if is, ok := s.(*ast.ImportSpec); ok && is.Path != nil {
					oldLineByPath[strings.Trim(is.Path.Value, "\"")] = fset.Position(is.End()).Line
				}
			}
		}
	}
	// Find (or create) the first import declaration.
	var importDecl *ast.GenDecl
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			importDecl = gd
			break
		}
	}
	newSpec := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: "\"" + path + "\""}}
	if importDecl == nil {
		// Let astutil create a proper single-line import with stable positions.
		astutil.AddImport(fset, file, path)
		// Refresh importDecl pointer after mutation
		for _, d := range file.Decls {
			if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
				importDecl = gd
				break
			}
		}
	} else {
		// Append then sort specs by path to get deterministic ordering without touching other fields.
		importDecl.Specs = append(importDecl.Specs, newSpec)
		sort.SliceStable(importDecl.Specs, func(i, j int) bool {
			isI, ok := importDecl.Specs[i].(*ast.ImportSpec)
			if !ok {
				return false
			}
			isJ, ok := importDecl.Specs[j].(*ast.ImportSpec)
			if !ok {
				return false
			}
			if isI == nil || isJ == nil || isI.Path == nil || isJ.Path == nil {
				return false
			}
			return strings.Trim(isI.Path.Value, "\"") < strings.Trim(isJ.Path.Value, "\"")
		})
	}
	// Rebuild file.Imports slice from all import decls to reflect current state.
	var imports []*ast.ImportSpec
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, s := range gd.Specs {
				if is, ok := s.(*ast.ImportSpec); ok {
					imports = append(imports, is)
				}
			}
		}
	}
	file.Imports = imports
	// Reposition any trailing comment groups that were on the same line as a spec before changes.
	repositionImportTrailingComments(fset, file, oldLineByPath)
}

// repositionImportTrailingComments reattaches comment groups that belonged to import specs
// identified by their old end-line positions before an import mutation.
func repositionImportTrailingComments(fset *token.FileSet, file *ast.File, oldLineByPath map[string]int) {
	if fset == nil || file == nil || len(file.Comments) == 0 || len(oldLineByPath) == 0 {
		return
	}
	// Build reverse map from old line -> path for quick lookup.
	lineToPath := make(map[int]string, len(oldLineByPath))
	for p, ln := range oldLineByPath {
		lineToPath[ln] = p
	}
	// Build current path -> spec mapping
	pathToSpec := map[string]*ast.ImportSpec{}
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, s := range gd.Specs {
				if is, ok := s.(*ast.ImportSpec); ok && is.Path != nil {
					pathToSpec[strings.Trim(is.Path.Value, "\"")] = is
				}
			}
		}
	}
	keep := make([]*ast.CommentGroup, 0, len(file.Comments))
	for _, cg := range file.Comments {
		if len(cg.List) == 0 {
			continue
		}
		line := fset.Position(cg.End()).Line
		if path, ok := lineToPath[line]; ok {
			if is := pathToSpec[path]; is != nil && is.Comment == nil {
				is.Comment = cg
				continue
			}
		}
		keep = append(keep, cg)
	}
	file.Comments = keep
}
