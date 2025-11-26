package goctx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Options configures the goctx run.
// WorkDir should point at the module root (or any subdir); we will load ./...
// Target syntax: path/to/file.go:FuncName[:N]
// StopAt optional syntax: same as Target.
// HTML: if true, stop when reaching http.HandlerFunc boundary and derive ctx from req.Context().
type Options struct {
	Target  string
	StopAt  string
	HTML    bool
	WorkDir string
}

// Run performs the goctx according to Options.
func Run(_ context.Context, opts Options) error {
	if opts.Target == "" {
		return errors.New("missing target argument")
	}

	// Load all packages in the workspace
	pkgs, err := loadAllPackages(firstNonEmpty(opts.WorkDir, "."))
	if err != nil {
		return err
	}

	// Parse target and optional stopAt
	tgtSpec, err := parseTargetSpec(opts.Target)
	if err != nil {
		return fmt.Errorf("parsing target: %w", err)
	}
	stopSpec, err := parseStopSpec(opts.StopAt)
	if err != nil {
		var noStop noStopSpecError
		if !errors.As(err, &noStop) {
			return err
		}
	}

	// Find the package and file for the target
	res, err := resolveTarget(pkgs, tgtSpec)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}

	modifiedFiles := map[string]bool{}

	// Decide if target already has a usable context.Context parameter (reuse case)
	reuseExistingCtxInTarget := functionHasContextParam(res.Decl, res.Info)

	// Ensure target function has ctx param (do not rename blank yet)
	ensureTargetHasCtx(res, modifiedFiles)

	// Traverse callers recursively and propagate ctx as needed, unless the target already has a context parameter
	var sawAnyCall bool
	if !reuseExistingCtxInTarget {
		if err := traverseAndPropagate(pkgs, res.Obj, opts, stopSpec, modifiedFiles, &sawAnyCall); err != nil {
			return err
		}
	}

	// If the target has a blank-named context param and there are no callers,
	// rename it to ctx (covers the dedicated rename test case) without affecting
	// the case where callers exist and we should preserve '_'.
	maybeRenameBlankCtxInTarget(res, modifiedFiles, sawAnyCall)

	// Write back modified files
	if err := writeModified(pkgs); err != nil {
		return fmt.Errorf("writing modified files: %w", err)
	}

	return nil
}

// loadAllPackages loads all packages in the workspace rooted at dir.
func loadAllPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}
	// Do not abort on initial type errors; we will be editing files to fix them.
	// Still print errors for visibility but continue.
	_ = packages.PrintErrors(pkgs)

	return pkgs, nil
}

// maybeRenameBlankCtxInTarget renames a blank-named context parameter to ctx for the target function
// only when no callers were found during traversal (standalone function case).
func maybeRenameBlankCtxInTarget(res *targetResolution, modifiedFiles map[string]bool, sawAnyCall bool) {
	if res == nil || res.Decl == nil || res.FileAST == nil || res.Fset == nil {
		return
	}
	if sawAnyCall {
		return // there are callers; preserve '_'
	}
	if ensureFuncHasCtxParam(res.Fset, res.FileAST, res.Decl, res.Info, true) {
		markFileModified(modifiedFiles, res.Fset, res.FileAST)
	}
}

// noStopSpecError is a sentinel error indicating the user did not provide a stop-at spec.
// It allows callers to distinguish between "no spec provided" and real errors, avoiding nil-nil returns.
type noStopSpecError struct{}

func (noStopSpecError) Error() string { return "no stop-at spec provided" }

// parseStopSpec parses the optional stop-at spec.
// When no spec is provided (empty/whitespace), it returns (nil, noStopSpecError{}).
func parseStopSpec(stopAt string) (*targetSpec, error) {
	if strings.TrimSpace(stopAt) == "" {
		return nil, noStopSpecError{}
	}
	ss, err := parseTargetSpec(stopAt)
	if err != nil {
		return nil, fmt.Errorf("parsing stop-at: %w", err)
	}

	return &ss, nil
}

// ensureTargetHasCtx guarantees the target function has a ctx parameter and marks file modified.
func ensureTargetHasCtx(res *targetResolution, modifiedFiles map[string]bool) {
	if ensureFuncHasCtxParam(res.Fset, res.FileAST, res.Decl, res.Info, false) {
		modifiedFiles[res.FileAST.Name.Name] = true // marker by pkg name; we'll use filenames later
		markFileModified(modifiedFiles, res.Fset, res.FileAST)
	}
}

