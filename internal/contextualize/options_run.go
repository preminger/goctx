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

	// Ensure target function has ctx param
	ensureTargetHasCtx(res, modifiedFiles)

	// Traverse callers recursively and propagate ctx as needed
	if err := traverseAndPropagate(pkgs, res.Obj, opts, stopSpec, modifiedFiles); err != nil {
		return err
	}

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
	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("loading packages: encountered errors")
	}
	return pkgs, nil
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
	if !funcHasCtxParam(res.Decl, res.Info) {
		ensureFuncHasCtxParam(res.Fset, res.FileAST, res.Decl)
		modifiedFiles[res.FileAST.Name.Name] = true // marker by pkg name; we'll use filenames later
		markFileModified(modifiedFiles, res.Fset, res.FileAST)
	}
}

// traverseAndPropagate walks callers recursively from start and ensures ctx propagation.
func traverseAndPropagate(pkgs []*packages.Package, start types.Object, opts Options, stopSpec *targetSpec, modifiedFiles map[string]bool) error {
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
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			calledObj = params.pkg.TypesInfo.Uses[fun]
		case *ast.SelectorExpr:
			calledObj = params.pkg.TypesInfo.Uses[fun.Sel]
		}
		if calledObj == nil || calledObj != params.curr {
			return true
		}
		// Found a call. Find enclosing function decl.
		enc := enclosingFuncDecl(params.fileAST, call)
		if enc == nil {
			return true
		}

		stopHere, stopReason := shouldStopAt(enc, params.pkg, params.opts, params.stopSpec)
		if stopHere {
			// At stop boundary: ensure a ctx exists, derive if necessary (main/http) and pass to call
			if _, err := ensureCtxAvailableAtBoundary(params.pkg, params.fileAST, enc, stopReason); err != nil {
				inspectErr = fmt.Errorf("ensuring ctx at stop boundary: %w", err)
				return false
			}
			ensureCallHasCtxArg(params.pkg, call)
			markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])
			return true
		}

		// If ctx in scope, just pass; else add to enclosing func and enqueue it
		if hasCtxInScope(enc, params.pkg) {
			ensureCallHasCtxArg(params.pkg, call)
			markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])
			return true
		}

		// Add ctx param to enclosing function signature if missing
		if !funcHasCtxParam(enc, params.pkg.TypesInfo) {
			ensureFuncHasCtxParam(params.pkg.Fset, params.fileAST, enc)
		}
		ensureCallHasCtxArg(params.pkg, call)
		markFileModified(params.modifiedFiles, params.pkg.Fset, params.pkg.Syntax[params.fileIndex])

		// Enqueue enclosing function's object to continue traversal upward
		if def := params.pkg.TypesInfo.Defs[enc.Name]; def != nil {
			*params.queue = append(*params.queue, def)
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
