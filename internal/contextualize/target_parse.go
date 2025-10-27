package contextualize

import (
	"errors"
	"fmt"
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

	return targetSpec{File: matches[1], FuncName: matches[2], LineNumber: ord}, nil
}
