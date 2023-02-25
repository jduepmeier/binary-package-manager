package main

import (
	"github.com/jduepmeier/binary-package-manager"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyUpdateManager struct {
	*bpm.DummyManager
	updateTestFunc testUpdateFunc
}

type testUpdateFunc func(t *testing.T, packages []string)

func (manager *dummyUpdateManager) Update(packages []string) error {
	if manager.updateTestFunc != nil {
		manager.updateTestFunc(manager.DummyManager.T, packages)
	}
	return nil
}

type testUpdateConfig struct {
	testConfig     testConfig
	updateTestFunc testUpdateFunc
}

func TestUpdate(t *testing.T) {
	cmd := "update"
	tests := []testUpdateConfig{
		{
			testConfig: testConfig{
				name:     "empty",
				exitCode: EXIT_SUCCESS,
				args:     []string{cmd},
				testFunc: emptyTestFunc,
			},
			updateTestFunc: func(t *testing.T, packages []string) {
				assert.Equal(t, len(packages), 0, "packages should be empty when given no package names")
			},
		},
		{
			testConfig: testConfig{
				name:     "success",
				exitCode: EXIT_SUCCESS,
				args:     []string{cmd, "testName"},
				testFunc: emptyTestFunc,
			},
			updateTestFunc: func(t *testing.T, packages []string) {
				assert.ElementsMatch(t, packages, []string{"testName"}, "packages should contain the names given on command line")
			},
		},
	}
	for _, testConfig := range tests {
		testConfig.testConfig.manager = &dummyUpdateManager{
			DummyManager:   &bpm.DummyManager{},
			updateTestFunc: testConfig.updateTestFunc,
		}
		runTest(t, &testConfig.testConfig)
	}
}
