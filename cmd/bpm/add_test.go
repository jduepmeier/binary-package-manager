package main

import (
	"bpm"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyAddManager struct {
	*bpm.DummyManager
	addTestFunc testAddFunc
}

type testAddFunc func(t *testing.T, name string, url string)

func (manager *dummyAddManager) Add(name string, url string) error {
	if manager.addTestFunc != nil {
		manager.addTestFunc(manager.DummyManager.T, name, url)
	}
	return nil
}

type testAddConfig struct {
	testConfig  testConfig
	addTestFunc testAddFunc
}

func TestAdd(t *testing.T) {
	manager := &dummyAddManager{
		DummyManager: &bpm.DummyManager{},
	}
	tests := []testAddConfig{
		{
			testConfig: testConfig{
				name:     "empty",
				exitCode: EXIT_CONFIG_ERROR,
				args:     []string{"add"},
				testFunc: testOutputContains("the required arguments `Name` and `URL` were not provided"),
			},
		},
		{
			testConfig: testConfig{
				name:     "no url",
				exitCode: EXIT_CONFIG_ERROR,
				args:     []string{"add", "testName"},
				testFunc: testOutputContains("the required argument `URL` was not provided"),
			},
		},
		{
			testConfig: testConfig{
				name:     "success",
				exitCode: EXIT_SUCCESS,
				args:     []string{"add", "testName", "testURL"},
				testFunc: emptyTestFunc,
			},
			addTestFunc: func(t *testing.T, name, url string) {
				assert.Equal(t, name, "testName")
				assert.Equal(t, url, "testURL")
			},
		},
	}
	for _, testConfig := range tests {
		manager.addTestFunc = testConfig.addTestFunc
		runTest(t, &testConfig.testConfig, manager)
	}
}
