package contextualize

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
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

func ensureFuncHasCtxParam(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl) {
	// Add import if necessary
	ensureImport(fset, file, "context")
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
	// Reset positions to let the formatter render without spurious trailing commas
	fn.Type.Params.Opening = token.NoPos
	fn.Type.Params.Closing = token.NoPos
	for _, fld := range fn.Type.Params.List {
		for _, nm := range fld.Names {
			nm.NamePos = token.NoPos
		}
	}
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
