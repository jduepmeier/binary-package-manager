package main

import (
	"github.com/jduepmeier/binary-package-manager"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyRemoveManager struct {
	*bpm.DummyManager
	removeTestFunc testRemoveFunc
}

type testRemoveFunc func(t *testing.T, name string)

func (manager *dummyRemoveManager) Remove(name string) error {
	if manager.removeTestFunc != nil {
		manager.removeTestFunc(manager.DummyManager.T, name)
	}
	return nil
}

type testRemoveConfig struct {
	testConfig     testConfig
	removeTestFunc testRemoveFunc
}

func TestRemove(t *testing.T) {
	tests := []testRemoveConfig{
		{
			testConfig: testConfig{
				name:     "empty",
				exitCode: EXIT_CONFIG_ERROR,
				args:     []string{"remove"},
				testFunc: testOutputContains("the required argument `Name` was not provided"),
			},
		},
		{
			testConfig: testConfig{
				name:     "success",
				exitCode: EXIT_SUCCESS,
				args:     []string{"remove", "testName"},
				testFunc: emptyTestFunc,
			},
			removeTestFunc: func(t *testing.T, name string) {
				assert.Equal(t, name, "testName")
			},
		},
	}
	for _, testConfig := range tests {
		testConfig.testConfig.manager = &dummyRemoveManager{
			DummyManager:   &bpm.DummyManager{},
			removeTestFunc: testConfig.removeTestFunc,
		}
		runTest(t, &testConfig.testConfig)
	}
}
