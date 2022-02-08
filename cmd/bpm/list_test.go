package main

import (
	"bpm"
	"testing"
)

func TestList(t *testing.T) {
	tests := []testConfig{
		{
			name:     "list",
			exitCode: EXIT_SUCCESS,
			args:     []string{"list"},
			testFunc: emptyTestFunc,
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig, &bpm.DummyManager{})
	}
}
