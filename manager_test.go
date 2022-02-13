package bpm

import (
	"bytes"
	"fmt"
	"io"
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
	configPath := path.Join(testDir, "config.yaml")
	err = dumpYaml(configPath, &config)
	if err != nil {
		t.Fatalf("could not create config file: %s", err)
	}

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
