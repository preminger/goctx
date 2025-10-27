package contextualize

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type StopReason int

const (
	StopReasonNone StopReason = iota
	StopReasonMain
	StopReasonHTTP
	StopReasonStopAt
)

// shouldStopAt evaluates termination conditions for the given enclosing function.
// Returns (true, reason) when we should not propagate further upward.
func shouldStopAt(funcDecl *ast.FuncDecl, pkg *packages.Package, opts Options, stopSpec *targetSpec) (bool, StopReason) {
	// stop-at specific
	if stopSpec != nil {
		if sameFile(pkg, funcDecl, stopSpec) && funcDecl.Name.Name == stopSpec.FuncName {
			// If no line number was provided, any matching function name in the file qualifies
			if stopSpec.LineNumber < 1 {
				return true, StopReasonStopAt
			}
			// When a line number was provided, ensure it matches the function's starting line
			start := pkg.Fset.Position(funcDecl.Pos()).Line
			if start == stopSpec.LineNumber {
				return true, StopReasonStopAt
			}
		}
	}

	// html handler boundary
	if opts.HTML {
		if isHTTPHandlerFunc(funcDecl, pkg) {
			return true, StopReasonHTTP
		}
	}

	// main termination
	if isMainFunction(funcDecl, pkg) {
		return true, StopReasonMain
	}

	return false, StopReasonNone
}

func isMainFunction(fn *ast.FuncDecl, pkg *packages.Package) bool {
	if fn == nil || fn.Recv != nil {
		return false
	}
	if fn.Name.Name != FuncNameMain {
		return false
	}
	if pkg.PkgPath != FuncNameMain && pkg.Name != FuncNameMain {
		return false
	}

	return true
}

func isHTTPHandlerFunc(fn *ast.FuncDecl, pkg *packages.Package) bool {
	if fn == nil {
		return false
	}
	params := fn.Type.Params
	if params == nil || len(params.List) != 2 {
		return false
	}
	// Second param should be *net/http.Request
	t := pkg.TypesInfo.TypeOf(params.List[1].Type)
	pt, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	namedType, ok := pt.Elem().(*types.Named)
	if !ok {
		return false
	}
	if namedType.Obj() == nil || namedType.Obj().Pkg() == nil {
		return false
	}
	if namedType.Obj().Name() == "Request" && namedType.Obj().Pkg().Path() == "net/http" {
		return true
	}

	return false
}

// ensureCtxAvailableAtBoundary ensures that inside fn, a ctx variable exists.
// If reason is StopReasonMain: inserts ctx := context.Background() at top if not present.
// If reason is OptNameHTTP: inserts ctx := <req>.Context() where <req> is the name of the *http.Request parameter.
func ensureCtxAvailableAtBoundary(pkg *packages.Package, file *ast.File, fn *ast.FuncDecl, reason StopReason) (bool, error) {
	if hasCtxInScope(fn, pkg) {
		return true, nil
	}
	switch reason {
	case StopReasonMain:
		ensureImport(pkg.Fset, file, "context")
		stmt := makeAssignCtxBackground()
		insertAfterLeadingBlankAssignsF(pkg.Fset, file, fn, stmt)
		return true, nil
	case StopReasonHTTP:
		reqName := findHTTPRequestParamName(fn, pkg)
		if reqName == "" {
			return false, errors.New("determining http request parameter name")
		}
		stmt := makeAssignCtxFromRequest(reqName)
		insertAfterLeadingBlankAssignsF(pkg.Fset, file, fn, stmt)
		return true, nil
	default:
		return false, nil
	}
}

// insertAfterLeadingBlankAssignsF works like insertAfterLeadingBlankAssigns but also
// adjusts the new statement's positions so that any trailing comments on the
// previous line stay attached to that previous statement.
func insertAfterLeadingBlankAssignsF(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, stmt ast.Stmt) {
	if fn == nil || fn.Body == nil {
		return
	}
	idx := 0
	for idx < len(fn.Body.List) {
		s := fn.Body.List[idx]
		as, ok := s.(*ast.AssignStmt)
		if !ok || len(as.Lhs) == 0 {
			break
		}
		if ident, ok := as.Lhs[0].(*ast.Ident); !ok || ident.Name != "_" {
			break
		}
		idx++
	}
	if idx > 0 {
		base := fn.Body.List[idx-1].End() + 1
		lastLine := fset.Position(fn.Body.List[idx-1].End()).Line
		// If there's a comment group on the same line, use its end as base.
		for _, cg := range file.Comments {
			if fset.Position(cg.Pos()).Line == lastLine {
				cend := cg.End()
				if cend > base {
					base = cend + 1
				}
				// Nudge the comment group's position slightly earlier so it remains a trailing
				// comment of the previous statement rather than being treated as a leading
				// comment for the newly inserted statement.
				for _, c := range cg.List {
					if fset.Position(c.Slash).Line == lastLine && c.Slash > fn.Body.List[idx-1].End() {
						c.Slash = fn.Body.List[idx-1].End() - 1
					}
				}
			}
		}
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			setAssignApproxPos(assign, base)
		}
	}
	fn.Body.List = append(fn.Body.List[:idx], append([]ast.Stmt{stmt}, fn.Body.List[idx:]...)...)
}

