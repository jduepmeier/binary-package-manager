package bpm

import (
	"testing"

	"github.com/rs/zerolog"
)

type DummyManager struct {
	config   *Config
	T        *testing.T
	counters map[string]int
}

type TestManager interface {
	Manager
	SetT(t *testing.T)
}

func NewDummyManager(configPath string, logger zerolog.Logger, migrate bool) (Manager, error) {
	return &DummyManager{}, nil
}

func (manager *DummyManager) SetT(t *testing.T) {
	manager.T = t
}

func (manager *DummyManager) bumpCounter(name string) {
	if manager.counters == nil {
		manager.counters = make(map[string]int)
	}

	manager.counters[name]++
}

func (manager *DummyManager) GetCounter(name string) int {
	if manager.counters == nil {
		return 0
	}
	counter, ok := manager.counters[name]
	if ok {
		return counter
	}
	return 0
}

func (manager *DummyManager) ResetCounters() {
	manager.counters = nil
}

func (manager *DummyManager) Config() *Config {
	if manager.config == nil {
		manager.config = &Config{}
	}
	return manager.config
}

func (manager *DummyManager) Init() error {
	manager.bumpCounter("Init")
	return nil
}

func (manager *DummyManager) SaveState() error {
	manager.bumpCounter("SaveState")
	return nil
}

func (manager *DummyManager) LoadState() error {
	manager.bumpCounter("LoadState")
	return nil
}

func (manager *DummyManager) Info(name string) error {
	manager.bumpCounter("Info")
	return nil
}

func (manager *DummyManager) Remove(name string) error {
	manager.bumpCounter("Remove")
	return nil
}

func (manager *DummyManager) List() error {
	manager.bumpCounter("List")
	return nil
}

func (manager *DummyManager) Installed() error {
	manager.bumpCounter("Installed")
	return nil
}

func (manager *DummyManager) Add(name string, url string) error {
	manager.bumpCounter("Add")
	return nil
}

func (manager *DummyManager) Outdated() error {
	manager.bumpCounter("Outdated")
	return nil
}

func (manager *DummyManager) Install(name string, force bool) error {
	manager.bumpCounter("Install")
	return nil
}

func (manager *DummyManager) Update(packageNames []string) error {
	manager.bumpCounter("Update")
	return nil
}

func (manager *DummyManager) Migrate() error {
	manager.bumpCounter("Migrate")
	return nil
}

func (manager *DummyManager) FetchFromDownloadURL(pkg Package, version string, cacheDir string) (path string, err error) {
	manager.bumpCounter("FetchFromDownloadURL")
	return "", nil
}
