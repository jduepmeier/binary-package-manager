package bpm

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Manager struct {
	Config    *Config
	StateFile *StateFile
	Providers map[string]PackageProvider
	Packages  map[string]Package
	logger    zerolog.Logger
	tmpDir    string
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
			manager.logger.Info().Msgf("found package %s", pkg.Name)
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
	pkgName, ok := manager.Packages[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrPackageNotFound, name)
	}
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.Encode(pkgName)
	version, ok := manager.StateFile.Packages[name]
	if !ok {
		version = "not installed"
	}
	fmt.Printf("version: %s\n", version)
	return nil
}

func (manager *Manager) List() error {
	for name := range manager.Packages {
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

func (manager *Manager) Install(name string, force bool) (err error) {
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
	if version == currentVersion && !force {
		manager.logger.Info().Msgf("version is already installed :)")
		return nil
	}

	manager.tmpDir, err = ioutil.TempDir("", "bpm-*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(manager.tmpDir)
		manager.tmpDir = ""
	}()

	path, err := provider.FetchPackage(pkg, manager.tmpDir)
	if err != nil {
		return err
	}

	if pkg.ArchiveFormat != "" {
		path, err = manager.extractPackage(&pkg, path)
		if err != nil {
			return err
		}
	}

	err = manager.install(&pkg, version, path)
	if err != nil {
		return nil
	}

	manager.StateFile.Packages[name] = version
	return nil
}

func (manager *Manager) install(pkg *Package, version string, sourceFile string) error {
	targetFile := filepath.Join(manager.Config.BinFolder, pkg.Name)
	manager.logger.Debug().Msgf("install file %s to %s", sourceFile, targetFile)
	// first copy the new file to target file
	inputFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer func() {
		inputFile.Close()
	}()
	targetPathWithVersion := filepath.Join(manager.Config.BinFolder, fmt.Sprintf("%s-%s", pkg.Name, version))
	outputFile, err := os.Create(targetPathWithVersion)
	if err != nil {
		return err
	}
	_, err = io.Copy(outputFile, inputFile)
	outputFile.Close()
	if err != nil {
		os.Remove(targetPathWithVersion)
		return err
	}
	// make it executable
	err = os.Chmod(targetPathWithVersion, 0755)
	if err != nil {
		os.Remove(targetPathWithVersion)
		return err
	}

	// then we can rename the file
	err = os.Rename(targetPathWithVersion, targetFile)
	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) extractPackage(pkg *Package, sourceFile string) (string, error) {
	manager.logger.Info().Msgf("extract package %s (format %s)", pkg.Name, pkg.ArchiveFormat)
	switch pkg.ArchiveFormat {
	case "tar":
		return manager.extractTar(pkg, sourceFile)
	case "tar.gz":
		return manager.extractTarGZ(pkg, sourceFile)
	case "zip":
		return manager.extractZip(pkg, sourceFile)
	default:
		return "", fmt.Errorf("unknown archive format %s", pkg.ArchiveFormat)
	}
}

func (manager *Manager) extractTar(pkg *Package, sourcePath string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()
	return manager.extractTarReader(pkg, sourceFile)
}
func (manager *Manager) extractTarReader(pkg *Package, reader io.Reader) (string, error) {
	tarReader := tar.NewReader(reader)
	var outputPath string
	binPattern, err := regexp.Compile(pkg.patternExpand(pkg.BinPattern))
	if err != nil {
		return outputPath, err
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return outputPath, fmt.Errorf("archive does not contain a file matchin pattern %s", pkg.BinPattern)
		} else if err != nil {
			return outputPath, err
		}

		// search for regular files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		name := strings.ToLower(header.Name)
		if binPattern.Match([]byte(name)) {
			outputPath = filepath.Join(manager.tmpDir, fmt.Sprintf("output-%s", pkg.Name))
			file, err := os.Create(outputPath)
			if err != nil {
				return outputPath, err
			}
			defer file.Close()
			_, err = io.Copy(file, tarReader)
			return outputPath, err
		}
	}
}

func (manager *Manager) extractTarGZ(pkg *Package, sourceFile string) (string, error) {
	file, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}

	return manager.extractTarReader(pkg, reader)
}

func (manager *Manager) extractZip(pkg *Package, sourceFile string) (string, error) {
	var outputPath string
	binPattern, err := regexp.Compile(pkg.patternExpand(pkg.BinPattern))
	if err != nil {
		return outputPath, err
	}

	reader, err := zip.OpenReader(sourceFile)
	if err != nil {
		return outputPath, err
	}
	reader.Close()
	for _, file := range reader.File {
		name := strings.ToLower(file.Name)
		if file.FileInfo().Mode().IsRegular() && binPattern.Match([]byte(name)) {
			outputPath = filepath.Join(manager.tmpDir, fmt.Sprintf("output-%s", pkg.Name))
			outputFile, err := os.Create(outputPath)
			if err != nil {
				return outputPath, err
			}
			defer outputFile.Close()
			zipFile, err := file.Open()
			if err != nil {
				return outputPath, err
			}
			defer zipFile.Close()
			_, err = io.Copy(outputFile, zipFile)
			return outputPath, err
		}
	}

	return outputPath, fmt.Errorf("archive does not contain a file matching pattern %s", pkg.BinPattern)
}
