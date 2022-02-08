package main

import (
	"bpm"
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type testFunc func(t *testing.T, buf *bytes.Buffer) bool
type testConfig struct {
	name     string
	args     []string
	message  string
	exitCode int
	testFunc testFunc
}

func emptyTestFunc(t *testing.T, buf *bytes.Buffer) bool { return true }

func testOutputContains(contains string) testFunc {
	return func(t *testing.T, buf *bytes.Buffer) bool {
		return assert.Contains(t, buf.String(), contains)
	}
}

func runTest(t *testing.T, testConfig *testConfig, manager bpm.TestManager) {
	t.Run(testConfig.name, func(t *testing.T) {
		newManager := func(configPath string, logger zerolog.Logger, migrate bool) (bpm.Manager, error) {
			return manager, nil
		}
		manager.SetT(t)
		var buf bytes.Buffer
		exitCode := run(newManager, &buf, &buf, testConfig.args)
		if assert.Equal(t, testConfig.exitCode, exitCode, testConfig.message, &buf) {
			testConfig.testFunc(t, &buf)
		}
		t.Logf("output: %s", buf.String())
	})
}

func TestMain(t *testing.T) {
	manager := &bpm.DummyManager{}
	tests := []testConfig{
		{
			name:     "empty args",
			exitCode: EXIT_CONFIG_ERROR,
			message:  "empty args should fail with missing command",
			args:     []string{},
			testFunc: emptyTestFunc,
		},
		{
			name:     "wrong command",
			exitCode: EXIT_CONFIG_ERROR,
			message:  "command should not exist",
			args:     []string{"wrong-command"},
			testFunc: emptyTestFunc,
		},
		{
			name:     "--help",
			exitCode: EXIT_SUCCESS,
			message:  "--help should not fail",
			args:     []string{"--help"},
			testFunc: emptyTestFunc,
		},
		{
			name:     "unknown loglevel",
			exitCode: EXIT_CONFIG_ERROR,
			message:  "-l blubb is not a valid loglevel",
			args:     []string{"-l", "blubb", "version"},
			testFunc: emptyTestFunc,
		},
		{
			name:     "init",
			exitCode: EXIT_SUCCESS,
			message:  "init is a valid command",
			args:     []string{"-l", "debug", "init"},
			testFunc: func(t *testing.T, buf *bytes.Buffer) bool {
				return assert.Equal(t, 1, manager.GetCounter("Init"), "manager.Init should be called one time")
			},
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig, manager)
		manager.ResetCounters()
	}
}
