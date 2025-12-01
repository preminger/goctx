package sub

import "fmt"

// FuncInNeedOfContext lives in a subdir package and is called from parent package main.
func FuncInNeedOfContext() {
	fmt.Println("sub")
}
