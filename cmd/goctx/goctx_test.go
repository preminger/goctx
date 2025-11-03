package goctx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoArgsPrintsHelp(t *testing.T) {
	ctx := t.Context()
	cmd := NewRootCmd(ctx)
	cmd.SetArgs([]string{})
	var stdoutBuf strings.Builder
	var stderrBuf strings.Builder
	cmd.SetOut(&stdoutBuf)
	cmd.SetErr(&stderrBuf)

	require.NoError(t, ExecuteWithFang(ctx, cmd))

	require.Regexp(t, `\bUSAGE\b`, stdoutBuf.String())
	require.Regexp(t, `\bEXAMPLES\b`, stdoutBuf.String())
	require.Regexp(t, `\bFLAGS\b`, stdoutBuf.String())
	require.Empty(t, stderrBuf.String())
}
