package goctx

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/preminger/goctx/pkg/util/fsutils"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func genGoldie(t *testing.T) *goldie.Goldie {
	t.Helper()

	return goldie.New(
		t,
		goldie.WithFixtureDir(fixturesDir(t)),
		goldie.WithNameSuffix(""),
	)
}

func fixturesDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "testdata", "golden", t.Name())
}

func normalizeNewlines(inBytes []byte) []byte {
	inStr := string(inBytes)
	inStr = strings.ReplaceAll(inStr, "\r\n", "\n")

	return []byte(inStr)
}

// inputDir returns the path to testdata/input alongside this test file.
func inputDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(file), "testdata", "input")
}

// writeTempModuleFromInput creates a temporary module for a test case.
// It copies input data to a temporary directory and writes a go.mod file.
func writeTempModuleFromInput(t *testing.T) string {
	t.Helper()

	src := filepath.Join(inputDir(t), t.Name())
	dst := t.TempDir()

	// Copy all files from src to dst, preserving relative paths
	require.NoError(t, copyDir(t, src, dst))

	// Always write a minimal go.mod to allow packages.Load to work in that dir
	gomod := filepath.Join(dst, "go.mod")
	require.NoError(t, os.WriteFile(gomod, []byte("module example.com/e2e\n\ngo 1.21\n"), 0o644))

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

func TestE2E_Simple_AddCtx_PreservesComments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)

	target := filepath.Join(dir, "main.go") + ":FuncInNeedOfContext"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	// Read back file and compare golden
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_Propagate_StopAtMain_PreservesComments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "a", "b.go") + ":Callee"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))

	bA := fsutils.MustRead(filepath.Join(dir, "a", "a.go"))
	bB := fsutils.MustRead(filepath.Join(dir, "a", "b.go"))
	bMain, err := os.ReadFile(filepath.Join(dir, "main.go"))
	require.NoError(t, err)

	g.Assert(t, "a.go", normalizeNewlines(bA))
	g.Assert(t, "b.go", normalizeNewlines(bB))
	g.Assert(t, "main.go", normalizeNewlines(bMain))
}

func TestE2E_Propagate_StopAtTestMain_PreservesComments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	// Start from callee and propagate upwards; should stop at TestMain and inject ctx := context.Background()
	target := filepath.Join(dir, "a", "b.go") + ":Callee"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	bA := fsutils.MustRead(filepath.Join(dir, "a", "a.go"))
	bB := fsutils.MustRead(filepath.Join(dir, "a", "b.go"))
	bMainTest := fsutils.MustRead(filepath.Join(dir, "main_test.go"))
	g.Assert(t, filepath.Join("a", "a.go"), normalizeNewlines(bA))
	g.Assert(t, filepath.Join("a", "b.go"), normalizeNewlines(bB))
	g.Assert(t, "main_test.go", normalizeNewlines(bMainTest))
}

func TestE2E_HTTPBoundary_PreservesComments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "srv", "srv.go") + ":inner"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir, HTML: true}))
	b := fsutils.MustRead(filepath.Join(dir, "srv", "srv.go"))
	g.Assert(t, "srv.go", normalizeNewlines(b))
}

func TestE2E_RenameBlankCtxParamToCtx(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "main.go") + ":target"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_UseExistingNamedContextParam(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "main.go") + ":target"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_UseExistingBlankContextParam(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "main.go") + ":target"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_ReuseExisting_Blank_Midlevel(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "main.go") + ":target"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_LineNumberDisambiguation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	// The (A) target starts at line 6 in the fixture
	target := filepath.Join(dir, "main.go") + ":target:6"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_BigExample(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)
	target := filepath.Join(dir, "main.go") + ":targetFunc"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}

