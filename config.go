package bpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	DefaultConfigPaths = []string{
		"config.yaml",
		"~/.config/bpm/config.yaml",
	}
)

type Config struct {
	BinFolder      string       `yaml:"bin_folder"`
	StateFolder    string       `yaml:"state_folder"`
	PackagesFolder string       `yaml:"packages_folder"`
	Quiet          bool         `yaml:"quiet"`
	Github         GithubConfig `yaml:"github"`
}

func ReadConfig(path string) (*Config, error) {
	config := &Config{
		BinFolder:   "$HOME/bin",
		StateFolder: "$HOME/.config/bpm",
	}
	if path != "" {
		err := loadYaml(path, &config)
		if err != nil {
			return config, fmt.Errorf("cannot %s", err)
		}
	} else {
		for _, path := range DefaultConfigPaths {
			err := loadYaml(expandPath(path), &config)
			if err == nil {
				break
			}
		}
	}

	config.BinFolder = expandPath(config.BinFolder)
	config.StateFolder = expandPath(config.StateFolder)
	if config.PackagesFolder == "" {
		config.PackagesFolder = filepath.Join(config.StateFolder, "packages")
	}

	return config, nil
}

func loadYaml(path string, obj interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(obj)
}

func dumpYaml(path string, obj interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	return encoder.Encode(obj)
}

func expandPath(path string) string {
	if path == "~" {
		path = "$HOME"
	} else if strings.HasPrefix(path, "~/") {
		path = strings.Replace(path, "~/", "$HOME/", 1)
	}
	return os.ExpandEnv(path)
}
