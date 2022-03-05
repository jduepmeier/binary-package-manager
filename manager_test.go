package bpm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func getDummyLogger() zerolog.Logger {
	return zerolog.Nop()
}

func getDummyState() *StateFile {
	return &StateFile{
		Version:  StateFileVersion,
		Packages: make(map[string]string),
	}
}

func writeTestConfig(t *testing.T, config *Config) string {
	testDir := t.TempDir()
	configPath := path.Join(testDir, "config.yaml")
	err := dumpYaml(configPath, &config)
	if err != nil {
		t.Fatalf("could not create config file: %s", err)
	}
	return configPath
}
func writeTestState(t *testing.T, state *StateFile) string {
	testDir := t.TempDir()
	statePath := path.Join(testDir, "state.yaml")
	err := dumpYaml(statePath, &state)
	if err != nil {
		t.Fatalf("could not create state file: %s", err)
	}
	return statePath
}

func generateTestConfig(t *testing.T) (string, *Config, *StateFile) {
	testDir := t.TempDir()
	state := getDummyState()
	err := dumpYaml(path.Join(testDir, "state.yaml"), &state)
	if err != nil {
		t.Fatalf("could not create state file: %s", err)
	}
	config := &Config{
		StateFolder:    testDir,
		PackagesFolder: path.Join(testDir, "packages"),
		BinFolder:      path.Join(testDir, "bin"),
	}
	configPath := writeTestConfig(t, config)

	return configPath, config, state
}

func getDummyManagerImpl(t *testing.T) *ManagerImpl {
	manager := &ManagerImpl{
		config:    getTestTmpDirConfig(t),
		Providers: make(map[string]PackageProvider),
		Packages:  make(map[string]Package),
		logger:    zerolog.Nop(),
		stdout:    &bytes.Buffer{},
	}
	err := manager.Init()
	assert.NoError(t, err, "the manager should be initialized (all folders should be created)")
	return manager
}

type DummyProvider struct {
	// name: version
	LatestPackages map[string]string
	// name: path-to-file
	FetchPackages map[string]string
}