func TestE2E_Methods_Propagation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	for iTest := range 2 {
		testLabel := strconv.Itoa(iTest + 1)
		t.Run(testLabel, func(t *testing.T) {
			t.Parallel()

			g := genGoldie(t)
			dir := writeTempModuleFromInput(t)
			target := filepath.Join(dir, "main.go") + ":target"
			require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
			b := fsutils.MustRead(filepath.Join(dir, "main.go"))
			g.Assert(t, "main.go", normalizeNewlines(b))
		})
	}
}

func TestE2E_ModuleWide_FromSubdir(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)

	// Run from the subdirectory where the target file resides, while callers exist outside of it.
	subdir := filepath.Join(dir, "sub")
	target := filepath.Join(subdir, "mylib.go") + ":FuncInNeedOfContext"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: subdir}))

	// Assert both the target file in subdir and the caller in parent (main.go) were updated.
	bSub := fsutils.MustRead(filepath.Join(subdir, "mylib.go"))
	bMain := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "mylib.go", normalizeNewlines(bSub))
	g.Assert(t, "main.go", normalizeNewlines(bMain))
}

func TestE2E_Handle_TestFiles_TargetIsTestAndCallsitesInTests(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)

	target := filepath.Join(dir, "helper_test.go") + ":HelperTarget"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	bHelper := fsutils.MustRead(filepath.Join(dir, "helper_test.go"))
	bCaller := fsutils.MustRead(filepath.Join(dir, "util_test.go"))
	g.Assert(t, "helper_test.go", normalizeNewlines(bHelper))
	g.Assert(t, "util_test.go", normalizeNewlines(bCaller))
}

func TestE2E_Handle_TestFiles_CallsitesInTestsForProdFunc(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)

	target := filepath.Join(dir, "util.go") + ":ProdFunc"
	require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
	bUtil := fsutils.MustRead(filepath.Join(dir, "util.go"))
	bTest := fsutils.MustRead(filepath.Join(dir, "util_test.go"))
	g.Assert(t, "util.go", normalizeNewlines(bUtil))
	g.Assert(t, "util_test.go", normalizeNewlines(bTest))
}

func TestE2E_Distinguish_Functions_Methods_And_Qualified(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	targetMap := map[string]func(string) string{
		"1": func(dir string) string { return filepath.Join(dir, "main.go") + ":MyFunc:11" },
		"2": func(dir string) string { return filepath.Join(dir, "main.go") + ":MyFunc:19" },
		"3": func(dir string) string { return filepath.Join(dir, "main.go") + ":MyFunc:25" },
		"4": func(dir string) string { return filepath.Join(dir, "xyz", "xyz.go") + ":MyFunc" },
	}

	for iTest := range 4 {
		testLabel := strconv.Itoa(iTest + 1)
		t.Run(testLabel, func(t *testing.T) {
			t.Parallel()

			g := genGoldie(t)
			dir := writeTempModuleFromInput(t)
			targetFunc, ok := targetMap[testLabel]
			require.True(t, ok)
			target := targetFunc(dir)
			require.NoError(t, Run(ctx, Options{Target: target, WorkDir: dir}))
			b := fsutils.MustRead(filepath.Join(dir, "main.go"))
			g.Assert(t, "main.go", normalizeNewlines(b))
		})
	}
}

func TestE2E_OnlyOneCtxParam_WhenCallingTwoCallees(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	g := genGoldie(t)
	dir := writeTempModuleFromInput(t)

	// Run twice to simulate two separate operations: first adding ctx to MyOtherFunc1 callers,
	// then to MyOtherFunc2 callers. MyFunc should only get one ctx parameter in total.
	target1 := filepath.Join(dir, "main.go") + ":MyOtherFunc1"
	require.NoError(t, Run(ctx, Options{Target: target1, WorkDir: dir}))
	target2 := filepath.Join(dir, "main.go") + ":MyOtherFunc2"
	require.NoError(t, Run(ctx, Options{Target: target2, WorkDir: dir}))
	b := fsutils.MustRead(filepath.Join(dir, "main.go"))
	g.Assert(t, "main.go", normalizeNewlines(b))
}
