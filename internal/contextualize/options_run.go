package contextualize

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Options configures the contextualization run.
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

// Run performs the contextualization according to Options.
func Run(_ context.Context, opts Options) error {
	if opts.Target == "" {
		return errors.New("missing target argument")
	}
	// Load all packages in the workspace
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Dir: firstNonEmpty(opts.WorkDir, "."),
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return errors.New("loading packages: encountered errors")
	}

	// Parse target and optional stopAt
	tgtSpec, err := parseTargetSpec(opts.Target)
	if err != nil {
		return fmt.Errorf("parsing target: %w", err)
	}
	var stopSpec *targetSpec
	if strings.TrimSpace(opts.StopAt) != "" {
		ss, err := parseTargetSpec(opts.StopAt)
		if err != nil {
			return fmt.Errorf("parsing stop-at: %w", err)
		}
		stopSpec = &ss
	}

	// Find the package and file for the target
	res, err := resolveTarget(pkgs, tgtSpec)
	if err != nil {
		return fmt.Errorf("resolving target: %w", err)
	}

	modifiedFiles := map[string]bool{}

	// Ensure target function has ctx param
	if !funcHasCtxParam(res.Decl, res.Info) {
		ensureFuncHasCtxParam(res.Fset, res.FileAST, res.Decl)
		modifiedFiles[res.FileAST.Name.Name] = true // marker by pkg name; we'll use filenames later
		markFileModified(modifiedFiles, res.Fset, res.FileAST)
	}

	// Traverse callers recursively
	visited := map[types.Object]bool{}
	queue := []types.Object{res.Obj}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true

		// Scan all packages/files for calls to curr
		for _, p := range pkgs {
			for i, f := range p.Syntax {
				fi := p.Fset.File(f.Pos())
				if fi == nil {
					continue
				}
				ast.Inspect(f, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					var calledObj types.Object
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						calledObj = p.TypesInfo.Uses[fun]
					case *ast.SelectorExpr:
						calledObj = p.TypesInfo.Uses[fun.Sel]
					}
					if calledObj == nil || calledObj != curr {
						return true
					}
					// Found a call. Find enclosing function decl.
					enc := enclosingFuncDecl(f, call)
					if enc == nil {
						return true
					}

					stopHere, stopReason := shouldStopAt(enc, p, opts, stopSpec)
					if stopHere {
						// At stop boundary: ensure a ctx exists, derive if necessary (main/http) and pass to call
						ensured, err := ensureCtxAvailableAtBoundary(p, f, enc, stopReason)
						if err != nil {
							// continue but report later? Simpler: fail fast
							panic(fmt.Errorf("ensuring ctx at stop boundary: %w", err))
						}
						_ = ensured
						ensureCallHasCtxArg(p, call)
						markFileModified(modifiedFiles, p.Fset, p.Syntax[i])
						return true
					}

					// If ctx in scope, just pass; else add to enclosing func and enqueue it
					if hasCtxInScope(enc, p) {
						ensureCallHasCtxArg(p, call)
						markFileModified(modifiedFiles, p.Fset, p.Syntax[i])
						return true
					}

					// Add ctx param to enclosing function signature
					if !funcHasCtxParam(enc, p.TypesInfo) {
						ensureFuncHasCtxParam(p.Fset, f, enc)
					}
					ensureCallHasCtxArg(p, call)
					markFileModified(modifiedFiles, p.Fset, p.Syntax[i])

					// Enqueue enclosing function's object to continue traversal upward
					if def := p.TypesInfo.Defs[enc.Name]; def != nil {
						queue = append(queue, def)
					}
					return true
				})
			}
		}
	}

	// Write back modified files
	if err := writeModified(pkgs); err != nil {
		return fmt.Errorf("writing modified files: %w", err)
	}
	return nil
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

// writeModified writes all package files back using their fset and syntax.
func writeModified(pkgs []*packages.Package) error {
	for _, p := range pkgs {
		for _, f := range p.Syntax {
			filename := p.Fset.File(f.Pos()).Name()
			var buf strings.Builder
			cfg := &printer.Config{Mode: printer.TabIndent}
			if err := cfg.Fprint(&buf, p.Fset, f); err != nil {
				return fmt.Errorf("printing file %s: %w", filename, err)
			}
			// Best-effort formatting already done; write file
			if err := os.WriteFile(filename, []byte(buf.String()), 0o644); err != nil { //nolint:gosec // This is an appropriate permissions setting for source-code files.
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
