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
	File     string
	FuncName string
	Ordinal  int // 1-based; 0 means unspecified
}

var targetRe = regexp.MustCompile(`^([^:]+):(\w+)(?::(\d+))?$`)

func parseTargetSpec(s string) (targetSpec, error) {
	s = filepath.ToSlash(strings.TrimSpace(s))
	m := targetRe.FindStringSubmatch(s)
	if m == nil {
		return targetSpec{}, errors.New("invalid target format, want path/to/file.go:Func[:N]")
	}
	ord := 0
	if m[3] != "" {
		v, err := strconv.Atoi(m[3])
		if err != nil || v <= 0 {
			return targetSpec{}, fmt.Errorf("invalid ordinal N in %q", s)
		}
		ord = v
	}
	return targetSpec{File: m[1], FuncName: m[2], Ordinal: ord}, nil
}
