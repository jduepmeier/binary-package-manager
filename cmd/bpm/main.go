//go:build !test
// +build !test

package main

import (
	"bpm"
	"os"
)

func main() {
	os.Exit(run(bpm.NewManager, os.Stderr, os.Stderr, os.Args[1:]))
}