func (provider *DummyProvider) GetLatest(pkg Package) (version string, err error) {
	if provider.LatestPackages != nil {
		if version, ok := provider.LatestPackages[pkg.Name]; ok {
			return version, nil
		}
	}
	return "", ErrProviderFetch
}
func (provider *DummyProvider) FetchPackage(pkg Package, version string, cacheDir string) (outPath string, err error) {
	if provider.FetchPackages != nil {
		if inPath, ok := provider.FetchPackages[pkg.Name]; ok {
			inFile, err := os.Open(inPath)
			if err != nil {
				return "", err
			}
			defer inFile.Close()
			outPath := path.Join(cacheDir, path.Base(inPath))
			outFile, err := os.Create(outPath)
			if err != nil {
				return "", err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			return outPath, err
		}
	}
	return "", ErrProviderFetch
}

func TestNewManager(t *testing.T) {
	logger := getDummyLogger()
	configPath, config, state := generateTestConfig(t)
	t.Run("default", func(t *testing.T) {
		manager, err := NewManager(configPath, logger, false)
		if assert.NoError(t, err) {
			managerReal := manager.(*ManagerImpl)
			assert.EqualValues(t, manager.Config(), config)
			assert.EqualValues(t, state, managerReal.StateFile)
			assert.Contains(t, managerReal.Providers, "github.com")
		}
		assert.DirExists(t, config.BinFolder, "bin folder should exist")
		assert.DirExists(t, config.PackagesFolder, "package folder should exist")
	})

	t.Run("with-package", func(t *testing.T) {
		pkg := dummyPackage()
		dumpYaml(path.Join(config.PackagesFolder, "test.yaml"), &pkg)
		manager, err := NewManager(configPath, logger, false)
		if assert.NoError(t, err) {
			managerReal := manager.(*ManagerImpl)
			assert.Contains(t, managerReal.Packages, pkg.Name)
			assert.EqualValues(t, *pkg, managerReal.Packages[pkg.Name])
		}
	})

	t.Run("missing-config", func(t *testing.T) {
		configPath := "/tmp/missing-config"
		_, err := NewManager(configPath, logger, false)
		assert.ErrorIs(t, err, ErrManagerCreate)
	})

	t.Run("broken-init", func(t *testing.T) {
		tmpFile := path.Join(t.TempDir(), "file.yaml")
		err := os.WriteFile(tmpFile, []byte("tmpFile\n"), 0644)
		if err != nil {
			t.Fatalf("cannot write tmpFile %s: %s", tmpFile, err)
		}
		config := &Config{
			StateFolder: tmpFile,
		}
		configPath := writeTestConfig(t, config)
		_, err = NewManager(configPath, logger, false)
		assert.ErrorIs(t, err, ErrManagerCreate)
	})

	t.Run("broken-load-state", func(t *testing.T) {
		state := &StateFile{
			Version: 0,
		}
		statePath := writeTestState(t, state)
		config := getTestTmpDirConfig(t)
		config.StateFolder = path.Dir(statePath)
		configPath := writeTestConfig(t, config)
		_, err := NewManager(configPath, logger, false)
		assert.ErrorIs(t, err, ErrManagerCreate)
		_, err = NewManager(configPath, logger, true)
		assert.NoError(t, err, "in migration mode the LoadState function should not be called")
	})
}

func TestManagerInit(t *testing.T) {
	tmpFile := path.Join(t.TempDir(), "file.yaml")
	err := os.WriteFile(tmpFile, []byte("tmpFile\n"), 0644)
	if err != nil {
		t.Fatalf("cannot write tmpFile %s: %s", tmpFile, err)
	}
	tests := []struct {
		name  string
		setup func(manager *ManagerImpl)
	}{
		{
			name: "missing-state-folder",
			setup: func(manager *ManagerImpl) {
				manager.config.StateFolder = tmpFile
			},
		},
		{
			name: "missing-bin-folder",
			setup: func(manager *ManagerImpl) {
				manager.config.BinFolder = tmpFile
			},
		},
		{
			name: "missing-packages-folder",
			setup: func(manager *ManagerImpl) {
				manager.config.PackagesFolder = tmpFile
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := getDummyManagerImpl(t)
			test.setup(manager)
			err := manager.Init()
			assert.Error(t, err, "Init() must return an error because a path is a file")
		})
	}
}

func TestManagerSaveState(t *testing.T) {
	manager := getDummyManagerImpl(t)
	manager.StateFile = getDummyState()
	err := manager.SaveState()
	assert.NoError(t, err, "save state should work without problems")
	statePath := path.Join(manager.config.StateFolder, "state.yaml")
	assert.FileExists(t, statePath, "the state file should now exist")
	newState := StateFile{}
	err = loadYaml(statePath, &newState)
	assert.NoError(t, err, "state should be loaded without problems")
	assert.EqualValues(t, manager.StateFile, &newState, "loaded state must be the same as dumped state")
}

func TestManagerInfo(t *testing.T) {
	testPkg := *dummyPackage()
	var testPkgString bytes.Buffer
	encoder := yaml.NewEncoder(&testPkgString)
	err := encoder.Encode(&testPkg)
	if err != nil {
		t.Fatalf("cannot create dummy package output: %s", err)
	}
	tests := []outputTest{
		{
			name:        "missing",
			packageName: "missing",
			pkg:         &testPkg,
			state:       getDummyState(),
			output:      "",
			err:         ErrPackageNotFound,
		},
		{
			name:        "not-installed",
			pkg:         &testPkg,
			packageName: testPkg.Name,
			state:       getDummyState(),
			output:      fmt.Sprintf("%sversion: not installed\n", testPkgString.String()),
			err:         nil,
		},
		{
			name:        "installed",
			packageName: testPkg.Name,
			pkg:         &testPkg,
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[testPkg.Name] = "v1.0.0"
				return state
			}(),
			output: fmt.Sprintf("%sversion: v1.0.0\n", testPkgString.String()),
			err:    nil,
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		return manager.Info(test.packageName)
	})
}

func TestManagerList(t *testing.T) {

	tests := []outputTest{
		{
			name:     "no-packages",
			pkg:      nil,
			err:      nil,
			state:    getDummyState(),
			provider: &DummyProvider{},
			output:   "",
		},
		{
			name:   "one-package",
			output: fmt.Sprintf("- %s\n", dummyPackage().Name),
			state:  getDummyState(),
			err:    nil,
			pkg:    dummyPackage(),
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		return manager.List()
	})
}

func TestManagerInstalled(t *testing.T) {

	tests := []outputTest{
		{
			name:     "no-packages",
			pkg:      nil,
			err:      nil,
			state:    getDummyState(),
			provider: &DummyProvider{},
			output:   "",
		},
		{
			name:   "not-installed",
			output: "",
			state:  getDummyState(),
			err:    nil,
			pkg:    dummyPackage(),
		},
		{
			name:   "installed",
			output: fmt.Sprintf("%s - v1.0.0\n", dummyPackage().Name),
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			err: nil,
			pkg: dummyPackage(),
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		return manager.Installed()
	})
}

func setBoolPointer(b bool) *bool {
	return &b
}

