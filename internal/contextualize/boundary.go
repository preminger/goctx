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
func shouldStopAt(fn *ast.FuncDecl, p *packages.Package, opts Options, stopSpec *targetSpec) (bool, StopReason) {
	// stop-at specific
	if stopSpec != nil {
		if sameFile(p, fn, stopSpec) && fn.Name.Name == stopSpec.FuncName {
			if stopSpec.Ordinal < 1 {
				return true, StopReasonStopAt
			}

			// If ordinal was provided, ensure it matches
			if idx := ordinalOfFuncInFile(p, fn, stopSpec); idx == stopSpec.Ordinal {
				return true, StopReasonStopAt
			}
		}
	}

	// html handler boundary
	if opts.HTML {
		if isHTTPHandlerFunc(fn, p) {
			return true, StopReasonHTTP
		}
	}

	// main termination
	if isMainFunction(fn, p) {
		return true, StopReasonMain
	}
	return false, StopReasonNone
}

func isMainFunction(fn *ast.FuncDecl, p *packages.Package) bool {
	if fn == nil || fn.Recv != nil {
		return false
	}
	if fn.Name.Name != FuncNameMain {
		return false
	}
	if p.PkgPath != FuncNameMain && p.Name != FuncNameMain {
		return false
	}
	return true
}

func isHTTPHandlerFunc(fn *ast.FuncDecl, p *packages.Package) bool {
	if fn == nil {
		return false
	}
	params := fn.Type.Params
	if params == nil || len(params.List) != 2 {
		return false
	}
	// Second param should be *net/http.Request
	t := p.TypesInfo.TypeOf(params.List[1].Type)
	pt, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	n, ok := pt.Elem().(*types.Named)
	if !ok {
		return false
	}
	if n.Obj() == nil || n.Obj().Pkg() == nil {
		return false
	}
	if n.Obj().Name() == "Request" && n.Obj().Pkg().Path() == "net/http" {
		return true
	}
	return false
}

// ensureCtxAvailableAtBoundary ensures that inside fn, a ctx variable exists.
// If reason is StopReasonMain: inserts ctx := context.Background() at top if not present.
// If reason is OptNameHTTP: inserts ctx := <req>.Context() where <req> is the name of the *http.Request parameter.
func ensureCtxAvailableAtBoundary(p *packages.Package, file *ast.File, fn *ast.FuncDecl, reason StopReason) (bool, error) {
	if hasCtxInScope(fn, p) {
		return true, nil
	}
	switch reason {
	case StopReasonMain:
		ensureImport(file, "context")
		stmt := makeAssignCtxBackground()
		fn.Body.List = append([]ast.Stmt{stmt}, fn.Body.List...)
		return true, nil
	case StopReasonHTTP:
		reqName := findHTTPRequestParamName(fn, p)
		if reqName == "" {
			return false, errors.New("determining http request parameter name")
		}
		stmt := makeAssignCtxFromRequest(reqName)
		fn.Body.List = append([]ast.Stmt{stmt}, fn.Body.List...)
		return true, nil
	default:
		return false, nil
	}
}

func makeAssignCtxBackground() ast.Stmt {
	// ctx := context.Background()
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(VarNameCtx)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("context"), Sel: ast.NewIdent("Background")}},
	}
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

func ensureCallHasCtxArg(_ *packages.Package, call *ast.CallExpr) {
	// Prepend ctx ident to arguments
	call.Args = append([]ast.Expr{ast.NewIdent(VarNameCtx)}, call.Args...)
}

func hasCtxInScope(fn *ast.FuncDecl, p *packages.Package) bool {
	found := false
	ast.Inspect(fn, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name != VarNameCtx {
			return true
		}
		// Check object type
		obj := p.TypesInfo.Uses[id]
		if obj == nil {
			obj = p.TypesInfo.Defs[id]
		}
		if obj == nil {
			return true
		}
		if types.TypeString(obj.Type(), func(p *types.Package) string { return p.Path() }) == "context.Context" {
			found = true
			return false
		}
		return true
	})
	return found
}