// setAssignApproxPos sets approximate token positions on an assignment statement to
// ensure it prints after a given base position, helping the formatter keep
// preceding trailing comments attached to their original lines.
func setAssignApproxPos(assign *ast.AssignStmt, base token.Pos) {
	assign.TokPos = base
	// LHS identifiers
	for _, lhs := range assign.Lhs {
		if id, ok := lhs.(*ast.Ident); ok {
			id.NamePos = base
		}
	}
	// RHS expressions (support common patterns we generate)
	for _, rhs := range assign.Rhs {
		switch rhsCast := rhs.(type) {
		case *ast.CallExpr:
			// Function part could be selector like x.Sel
			if sel, ok := rhsCast.Fun.(*ast.SelectorExpr); ok {
				if xid, ok := sel.X.(*ast.Ident); ok {
					xid.NamePos = base
				}
				sel.Sel.NamePos = base
			}
		case *ast.SelectorExpr:
			if xid, ok := rhsCast.X.(*ast.Ident); ok {
				xid.NamePos = base
			}
			rhsCast.Sel.NamePos = base
		}
	}
}

func makeAssignCtxBackground() ast.Stmt {
	// ctx := context.Background()
	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(VarNameCtx)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("context"), Sel: ast.NewIdent("Background")}}},
	}

	return assign
}

func makeAssignCtxFromRequest(req string) ast.Stmt {
	// ctx := <req>.Context()
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(VarNameCtx)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent(req), Sel: ast.NewIdent("Context")}}},
	}
}

func findHTTPRequestParamName(fn *ast.FuncDecl, p *packages.Package) string {
	if fn.Type.Params == nil {
		return ""
	}
	for _, field := range fn.Type.Params.List {
		t := p.TypesInfo.TypeOf(field.Type)
		pt, ok := t.(*types.Pointer)
		if !ok {
			continue
		}
		n, ok := pt.Elem().(*types.Named)
		if !ok || n.Obj() == nil || n.Obj().Pkg() == nil {
			continue
		}
		if n.Obj().Name() == "Request" && n.Obj().Pkg().Path() == "net/http" {
			if len(field.Names) > 0 {
				return field.Names[0].Name
			}
			return "req"
		}
	}

	return ""
}

func ensureCallHasCtxArg(_ *packages.Package, call *ast.CallExpr, ctxName string) {
	if ctxName == "" {
		ctxName = VarNameCtx
	}
	// Prepend ctx ident to arguments
	call.Args = append([]ast.Expr{ast.NewIdent(ctxName)}, call.Args...)
}

func getCtxIdentInScope(fn *ast.FuncDecl, pkg *packages.Package) string {
	// Prefer a parameter of type context.Context with a usable name (not underscore)
	if fn != nil && fn.Type != nil && fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if field == nil || field.Type == nil {
				continue
			}
			t := pkg.TypesInfo.TypeOf(field.Type)
			if types.TypeString(t, func(p *types.Package) string { return p.Path() }) != ContextContext {
				continue
			}
			if len(field.Names) > 0 {
				name := field.Names[0].Name
				if name != "_" && name != "" {
					return name
				}
			}
		}
	}
	// Otherwise, search identifiers used/defined in body with type context.Context
	var found string
	ast.Inspect(fn, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name == "_" || id.Name == "" {
			return true
		}
		obj := pkg.TypesInfo.Uses[id]
		if obj == nil {
			obj = pkg.TypesInfo.Defs[id]
		}
		if obj == nil {
			return true
		}
		// Only consider value identifiers (variables/params), ignore type names like the 'Context' in 'context.Context'.
		if v, ok := obj.(*types.Var); ok {
			if types.TypeString(v.Type(), func(p *types.Package) string { return p.Path() }) == ContextContext {
				found = id.Name
				return false
			}
		}
		return true
	})

	return found
}

func hasCtxInScope(fn *ast.FuncDecl, pkg *packages.Package) bool {
	return getCtxIdentInScope(fn, pkg) != ""
}
