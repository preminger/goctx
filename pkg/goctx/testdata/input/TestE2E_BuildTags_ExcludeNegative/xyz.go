//go:build xyz

package main

// FuncOnlyWhenXYZ exists only when the 'xyz' tag is enabled.
func FuncOnlyWhenXYZ() {}
