package main

import (
	"bpm"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyInfoManager struct {
	*bpm.DummyManager
	infoTestFunc testInfoFunc
}

type testInfoFunc func(t *testing.T, name string)

func (manager *dummyInfoManager) Info(name string) error {
	if manager.infoTestFunc != nil {
		manager.infoTestFunc(manager.DummyManager.T, name)
	}
	return nil
}

type testInfoConfig struct {
	testConfig   testConfig
	infoTestFunc testInfoFunc
}

func TestInfo(t *testing.T) {
	cmd := "info"
	tests := []testInfoConfig{
		{
			testConfig: testConfig{
				name:     "empty",
				exitCode: EXIT_CONFIG_ERROR,
				args:     []string{cmd},
				testFunc: testOutputContains("the required argument `Name` was not provided"),
			},
		},
		{
			testConfig: testConfig{
				name:     "success",
				exitCode: EXIT_SUCCESS,
				args:     []string{cmd, "testName"},
				testFunc: emptyTestFunc,
			},
			infoTestFunc: func(t *testing.T, name string) {
				assert.Equal(t, name, "testName")
			},
		},
	}
	for _, testConfig := range tests {
		testConfig.testConfig.manager = &dummyInfoManager{
			DummyManager: &bpm.DummyManager{},
			infoTestFunc: testConfig.infoTestFunc,
		}
		runTest(t, &testConfig.testConfig)
	}
}
