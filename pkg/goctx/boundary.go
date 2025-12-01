package goctx

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"

	"golang.org/x/tools/go/packages"
)

type StopReason int

const (
	StopReasonNone StopReason = iota
	StopReasonMain
	StopReasonHTTP
	StopReasonTest
	StopReasonStopAt
)

// shouldStopAt evaluates termination conditions for the given enclosing function.
// Returns (true, reason) when we should not propagate further upward.
func shouldStopAt(funcDecl *ast.FuncDecl, pkg *packages.Package, opts Options, stopSpec *targetSpec) (bool, StopReason, error) {
	// stop-at specific
	if stopSpec != nil {
		isSameFile, err := sameFile(pkg, funcDecl, stopSpec)
		if err != nil {
			return false, StopReasonNone, fmt.Errorf("determining if stop-at is in same file: %w", err)
		}

		if isSameFile && funcDecl.Name.Name == stopSpec.FuncName {
			// If no line number was provided, any matching function name in the file qualifies
			if stopSpec.LineNumber < 1 {
				slog.Debug("stopAt matched by name", slog.String("func", funcDecl.Name.Name))
				return true, StopReasonStopAt, nil
			}
			// When a line number was provided, ensure it matches the function's starting line
			start := pkg.Fset.Position(funcDecl.Pos()).Line
			if start == stopSpec.LineNumber {
				slog.Debug("stopAt matched by line", slog.String("func", funcDecl.Name.Name), slog.Int("line", start))
				return true, StopReasonStopAt, nil
			}
		}
	}

	// testing boundary: any function with testing.T, testing.B, testing.F, or testing.TB (or pointer)
	if isTestingBoundary(funcDecl, pkg) {
		slog.Debug("stop at testing boundary", slog.String("func", funcDecl.Name.Name))
		return true, StopReasonTest, nil
	}

	// html handler boundary
	if opts.HTML {
		if isHTTPHandlerFunc(funcDecl, pkg) {
			slog.Debug("stop at HTTP boundary", slog.String("func", funcDecl.Name.Name))
			return true, StopReasonHTTP, nil
		}
	}

	// main termination
	if isMainFunction(funcDecl, pkg) {
		slog.Debug("stop at main function", slog.String("func", funcDecl.Name.Name))
		return true, StopReasonMain, nil
	}

	return false, StopReasonNone, nil
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
		slog.Debug("ctx already in scope at boundary", slog.String("func", fn.Name.Name))
		return true, nil
	}
	switch reason {
	case StopReasonMain:
		ensureImport(pkg.Fset, file, "context")
		stmt := makeAssignCtxBackground()
		insertAfterLeadingBlankAssignsF(pkg.Fset, file, fn, stmt)
		slog.Debug("inserted ctx := context.Background()", slog.String("func", fn.Name.Name))
		return true, nil
	case StopReasonHTTP:
		reqName := findHTTPRequestParamName(fn, pkg)
		if reqName == "" {
			return false, errors.New("determining http request parameter name")
		}
		stmt := makeAssignCtxFromRequest(reqName)
		insertAtFuncStartF(pkg.Fset, file, fn, stmt)
		slog.Debug("inserted ctx := req.Context()", slog.String("func", fn.Name.Name), slog.String("req", reqName))
		return true, nil
	case StopReasonTest:
		testName := findTestingParamName(fn, pkg)
		if testName == "" {
			// Fall back to background if we cannot determine a testing param name (should be rare)
			ensureImport(pkg.Fset, file, "context")
			stmt := makeAssignCtxBackground()
			insertAtFuncStartF(pkg.Fset, file, fn, stmt)
			slog.Debug("inserted ctx := context.Background() (fallback for testing boundary)", slog.String("func", fn.Name.Name))
			return true, nil
		}
		stmt := makeAssignCtxFromTesting(testName)
		// For testing boundaries, ensure ctx is initialized BEFORE any statements (including
		// leading blank assigns like `_ = HelperTarget(...)`) so that those calls can use ctx.
		insertAtFuncStartF(pkg.Fset, file, fn, stmt)
		slog.Debug("inserted ctx := t.Context()", slog.String("func", fn.Name.Name), slog.String("t", testName))
		return true, nil
	default:
		return false, nil
	}
}

