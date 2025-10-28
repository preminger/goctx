package contextualize

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func fixturesDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "testdata", "golden")
}

func normalizeNewlines(s string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(s, "\r\n", "\n")
	}

	return s
}

func TestE2E_Simple_AddCtx_PreservesComments(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_simple_addctx_preserve_comments")

	target := filepath.Join(dir, "main.go") + ":DoThing"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// Read back file and compare golden
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	g.Assert(t, "e2e_simple_addctx_preserve_comments", []byte(normalizeNewlines(string(b))))
}

func TestE2E_Propagate_StopAtMain_PreservesComments(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_propagate")
	target := filepath.Join(dir, "a", "b.go") + ":Callee"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	bA, err := os.ReadFile(filepath.Join(dir, "a", "a.go"))
	if err != nil {
		t.Fatalf("read a.go: %v", err)
	}
	bB, err := os.ReadFile(filepath.Join(dir, "a", "b.go"))
	if err != nil {
		t.Fatalf("read b.go: %v", err)
	}
	bMain, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_propagate_a_go", []byte(normalizeNewlines(string(bA))))
	g.Assert(t, "e2e_propagate_b_go", []byte(normalizeNewlines(string(bB))))
	g.Assert(t, "e2e_propagate_main_go", []byte(normalizeNewlines(string(bMain))))
}

func TestE2E_HTTPBoundary_PreservesComments(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_http_boundary")
	target := filepath.Join(dir, "srv", "srv.go") + ":inner"
	if err := Run(ctx, Options{Target: target, WorkDir: dir, HTML: true}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "srv", "srv.go"))
	if err != nil {
		t.Fatalf("read srv.go: %v", err)
	}
	g.Assert(t, "e2e_http_boundary_srv_go", []byte(normalizeNewlines(string(b))))
}

func TestE2E_RenameBlankCtxParamToCtx(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_ctxparam_blank_to_ctx")
	target := filepath.Join(dir, "main.go") + ":target"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_ctxparam_blank_to_ctx_main_go", []byte(normalizeNewlines(string(b))))
}

func TestE2E_UseExistingNamedContextParam(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_ctxparam_named_use_existing")
	target := filepath.Join(dir, "main.go") + ":target"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_ctxparam_named_use_existing_main_go", []byte(normalizeNewlines(string(b))))
}

func TestE2E_UseExistingBlankContextParam(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_ctxparam_blank_use_existing")
	target := filepath.Join(dir, "main.go") + ":target"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_ctxparam_blank_use_existing_main_go", []byte(normalizeNewlines(string(b))))
}

func TestE2E_ReuseExisting_Blank_Midlevel(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_reuse_midlevel_blank")
	target := filepath.Join(dir, "main.go") + ":target"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_reuse_midlevel_blank_main_go", []byte(normalizeNewlines(string(b))))
}

func TestE2E_LineNumberDisambiguation(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_line_number_disambiguation")
	// The (A) target starts at line 6 in the fixture
	target := filepath.Join(dir, "main.go") + ":target:6"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_line_number_disambiguation_main_go", []byte(normalizeNewlines(string(b))))
}

// inputDir returns the path to testdata/input alongside this test file.
func inputDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "testdata", "input")
}

// writeTempModuleFromInput copies a fixture case from testdata/input/<caseName>
// into a new temporary directory and writes a minimal go.mod. It returns the temp dir.
func writeTempModuleFromInput(t *testing.T, caseName string) string {
	t.Helper()
	src := filepath.Join(inputDir(t), caseName)
	dst := t.TempDir()
	// Copy all files from src to dst, preserving relative paths
	if err := copyDir(t, src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	// Always write a minimal go.mod to allow packages.Load to work in that dir
	gomod := filepath.Join(dst, "go.mod")
	if err := os.WriteFile(gomod, []byte("module example.com/e2e\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	return dst
}

// copyDir recursively copies files from src to dst.
func copyDir(t *testing.T, src, dst string) error {
	t.Helper()
	return filepath.WalkDir(src, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if dirEntry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		// file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func TestE2E_BigExample(t *testing.T) {
	ctx := t.Context()
	g := goldie.New(t, goldie.WithFixtureDir(fixturesDir(t)))
	dir := writeTempModuleFromInput(t, "e2e_big_example")
	target := filepath.Join(dir, "main.go") + ":target"
	if err := Run(ctx, Options{Target: target, WorkDir: dir}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	g.Assert(t, "e2e_big_example_main_go", []byte(normalizeNewlines(string(b))))
}
