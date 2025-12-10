package goctx

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"

	"github.com/yaklabco/stave/pkg/fsutils"
	"golang.org/x/tools/go/packages"
)

// targetResolution bundles the data about a resolved target function.
// It replaces the previous multiple return values of resolveTarget.
// Fields are intentionally exported within the package for clarity.
// Not exported outside the package as the type itself is unexported.
//
// Pkg: the package containing the file and function
// FileAST: the AST of the file containing the function
// Fset: the token.FileSet associated with the package
// Info: the types.Info for the package
// Decl: the function declaration found
// Obj: the types.Object corresponding to Decl.
type targetResolution struct {
	Pkg     *packages.Package
	FileAST *ast.File
	Fset    *token.FileSet
	Info    *types.Info
	Decl    *ast.FuncDecl
	Obj     types.Object
}

// resolveTarget locates the target function declaration and its types.Object.
// It now returns a targetResolution pointer and an error.
func resolveTarget(pkgs []*packages.Package, spec targetSpec) (*targetResolution, error) {
	slog.Debug("resolveTarget start", slog.String("file", spec.File), slog.String("func", spec.FuncName), slog.Int("line", spec.LineNumber))
	absFile, err := fsutils.TruePath(spec.File)
	if err != nil {
		return nil, fmt.Errorf("ascertaining true path: %w", err)
	}
	for _, pkg := range pkgs {
		for _, fileAST := range pkg.Syntax {
			posFile := pkg.Fset.File(fileAST.Pos())
			if posFile == nil {
				continue
			}
			fp, err := fsutils.TruePath(posFile.Name())
			if err != nil {
				return nil, fmt.Errorf("ascertaining true path: %w", err)
			}

			if fp != absFile {
				continue
			}
			candidates := findFuncDeclsByName(fileAST, spec.FuncName)
			if len(candidates) == 0 {
				return nil, fmt.Errorf("no function or method named %s in %s", spec.FuncName, spec.File)
			}
			// If a line number was provided, pick the function starting at that line
			if spec.LineNumber > 0 {
				var decl *ast.FuncDecl
				for _, c := range candidates {
					start := pkg.Fset.Position(c.Pos()).Line
					if start == spec.LineNumber {
						decl = c
						break
					}
				}
				if decl == nil {
					return nil, fmt.Errorf("no %s starting at line %d in %s", spec.FuncName, spec.LineNumber, spec.File)
				}
				obj := pkg.TypesInfo.Defs[decl.Name]
				if obj == nil {
					return nil, fmt.Errorf("resolving function object for %s", spec.FuncName)
				}
				tr := &targetResolution{Pkg: pkg, FileAST: fileAST, Fset: pkg.Fset, Info: pkg.TypesInfo, Decl: decl, Obj: obj}
				slog.Debug("resolveTarget found by line", slog.String("file", pkg.Fset.File(decl.Pos()).Name()), slog.String("func", decl.Name.Name), slog.Int("line", spec.LineNumber))

				return tr, nil
			}
			// No line provided
			if len(candidates) > 1 {
				return nil, fmt.Errorf(
					"ambiguous function %s in %s: found %d; please disambiguate using a line number as %s:%s:N (N is 1-based line)",
					spec.FuncName, spec.File, len(candidates), spec.File, spec.FuncName,
				)
			}
			decl := candidates[0]
			obj := pkg.TypesInfo.Defs[decl.Name]
			if obj == nil {
				return nil, fmt.Errorf("resolving function object for %s", spec.FuncName)
			}
			tr := &targetResolution{Pkg: pkg, FileAST: fileAST, Fset: pkg.Fset, Info: pkg.TypesInfo, Decl: decl, Obj: obj}
			slog.Debug("resolveTarget found", slog.String("file", pkg.Fset.File(decl.Pos()).Name()), slog.String("func", decl.Name.Name))

			return tr, nil
		}
	}

	return nil, fmt.Errorf("could not find file %s in loaded packages", spec.File)
}

func findFuncDeclsByName(file *ast.File, name string) []*ast.FuncDecl {
	var out []*ast.FuncDecl
	for _, d := range file.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Name.Name == name {
			out = append(out, fd)
		}
	}

	return out
}

func sameFile(p *packages.Package, fn *ast.FuncDecl, spec *targetSpec) (bool, error) {
	if spec == nil {
		return false, nil
	}
	fi := p.Fset.File(fn.Pos())
	if fi == nil {
		return false, nil
	}
	abs1, err := fsutils.TruePath(fi.Name())
	if err != nil {
		return false, fmt.Errorf("ascertaining true path: %w", err)
	}

	abs2, err := fsutils.TruePath(spec.File)
	if err != nil {
		return false, fmt.Errorf("ascertaining true path: %w", err)
	}

	return abs1 == abs2, nil
}

func enclosingFuncDecl(file *ast.File, targetNode ast.Node) *ast.FuncDecl {
	var stack []ast.Node
	var found *ast.FuncDecl
	ast.Inspect(file, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		if node == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return true
		}
		stack = append(stack, node)
		if node == targetNode {
			// walk back to find enclosing FuncDecl
			for i := len(stack) - 1; i >= 0; i-- {
				if fd, ok := stack[i].(*ast.FuncDecl); ok {
					found = fd
					break
				}
			}
			return false
		}
		return true
	})

	return found
}