func TestManagerInstall(t *testing.T) {

	tests := []outputTest{
		{
			name:        "no-packages",
			packageName: dummyPackage().Name,
			pkg:         nil,
			err:         ErrPackageNotFound,
			state:       getDummyState(),
			provider:    &DummyProvider{},
			output:      "",
		},
		{
			name:        "missing-packages",
			packageName: "missing",
			pkg:         nil,
			err:         ErrPackageNotFound,
			state:       getDummyState(),
			provider:    &DummyProvider{},
			output:      "",
		},
		{
			name:        "unknown-provider",
			packageName: dummyPackage().Name,
			pkg:         dummyPackage(),
			err:         ErrProviderNotFound,
			state:       getDummyState(),
			output:      "",
		},
		{
			name:        "cannot-fetch-latest-version",
			packageName: dummyPackage().Name,
			pkg:         dummyPackage(),
			err:         ErrProviderFetch,
			state:       getDummyState(),
			provider:    &DummyProvider{},
			output:      "",
		},
		{
			name:        "cannot-fetch-file",
			packageName: dummyPackage().Name,
			pkg:         dummyPackage(),
			err:         ErrProviderFetch,
			state:       getDummyState(),
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
			},
			output: "",
		},
		{
			name:        "not-installed",
			packageName: dummyPackage().Name,
			output:      "",
			state:       getDummyState(),
			err:         nil,
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
				FetchPackages: map[string]string{
					dummyPackage().Name: getTestPath("files", "dummy-bin.sh"),
				},
			},
			installed: setBoolPointer(true),
			pkg:       dummyPackage(),
		},
		{
			name:        "not-installed-tar",
			packageName: dummyPackage().Name,
			output:      "",
			state:       getDummyState(),
			err:         nil,
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
				FetchPackages: map[string]string{
					dummyPackage().Name: getTestPath("files", "dummy-bin.sh.tar"),
				},
			},
			installed: setBoolPointer(true),
			pkg: func() *Package {
				pkg := dummyPackage()
				pkg.ArchiveFormat = "tar"
				pkg.BinPattern = "dummy-bin.sh"
				return pkg
			}(),
		},
		{
			name:        "not-installed-tar-gz",
			packageName: dummyPackage().Name,
			output:      "",
			state:       getDummyState(),
			err:         nil,
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
				FetchPackages: map[string]string{
					dummyPackage().Name: getTestPath("files", "dummy-bin.sh.tar.gz"),
				},
			},
			installed: setBoolPointer(true),
			pkg: func() *Package {
				pkg := dummyPackage()
				pkg.ArchiveFormat = "tar.gz"
				pkg.BinPattern = "dummy-bin.sh"
				return pkg
			}(),
		},
		{
			name:        "not-installed-zip",
			packageName: dummyPackage().Name,
			output:      "",
			state:       getDummyState(),
			err:         nil,
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
				FetchPackages: map[string]string{
					dummyPackage().Name: getTestPath("files", "dummy-bin.sh.zip"),
				},
			},
			installed: setBoolPointer(true),
			pkg: func() *Package {
				pkg := dummyPackage()
				pkg.ArchiveFormat = "zip"
				pkg.BinPattern = "dummy-bin.sh"
				return pkg
			}(),
		},
		{
			name:        "installed",
			packageName: dummyPackage().Name,
			output:      "",
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: &DummyProvider{
				LatestPackages: map[string]string{
					dummyPackage().Name: "v1.0.0",
				},
			},
			err: nil,
			pkg: dummyPackage(),
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		err := manager.Install(test.packageName, false)
		if assert.ErrorIs(t, err, test.err) {
			if test.installed != nil {
				if *test.installed {
					assert.FileExists(t, path.Join(manager.config.BinFolder, test.packageName))
				} else {
					assert.NoFileExists(t, path.Join(manager.config.BinFolder, test.packageName))
				}
			}
		}
		return err
	})
}

func TestManagerAdd(t *testing.T) {
	manager := getDummyManagerImpl(t)
	pkgName := "testPkg"
	providerName := "dummy"
	testUrl := fmt.Sprintf("%s/test", providerName)
	err := manager.Add(pkgName, testUrl)
	assert.NoError(t, err, "add should not return an error")
	pkgPath := path.Join(manager.config.PackagesFolder, fmt.Sprintf("%s.yaml", pkgName))
	if assert.FileExists(t, pkgPath, "manager.Add should have created a package file") {
		pkg := Package{}
		content, err := ioutil.ReadFile(pkgPath)
		if err == nil {
			t.Logf("package content: %s", string(content))
		}
		err = loadYaml(pkgPath, &pkg)
		if assert.NoError(t, err, "the package should be valid yaml and can be loaded") {
			assert.Equal(t, pkgName, pkg.Name)
			assert.Equal(t, providerName, pkg.Provider)
		}
	}
}