// insertAfterLeadingBlankAssignsF inserts a statement after leading blank assigns.
// Adjusts formatting and positions to maintain proper syntax and style.
// Handles positioning relative to comments or existing statements in the function.
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
	// Compute a base position for nicer formatting.
	var base token.Pos
	if idx > 0 {
		base = fn.Body.List[idx-1].End() + 1
		lastLine := fset.Position(fn.Body.List[idx-1].End()).Line
		// If there's a comment group on the same line, use its end as base and nudge
		// its tokens so it remains trailing on the previous statement.
		for _, cg := range file.Comments {
			if fset.Position(cg.Pos()).Line == lastLine {
				cend := cg.End()
				if cend > base {
					base = cend + 1
				}
				for _, c := range cg.List {
					if fset.Position(c.Slash).Line == lastLine && c.Slash > fn.Body.List[idx-1].End() {
						c.Slash = fn.Body.List[idx-1].End() - 1
					}
				}
			}
		}
	} else {
		// idx == 0: place after any leading comment groups present before the first statement.
		if len(fn.Body.List) > 0 {
			firstStmtPos := fn.Body.List[0].Pos()
			base = fn.Body.Lbrace + 1
			for _, cg := range file.Comments {
				if cg.Pos() > fn.Body.Lbrace && cg.End() < firstStmtPos {
					if cg.End()+1 > base {
						base = cg.End() + 1
					}
				}
			}
		}
	}
	if assign, ok := stmt.(*ast.AssignStmt); ok && base != token.NoPos {
		setAssignApproxPos(assign, base)
	}
	fn.Body.List = append(fn.Body.List[:idx], append([]ast.Stmt{stmt}, fn.Body.List[idx:]...)...)
}

// insertAtFuncStartF inserts stmt as the first statement of fn.Body, adjusting
// token positions so formatting is stable and existing comments remain attached
// to their intended lines.
func insertAtFuncStartF(fset *token.FileSet, file *ast.File, fn *ast.FuncDecl, stmt ast.Stmt) {
	if fn == nil || fn.Body == nil {
		return
	}
	fn.Body.List = append([]ast.Stmt{stmt}, fn.Body.List...)
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

func makeAssignCtxFromTesting(tvar string) ast.Stmt {
	// ctx := <tvar>.Context()
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(VarNameCtx)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent(tvar), Sel: ast.NewIdent("Context")}}},
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

func ensureCallHasCtxArg(pkg *packages.Package, call *ast.CallExpr, ctxName string) {
	if ctxName == "" {
		ctxName = VarNameCtx
	}
	// If there's already a first argument and it's either the same identifier name
	// or it is of type context.Context, avoid adding a duplicate.
	if len(call.Args) > 0 {
		// Case 1: first arg is an ident with the same name (ctx)
		if id, ok := call.Args[0].(*ast.Ident); ok {
			if id.Name == ctxName {
				return
			}
		}
		// Case 2: first arg type is context.Context
		if pkg != nil && pkg.TypesInfo != nil {
			if t := pkg.TypesInfo.TypeOf(call.Args[0]); t != nil {
				if types.TypeString(t, func(p *types.Package) string { return p.Path() }) == ContextContext {
					return
				}
			}
		}
	}

	// Prepend ctx ident to arguments
	call.Args = append([]ast.Expr{ast.NewIdent(ctxName)}, call.Args...)
}

func getCtxIdentInScope(fn *ast.FuncDecl, pkg *packages.Package) string {
	// 1) Prefer a parameter literally named "ctx" (regardless of type info availability)
	if fn != nil && fn.Type != nil && fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				if name != nil && name.Name == VarNameCtx {
					return VarNameCtx
				}
			}
		}
	}
	// 2) Prefer any parameter of type context.Context with a usable name (not underscore)
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
	// 3) If there's any local identifier literally named "ctx" in this function's body (e.g., ctx := ...), reuse it.
	var found string
	ast.Inspect(fn, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if id.Name == VarNameCtx {
				found = VarNameCtx
				return false
			}
		}
		return true
	})
	if found != "" {
		return found
	}
	// 4) Finally, search identifiers with type context.Context via types info (best-effort when info is present)
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

// isTestingBoundary reports whether the function has a parameter of type testing.T, testing.B,
// testing.F, or testing.TB (or a pointer to any of these), indicating a Go test entry point.
func isTestingBoundary(fn *ast.FuncDecl, pkg *packages.Package) bool {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return false
	}
	for _, field := range fn.Type.Params.List {
		if isTestingParamType(pkg, field.Type) {
			return true
		}
	}
	return false
}

func isTestingParamType(pkg *packages.Package, expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	theType := pkg.TypesInfo.TypeOf(expr)
	// Unwrap pointers
	if pt, ok := theType.(*types.Pointer); ok {
		theType = pt.Elem()
	}
	// Named types from testing package: T, B, F, TB
	if named, ok := theType.(*types.Named); ok {
		if named.Obj() == nil || named.Obj().Pkg() == nil {
			return false
		}
		if named.Obj().Pkg().Path() != "testing" {
			return false
		}
		switch named.Obj().Name() {
		case "T", "B", "F", "TB":
			return true
		}
	}
	return false
}

// findTestingParamName returns the identifier name of the testing parameter (t, b, f, tb, etc.).
// If unnamed, it returns a sensible default "t".
func findTestingParamName(fn *ast.FuncDecl, pkg *packages.Package) string {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return ""
	}
	for _, field := range fn.Type.Params.List {
		if isTestingParamType(pkg, field.Type) {
			if len(field.Names) > 0 && field.Names[0] != nil && field.Names[0].Name != "" && field.Names[0].Name != "_" {
				return field.Names[0].Name
			}
			return "t"
		}
	}
	return ""
}
