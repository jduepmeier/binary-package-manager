//go:build !test
// +build !test

package main

import (
	"github.com/jduepmeier/binary-package-manager"
	"os"
)

func main() {
	os.Exit(run(bpm.NewManager, os.Stderr, os.Stderr, os.Args[1:]))
}
