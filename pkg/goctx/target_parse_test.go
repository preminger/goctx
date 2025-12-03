package goctx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTargetSpec_OK(t *testing.T) {
	sp, err := parseTargetSpec("pkg/foo.go:FuncInNeedOfContext")
	require.NoError(t, err)

	assert.Equal(t, "pkg/foo.go", sp.File)
	assert.Equal(t, "FuncInNeedOfContext", sp.FuncName)
	assert.Equal(t, 0, sp.LineNumber)

	sp, err = parseTargetSpec("pkg/foo.go:FuncInNeedOfContext:2")
	require.NoError(t, err)
	assert.Equal(t, 2, sp.LineNumber)
}

func TestParseTargetSpec_Errors(t *testing.T) {
	var err error

	_, err = parseTargetSpec("")
	require.Error(t, err)

	_, err = parseTargetSpec("pkg/foo.go:")
	require.Error(t, err)

	_, err = parseTargetSpec("pkg/foo.go:Func:0")
	require.Error(t, err)

	_, err = parseTargetSpec("notvalid")
	require.Error(t, err)
}
