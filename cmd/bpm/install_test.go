package main

import (
	"github.com/jduepmeier/binary-package-manager"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyInstallManager struct {
	*bpm.DummyManager
	installTestFunc testInstallFunc
}

type testInstallFunc func(t *testing.T, name string, force bool)

func (manager *dummyInstallManager) Install(name string, force bool) error {
	if manager.installTestFunc != nil {
		manager.installTestFunc(manager.DummyManager.T, name, force)
	}
	return nil
}

type testInstallConfig struct {
	testConfig      testConfig
	installTestFunc testInstallFunc
}

func TestInstall(t *testing.T) {
	cmd := "install"
	tests := []testInstallConfig{
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
			installTestFunc: func(t *testing.T, name string, force bool) {
				assert.Equal(t, name, "testName")
				assert.False(t, force, "force should be false on default")
			},
		},
		{
			testConfig: testConfig{
				name:     "success with force",
				exitCode: EXIT_SUCCESS,
				args:     []string{cmd, "testName", "--force"},
				testFunc: emptyTestFunc,
			},
			installTestFunc: func(t *testing.T, name string, force bool) {
				assert.Equal(t, name, "testName")
				assert.True(t, force, "force should be true with --force")
			},
		},
	}
	for _, testConfig := range tests {
		testConfig.testConfig.manager = &dummyInstallManager{
			DummyManager:    &bpm.DummyManager{},
			installTestFunc: testConfig.installTestFunc,
		}
		runTest(t, &testConfig.testConfig)
	}
}