// traverseAndPropagate walks callers recursively from start and ensures ctx propagation.
func traverseAndPropagate(pkgs []*packages.Package, start types.Object, opts Options, stopSpec *targetSpec, modifiedFiles map[string]bool, sawAnyCall *bool) error {
	visited := map[types.Object]bool{}
	queue := []types.Object{start}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true

		for _, pkg := range pkgs {
			for fileIndex, fileAST := range pkg.Syntax {
				fi := pkg.Fset.File(fileAST.Pos())
				if fi == nil {
					continue
				}
				params := processCallSitesParams{
					pkg:           pkg,
					fileIndex:     fileIndex,
					fileAST:       fileAST,
					curr:          curr,
					opts:          opts,
					stopSpec:      stopSpec,
					modifiedFiles: modifiedFiles,
					queue:         &queue,
					sawAnyCall:    sawAnyCall,
				}
				if err := processCallSites(params); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// processCallSites scans a single file for calls to curr and performs required modifications.
type processCallSitesParams struct {
	pkg           *packages.Package
	fileIndex     int
	fileAST       *ast.File
	curr          types.Object
	opts          Options
	stopSpec      *targetSpec
	modifiedFiles map[string]bool
	queue         *[]types.Object
	sawAnyCall    *bool
}

func processCallSites(params processCallSitesParams) error {
	var inspectErr error
	ast.Inspect(params.fileAST, func(n ast.Node) bool {
		if inspectErr != nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		var calledObj types.Object
		var funName string
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			calledObj = params.pkg.TypesInfo.Uses[fun]
			funName = fun.Name
		case *ast.SelectorExpr:
			calledObj = params.pkg.TypesInfo.Uses[fun.Sel]
			funName = fun.Sel.Name
		}
		match := calledObj == params.curr
		if !match {
			// Fallback by name and package path to be resilient to partial type info
			if funName != "" && params.curr != nil && params.curr.Pkg() != nil && params.pkg.PkgPath == params.curr.Pkg().Path() && funName == params.curr.Name() {
				match = true
			}
		}
		if !match {
			return true
		}
		// Found a call. Mark that at least one call site exists.
		if params.sawAnyCall != nil {
			*params.sawAnyCall = true
		}
		// Find enclosing function decl.
		enc := enclosingFuncDecl(params.fileAST, call)
		if enc == nil {
			return true
		}

		stopHere, stopReason := shouldStopAt(enc, params.pkg, params.opts, params.stopSpec)

		if stopHere {
			// At stop boundary: ensure a ctx exists, derive if necessary (main/http) and always pass ctx to call
			if _, err := ensureCtxAvailableAtBoundary(params.pkg, params.fileAST, enc, stopReason); err != nil {
				inspectErr = fmt.Errorf("ensuring ctx at stop boundary: %w", err)
				return false
			}
			ctxName := getCtxIdentInScope(enc, params.pkg)
			ensureCallHasCtxArg(params.pkg, call, ctxName)
			markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])
			return true
		}

		// If ctx in scope, just pass; do not enqueue since callers already pass their own context
		if hasCtxInScope(enc, params.pkg) {
			ctxName := getCtxIdentInScope(enc, params.pkg)
			ensureCallHasCtxArg(params.pkg, call, ctxName)
			markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])
			return true
		}

		// Determine whether the enclosing function already has a context parameter (possibly named "_")
		hadCtxParam := functionHasContextParam(enc, params.pkg.TypesInfo)
		// Ensure a usable ctx param exists (adds one if missing, or renames '_' to 'ctx')
		ensureFuncHasCtxParam(params.pkg.Fset, params.fileAST, enc, params.pkg.TypesInfo, true)
		ctxName := getCtxIdentInScope(enc, params.pkg)
		ensureCallHasCtxArg(params.pkg, call, ctxName)
		// Mark file modified (either signature or call site changed)
		markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])

		// Only enqueue if we had to ADD a brand new context parameter. If we are reusing an existing one,
		// do not traverse further; callers of this function already pass their context argument.
		if !hadCtxParam {
			if def := params.pkg.TypesInfo.Defs[enc.Name]; def != nil {
				*params.queue = append(*params.queue, def)
			}
		}
		return true
	})

	return inspectErr
}

// Helper: mark the concrete filename modified.
func markFileModified(mod map[string]bool, fset *token.FileSet, file *ast.File) {
	if fset == nil || file == nil {
		return
	}
	if file.Pos().IsValid() {
		if f := fset.File(file.Pos()); f != nil {
			mod[f.Name()] = true
		}
	}
}

// writeModified writes all package files back using canonical gofmt formatting.
func writeModified(pkgs []*packages.Package) error {
	for _, p := range pkgs {
		for _, syntaxTree := range p.Syntax {
			filename := p.Fset.File(syntaxTree.Pos()).Name()
			var buf bytes.Buffer
			// Format using go/format to preserve standard gofmt style and comments
			if err := format.Node(&buf, p.Fset, syntaxTree); err != nil {
				return fmt.Errorf("formatting file %s: %w", filename, err)
			}
			if err := os.WriteFile(filename, buf.Bytes(), 0o644); err != nil { //nolint:gosec // Appropriate permissions for source files
				return fmt.Errorf("writing file %s: %w", filename, err)
			}
		}
	}

	return nil
}

// Utility.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}

	return ""
}

// functionHasContextParam reports whether the given function declaration already has
// a parameter of type context.Context, regardless of its name. This influences whether
// we need to traverse callers: when true, callers already pass the corresponding argument.
func functionHasContextParam(fn *ast.FuncDecl, info *types.Info) bool {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return false
	}
	for _, field := range fn.Type.Params.List {
		if field == nil || field.Type == nil {
			continue
		}
		if isContextType(info, field.Type) {
			return true
		}
	}

	return false
}
