package bpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BinFolder      string `yaml:"bin_folder"`
	StateFolder    string `yaml:"state_folder"`
	PackagesFolder string `yaml:"packages_folder"`
}

func ReadConfig(path string) (*Config, error) {
	config := &Config{
		BinFolder:   "$HOME/bin",
		StateFolder: "$HOME/.config/bpm",
	}

	err := loadYaml(path, &config)
	if err != nil {
		return config, fmt.Errorf("cannot read %s", err)
	}

	config.BinFolder = expandPath(config.BinFolder)
	config.StateFolder = expandPath(config.StateFolder)
	if config.PackagesFolder == "" {
		config.PackagesFolder = filepath.Join(config.StateFolder, "packages")
	}

	return config, err
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
