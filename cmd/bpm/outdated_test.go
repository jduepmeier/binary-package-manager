package main

import (
	"testing"
)

func TestOutdated(t *testing.T) {
	cmd := "outdated"
	tests := []testConfig{
		{
			name:     cmd,
			exitCode: EXIT_SUCCESS,
			args:     []string{cmd},
			testFunc: emptyTestFunc,
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig)
	}
}
