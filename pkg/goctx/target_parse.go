package goctx

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type targetSpec struct {
	File       string
	FuncName   string
	LineNumber int // Interpreted as 1-based line number when provided; 0 means unspecified
}

var targetRe = regexp.MustCompile(`^([^:]+):(\w+)(?::(\d+))?$`)

func parseTargetSpec(specStr string) (targetSpec, error) {
	slog.Debug("parseTargetSpec start", slog.String("spec", strings.TrimSpace(specStr)))
	specStr = filepath.ToSlash(strings.TrimSpace(specStr))
	matches := targetRe.FindStringSubmatch(specStr)
	if matches == nil {
		return targetSpec{}, errors.New("invalid target format, want path/to/file.go:Func[:N]")
	}
	ord := 0
	if matches[3] != "" {
		v, err := strconv.Atoi(matches[3])
		if err != nil || v <= 0 {
			return targetSpec{}, fmt.Errorf("invalid line number N in %q", specStr)
		}
		ord = v
	}

	spec := targetSpec{File: matches[1], FuncName: matches[2], LineNumber: ord}
	slog.Debug("parseTargetSpec done", slog.String("file", spec.File), slog.String("func", spec.FuncName), slog.Int("line", spec.LineNumber))

	return spec, nil
}
