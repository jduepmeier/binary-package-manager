package main

import (
	"bpm"
	"testing"
)

func TestVersion(t *testing.T) {
	cmd := "version"
	tests := []testConfig{
		{
			name:     cmd,
			exitCode: EXIT_SUCCESS,
			args:     []string{cmd},
			testFunc: emptyTestFunc,
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig, &bpm.DummyManager{})
	}
}
