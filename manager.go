package bpm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config    *Config
	StateFile *StateFile
	Providers map[string]PackageProvider
	Packages  map[string]Package
	logger    zerolog.Logger
}

func NewManager(configPath string, logger zerolog.Logger) (*Manager, error) {
	config, err := ReadConfig(configPath)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		Config:    config,
		Providers: make(map[string]PackageProvider),
		Packages:  make(map[string]Package),
		logger:    logger.With().Str("module", "manage").Logger(),
	}

	for name, providerFunc := range PackageProviders {
		manager.Providers[name] = providerFunc(manager.logger)
	}
	err = manager.Init()
	if err != nil {
		return manager, err
	}
	err = manager.LoadState()

	return manager, err
}

func (manager *Manager) Init() error {
	err := os.MkdirAll(manager.Config.StateFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(manager.Config.BinFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(manager.Config.PackagesFolder, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) SaveState() error {
	stateFile := filepath.Join(manager.Config.StateFolder, "state.yaml")
	return dumpYaml(stateFile, &manager.StateFile)
}

func (manager *Manager) LoadState() error {
	manager.StateFile = &StateFile{
		Version:  1,
		Packages: make(map[string]string),
	}

	err := loadYaml(filepath.Join(manager.Config.StateFolder, "state.yaml"), &manager.StateFile)

	// no state file exists.
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = filepath.Walk(manager.Config.PackagesFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".yaml") {
			pkg := Package{}
			err = loadYaml(path, &pkg)
			if err != nil {
				return err
			}
			if pkg.GOOS == "" {
				pkg.GOOS = runtime.GOOS
			}
			if pkg.GOARCH == "" {
				pkg.GOARCH = runtime.GOARCH
			}
			manager.Packages[pkg.Name] = pkg
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		manager.logger.Warn().Msgf("got error from loading packages: %s", err)
	}
	return err
}

func (manager *Manager) Info(name string) error {
	for _, pkg := range manager.StateFile.Packages {
		fmt.Printf("- %s\n", pkg)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrPackageNotFound, name)
}

func (manager *Manager) List() error {
	for name, _ := range manager.Packages {
		fmt.Printf("- %s\n", name)
	}
	return nil
}
func (manager *Manager) Installed() error {
	for name, version := range manager.StateFile.Packages {
		fmt.Printf("%s - %s\n", name, version)
	}
	return nil
}

func (manager *Manager) Add(name string, url string) error {
	splitted := strings.Split(url, "/")
	provider := splitted[0]
	pkg := Package{
		Name:     name,
		URL:      url,
		Provider: provider,
	}
	manager.Packages[name] = pkg
	return dumpYaml(filepath.Join(manager.Config.PackagesFolder, name+".yaml"), &pkg)
}

func (manager *Manager) Install(name string) (err error) {
	pkg, ok := manager.Packages[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrPackageNotFound, name)
	}
	provider, ok := manager.Providers[pkg.Provider]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, pkg.Provider)
	}
	currentVersion, ok := manager.StateFile.Packages[name]
	var version string
	if !ok || currentVersion == "" {
		version, err = provider.GetLatest(pkg)
		if err != nil {
			return err
		}
	}
	manager.logger.Info().Msgf("find package version %s", version)
	if version == currentVersion {
		manager.logger.Info().Msgf("version is already installed :)")
		return nil
	}

	cacheDir, err := ioutil.TempDir("", "bpm-*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(cacheDir)
	}()

	path, err := provider.FetchPackage(pkg, cacheDir)
	if err != nil {
		return err
	}

	err = manager.install(&pkg, path)
	if err != nil {
		return nil
	}

	manager.StateFile.Packages[name] = version
	return nil
}

func (manager *Manager) install(pkg *Package, sourceFile string) error {
	err := os.Rename(sourceFile, filepath.Join(manager.Config.BinFolder, pkg.Name))
	if err != nil {
		return err
	}
	return nil
}

func (manager *Manager) extractPackage(pkg *Package, sourceFile string) error {
	panic("not implemented")
}
