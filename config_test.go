package bpm

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getTestPath returns the path from tests directory.
func getTestPath(subPaths ...string) string {
	base := []string{"tests"}
	return path.Join(append(base, subPaths...)...)
}

func getTestConfig(expand bool) *Config {
	stateFolder := "$HOME/.config/bpm"
	binFolder := "$HOME/bin"
	packagesFolder := ""
	if expand {
		stateFolder = os.ExpandEnv(stateFolder)
		binFolder = os.ExpandEnv(binFolder)
		packagesFolder = path.Join(stateFolder, "packages")
	}
	return &Config{
		BinFolder:      binFolder,
		StateFolder:    stateFolder,
		PackagesFolder: packagesFolder,
	}
}

func TestConfigLoad(t *testing.T) {

	tests := []struct {
		name               string
		filename           string
		result             *Config
		err                error
		errTestDescription string
	}{
		{
			name:               "missing",
			result:             getTestConfig(false),
			err:                ErrConfigLoad,
			errTestDescription: "missing file should return an error because the config file does not exist",
		},
		{
			name:               "broken",
			result:             getTestConfig(false),
			err:                ErrConfigLoad,
			errTestDescription: "broken file should return an error because the file is no valid yaml file",
		},
		{
			name:   "empty",
			result: getTestConfig(true),
		},
		{
			name: "replace-paths",
			result: func() *Config {
				config := getTestConfig(true)
				config.BinFolder = os.ExpandEnv("$HOME/bin")
				config.StateFolder = os.ExpandEnv("$HOME/state")
				config.PackagesFolder = os.ExpandEnv("$HOME/packages")
				return config
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			basePath := getTestPath("configs", fmt.Sprintf("%s.yaml", test.name))
			t.Logf("load config from path %s", basePath)
			config, err := ReadConfig(basePath)
			if test.err == nil {
				assert.NoError(t, err, test.errTestDescription)
			} else {
				assert.ErrorIs(t, err, test.err, test.errTestDescription)
			}
			assert.EqualValues(t, test.result, config, "Config should be equal.")
		})
	}
}

func TestDefaultConfigPaths(t *testing.T) {

	tests := []struct {
		name    string
		result  *Config
		message string
		paths   []string
	}{
		{
			name:   "empty",
			paths:  []string{},
			result: getTestConfig(true),
		},
		{
			name: "all-missing",
			paths: []string{
				"/tmp/missing-config-1",
				"/tmp/missing-config-2",
			},
			result: getTestConfig(true),
		},
		{
			name: "found-first",
			paths: []string{
				getTestPath("configs", "empty.yaml"),
				getTestPath("configs", "quiet.yaml"),
			},
			result: getTestConfig(true),
		},
		{
			name: "found-second",
			paths: []string{
				"/tmp/missing-config-1",
				getTestPath("configs", "quiet.yaml"),
			},
			result: func() *Config {
				config := getTestConfig(true)
				config.Quiet = true
				return config
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backupDefaultPaths := DefaultConfigPaths
			defer func() {
				DefaultConfigPaths = backupDefaultPaths
			}()
			DefaultConfigPaths = test.paths
			config, err := ReadConfig("")
			assert.NoError(t, err, "ReadConfig should not return an error when searching default configs")
			if test.message == "" {
				test.message = "read config does not match expected config"
			}
			assert.EqualValues(t, test.result, config, test.message)
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "tilde",
			input:  "~",
			output: os.ExpandEnv("$HOME"),
		},
		{
			name:   "env",
			input:  "$HOME/test",
			output: os.ExpandEnv("$HOME/test"),
		},
		{
			name:   "no-replace",
			input:  "~no-replace",
			output: "~no-replace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.output, expandPath(test.input))
		})
	}
}