func TestManagerOutdated(t *testing.T) {

	tests := []outputTest{
		{
			name:     "no-packages",
			pkg:      nil,
			err:      nil,
			state:    getDummyState(),
			provider: &DummyProvider{},
			output:   "",
		},
		{
			name:     "not-installed",
			output:   "",
			state:    getDummyState(),
			provider: &DummyProvider{},
			err:      nil,
			pkg:      dummyPackage(),
		},
		{
			name:   "installed-not-outdated",
			output: "",
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: func() PackageProvider {
				return &DummyProvider{
					LatestPackages: map[string]string{
						dummyPackage().Name: "v1.0.0",
					},
				}
			}(),
			err: nil,
			pkg: dummyPackage(),
		},
		{
			name:   "installed-outdated",
			output: fmt.Sprintf("%s: v1.0.0 => v1.1.0\n", dummyPackage().Name),
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: func() PackageProvider {
				return &DummyProvider{
					LatestPackages: map[string]string{
						dummyPackage().Name: "v1.1.0",
					},
				}
			}(),
			err: nil,
			pkg: dummyPackage(),
		},
		{
			name:   "installed-missing-provider",
			output: "",
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: nil,
			err:      ErrProviderNotFound,
			pkg:      dummyPackage(),
		},
		{
			name:   "installed-error-fetch-provider",
			output: "",
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: &DummyProvider{},
			err:      ErrProviderFetch,
			pkg:      dummyPackage(),
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		return manager.Outdated()
	})
}

func TestManagerUpdate(t *testing.T) {
	tests := []outputTest{
		{
			name:        "no-packages",
			packageName: dummyPackage().Name,
			pkg:         nil,
			err:         nil,
			state:       getDummyState(),
			provider:    &DummyProvider{},
			output:      "",
		},
		{
			name:        "specific-packages",
			packageName: dummyPackage().Name,
			pkg:         dummyPackage(),
			err:         nil,
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: &DummyProvider{
				FetchPackages: map[string]string{
					dummyPackage().Name: "v1.1.0",
				},
			},
			output: "",
		},
		{
			name:        "packages",
			packageName: "",
			pkg:         dummyPackage(),
			err:         nil,
			state: func() *StateFile {
				state := getDummyState()
				state.Packages[dummyPackage().Name] = "v1.0.0"
				return state
			}(),
			provider: &DummyProvider{
				FetchPackages: map[string]string{
					dummyPackage().Name: "v1.1.0",
				},
			},
			output: "",
		},
	}

	runOutputTests(t, tests, func(t *testing.T, test *outputTest, manager *ManagerImpl) error {
		if test.pkg != nil {
			manager.Packages = map[string]Package{
				test.pkg.Name: *test.pkg,
			}
		}
		packages := make([]string, 0)
		if test.packageName != "" {
			packages = append(packages, test.packageName)
		}
		err := manager.Update(packages)
		return err
	})
}

func TestInPackageList(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		pkgName  string
		result   bool
	}{
		{
			name:     "no-packages",
			packages: []string{},
			pkgName:  "",
			result:   false,
		},
		{
			name:     "no-packages-with-pkg-name",
			packages: []string{},
			pkgName:  "test",
			result:   false,
		},
		{
			name:     "packages-wrong-name",
			packages: []string{"pkg"},
			pkgName:  "test",
			result:   false,
		},
		{
			name:     "packages-correct",
			packages: []string{"pkg"},
			pkgName:  "pkg",
			result:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := getDummyManagerImpl(t)
			assert.Equal(t, test.result, manager.inPackageList(test.pkgName, test.packages))
		})
	}
}

type outputTest struct {
	name        string
	packageName string
	output      string
	state       *StateFile
	pkg         *Package
	provider    PackageProvider
	installed   *bool
	err         error
}

type outputTestFunc func(t *testing.T, test *outputTest, manager *ManagerImpl) error

func runOutputTests(t *testing.T, tests []outputTest, outputTestFunc outputTestFunc) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := getDummyManagerImpl(t)
			manager.StateFile = test.state
			if test.pkg != nil {
				manager.Packages[test.pkg.Name] = *test.pkg
			}
			if test.provider != nil {
				manager.Providers[dummyProviderName] = test.provider
			}
			var buf bytes.Buffer
			manager.stdout = &buf
			err := outputTestFunc(t, &test, manager)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.output, buf.String())
		})
	}
}
