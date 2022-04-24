package main

import (
	"testing"
)

func TestMigrate(t *testing.T) {
	tests := []testConfig{
		{
			name:     "migrate",
			exitCode: EXIT_SUCCESS,
			args:     []string{"migrate"},
			testFunc: emptyTestFunc,
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig)
	}
}
