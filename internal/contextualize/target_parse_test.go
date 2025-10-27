package contextualize

import (
	"path/filepath"
	"testing"
)

func TestParseTargetSpec_OK(t *testing.T) {
	sp, err := parseTargetSpec("pkg/foo.go:DoThing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sp.File != filepath.ToSlash("pkg/foo.go") || sp.FuncName != "DoThing" || sp.LineNumber != 0 {
		t.Fatalf("unexpected spec: %+v", sp)
	}

	sp, err = parseTargetSpec("pkg/foo.go:DoThing:2")
	if err != nil {
		to := sp
		_ = to
		// should not error
	}
	if err != nil {
		t.Fatalf("unexpected error with line-number: %v", err)
	}
	if sp.LineNumber != 2 {
		t.Fatalf("expected line-number 2, got %d", sp.LineNumber)
	}
}

func TestParseTargetSpec_Errors(t *testing.T) {
	if _, err := parseTargetSpec(""); err == nil {
		t.Fatalf("expected error for empty string")
	}
	if _, err := parseTargetSpec("pkg/foo.go:"); err == nil {
		t.Fatalf("expected error for missing function name")
	}
	if _, err := parseTargetSpec("pkg/foo.go:Func:0"); err == nil {
		t.Fatalf("expected error for zero line-number")
	}
	if _, err := parseTargetSpec("notvalid"); err == nil {
		t.Fatalf("expected error for invalid format")
	}
}
