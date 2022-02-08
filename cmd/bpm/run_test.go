package main

import (
	"bpm"
	"bytes"
	"fmt"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type testFunc func(t *testing.T, manager bpm.TestManager, buf *bytes.Buffer) bool
type testConfig struct {
	name              string
	args              []string
	message           string
	exitCode          int
	testFunc          testFunc
	manager           bpm.TestManager
	managerCreateFunc bpm.ManagerCreateFunc
}

type brokenSubCommand struct {
	*InitSubCommand
}

func (cmd *brokenSubCommand) AddCommand(parser *flags.Parser) error {
	return fmt.Errorf("called broken subcommand")
}

type dummyRunErrorManager struct {
	*bpm.DummyManager
	failSaveState bool
	failRun       bool
}

func (manager *dummyRunErrorManager) Init() error {
	if manager.failRun {
		return fmt.Errorf("run error")
	}
	return nil
}

func (manager *dummyRunErrorManager) SaveState() error {
	if manager.failSaveState {
		return fmt.Errorf("fail save state error")
	}
	return nil
}

func emptyTestFunc(t *testing.T, manager bpm.TestManager, buf *bytes.Buffer) bool { return true }

func testOutputContains(contains string) testFunc {
	return func(t *testing.T, manager bpm.TestManager, buf *bytes.Buffer) bool {
		return assert.Contains(t, buf.String(), contains)
	}
}

func runTest(t *testing.T, testConfig *testConfig) {
	t.Run(testConfig.name, func(t *testing.T) {
		if testConfig.manager == nil {
			testConfig.manager = &bpm.DummyManager{}
		}
		if testConfig.managerCreateFunc == nil {
			testConfig.managerCreateFunc = func(configPath string, logger zerolog.Logger, migrate bool) (bpm.Manager, error) {
				return testConfig.manager, nil
			}
		}
		testConfig.manager.SetT(t)
		var buf bytes.Buffer
		exitCode := run(testConfig.managerCreateFunc, &buf, &buf, testConfig.args)
		if assert.Equal(t, testConfig.exitCode, exitCode, testConfig.message, &buf) {
			testConfig.testFunc(t, testConfig.manager, &buf)
		}
		t.Logf("output: %s", buf.String())
	})
}

func TestMain(t *testing.T) {
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
			name:     "run error",
			exitCode: EXIT_ERROR,
			message:  "exit error should be EXIT_ERROR because the run method failed",
			args:     []string{"init"},
			testFunc: emptyTestFunc,
			manager: &dummyRunErrorManager{
				DummyManager: &bpm.DummyManager{},
				failRun:      true,
			},
		},
		{
			name:     "fail save state",
			exitCode: EXIT_ERROR,
			message:  "exit error should be EXIT_ERROR because the SaveState method failed",
			args:     []string{"init"},
			testFunc: emptyTestFunc,
			manager: &dummyRunErrorManager{
				DummyManager:  &bpm.DummyManager{},
				failSaveState: true,
			},
		},
		{
			name:     "manager creation failed",
			exitCode: EXIT_CONFIG_ERROR,
			message:  "EXIT_CONFIG_ERROR should be returned because the manager cannot be created",
			args:     []string{"init"},
			testFunc: emptyTestFunc,
			managerCreateFunc: func(configPath string, logger zerolog.Logger, migrate bool) (bpm.Manager, error) {
				return nil, fmt.Errorf("manager creation failed")
			},
		},
		{
			name:     "init",
			exitCode: EXIT_SUCCESS,
			message:  "init is a valid command",
			args:     []string{"-l", "debug", "init"},
			testFunc: func(t *testing.T, manager bpm.TestManager, buf *bytes.Buffer) bool {
				realManager := manager.(*bpm.DummyManager)
				assert.Equal(t, 1, realManager.GetCounter("Init"), "manager.Init should be called one time")
				return assert.Equal(t, 1, realManager.GetCounter("SaveState"), "manager.SaveState should be called one time")
			},
		},
	}
	for _, testConfig := range tests {
		runTest(t, &testConfig)
	}

	subCommands["broken"] = &brokenSubCommand{}
	defer func() {
		delete(subCommands, "broken")
	}()

	testConfig := testConfig{
		name:     "broken command",
		exitCode: EXIT_CONFIG_ERROR,
		message:  "command broken should return an error when adding the command to the parser",
		args:     []string{"broken"},
		testFunc: emptyTestFunc,
	}
	runTest(t, &testConfig)
}
