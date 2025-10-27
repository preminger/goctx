package contextualize

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

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
	absFile, err := filepath.Abs(spec.File)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}
	for _, p := range pkgs {
		for _, f := range p.Syntax {
			posFile := p.Fset.File(f.Pos())
			if posFile == nil {
				continue
			}
			fp, err := filepath.Abs(posFile.Name())
			if err != nil {
				return nil, fmt.Errorf("resolving absolute path: %w", err)
			}

			if fp != absFile {
				continue
			}
			candidates := findFuncDeclsByName(f, spec.FuncName)
			if len(candidates) == 0 {
				return nil, fmt.Errorf("no function named %s in %s", spec.FuncName, spec.File)
			}
			if len(candidates) > 1 {
				if spec.Ordinal == 0 {
					return nil, fmt.Errorf("ambiguous function %s in %s: found %d; please specify as %s:%s:N", spec.FuncName, spec.File, len(candidates), spec.File, spec.FuncName)
				}
				if spec.Ordinal < 1 || spec.Ordinal > len(candidates) {
					return nil, fmt.Errorf("invalid ordinal N=%d for %s in %s (have %d)", spec.Ordinal, spec.FuncName, spec.File, len(candidates))
				}
			}
			idx := spec.Ordinal
			if idx == 0 {
				idx = 1
			}
			decl := candidates[idx-1]
			obj := p.TypesInfo.Defs[decl.Name]
			if obj == nil {
				return nil, fmt.Errorf("resolving function object for %s", spec.FuncName)
			}
			return &targetResolution{Pkg: p, FileAST: f, Fset: p.Fset, Info: p.TypesInfo, Decl: decl, Obj: obj}, nil
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

func sameFile(p *packages.Package, fn *ast.FuncDecl, spec *targetSpec) bool {
	if spec == nil {
		return false
	}
	fi := p.Fset.File(fn.Pos())
	if fi == nil {
		return false
	}
	abs1, err := filepath.Abs(fi.Name())
	if err != nil {
		return false
	}

	abs2, err := filepath.Abs(spec.File)
	if err != nil {
		return false
	}

	return abs1 == abs2
}

func ordinalOfFuncInFile(p *packages.Package, fn *ast.FuncDecl, spec *targetSpec) int {
	fi := p.Fset.File(fn.Pos())
	if fi == nil {
		return 0
	}
	var idx int
	ast.Inspect(getFileForPos(p, fn), func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fd.Name.Name == spec.FuncName {
			idx++
			if fd == fn {
				return false
			}
		}
		return true
	})
	return idx
}

func getFileForPos(p *packages.Package, node ast.Node) *ast.File {
	for _, f := range p.Syntax {
		if f.Pos() <= node.Pos() && node.Pos() <= f.End() {
			return f
		}
	}
	return nil
}

func enclosingFuncDecl(file *ast.File, n ast.Node) *ast.FuncDecl {
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
		if node == n {
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
