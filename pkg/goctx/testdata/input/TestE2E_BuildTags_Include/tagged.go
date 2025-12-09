//go:build mytag

package main

// FuncInNeedOfContext is present only when built with the 'mytag' build tag.
// It intentionally has no context parameter so goctx has work to do.
func FuncInNeedOfContext() {
	// no-op
}
