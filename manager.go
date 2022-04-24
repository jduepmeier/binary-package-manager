package bpm

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

const (
	StateFileVersion = 1
)

type SchemaVersion struct {
	Version int
}

type ManagerCreateFunc func(configPath string, logger zerolog.Logger, migrate bool) (Manager, error)

type Manager interface {
	Config() *Config
	Init() error
	SaveState() error
	LoadState() error
	Info(name string) error
	List() error
	Installed() error
	Add(name string, url string) error
	Outdated() error
	Install(name string, force bool) error
	Update(packageNames []string) error
	Migrate() error
	FetchFromDownloadURL(pkg Package, version string, cacheDir string) (path string, err error)
}

type ManagerImpl struct {
	config    *Config
	StateFile *StateFile
	Providers map[string]PackageProvider
	Packages  map[string]Package
	logger    zerolog.Logger
	// Place to write stdout message to. Defaults to os.Stdout. Used for testing.
	stdout io.Writer
	tmpDir string
}

func NewManager(configPath string, logger zerolog.Logger, migrate bool) (Manager, error) {
	config, err := ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrManagerCreate, err)
	}

	manager := &ManagerImpl{
		config:    config,
		Providers: make(map[string]PackageProvider),
		Packages:  make(map[string]Package),
		logger:    logger.With().Str("module", "manage").Logger(),
		stdout:    os.Stdout,
	}

	for name, providerFunc := range PackageProviders {
		manager.Providers[name] = providerFunc(manager.logger, config)
	}
	err = manager.Init()
	if err != nil {
		return manager, fmt.Errorf("%w: %s", ErrManagerCreate, err)
	}

	if !migrate {
		err = manager.LoadState()
		if err != nil {
			err = fmt.Errorf("%w: %s", ErrManagerCreate, err)
		}
	}
	return manager, err
}

func (manager *ManagerImpl) Config() *Config {
	return manager.config
}

func (manager *ManagerImpl) Init() error {
	err := os.MkdirAll(manager.config.StateFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(manager.config.BinFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(manager.config.PackagesFolder, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (manager *ManagerImpl) SaveState() error {
	stateFile := filepath.Join(manager.config.StateFolder, "state.yaml")
	return dumpYaml(stateFile, &manager.StateFile)
}

func (manager *ManagerImpl) LoadState() error {
	manager.StateFile = &StateFile{
		Version:  1,
		Packages: make(map[string]string),
	}
	stateFilePath := filepath.Join(manager.config.StateFolder, "state.yaml")
	err := loadYaml(stateFilePath, &manager.StateFile)

	if manager.StateFile.Version != StateFileVersion {
		return fmt.Errorf("%w: %s", ErrMigrateNeeded, stateFilePath)
	}

	// no state file exists.
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = filepath.Walk(manager.config.PackagesFolder, func(path string, info os.FileInfo, err error) error {
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

func (manager *ManagerImpl) Info(name string) error {
	pkgName, ok := manager.Packages[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrPackageNotFound, name)
	}
	encoder := yaml.NewEncoder(manager.stdout)
	encoder.Encode(pkgName)
	version, ok := manager.StateFile.Packages[name]
	if !ok {
		version = "not installed"
	}
	fmt.Fprintf(manager.stdout, "version: %s\n", version)
	return nil
}

func (manager *ManagerImpl) List() error {
	for name := range manager.Packages {
		fmt.Fprintf(manager.stdout, "- %s\n", name)
	}
	return nil
}
func (manager *ManagerImpl) Installed() error {
	for name, version := range manager.StateFile.Packages {
		fmt.Fprintf(manager.stdout, "%s - %s\n", name, version)
	}
	return nil
}

func (manager *ManagerImpl) Add(name string, url string) error {
	splitted := strings.Split(url, "/")
	provider := splitted[0]
	pkg := Package{
		PackageV2: PackageV2{
			SchemaVersion: PackageSchemaVersion,
			Name:          name,
			URL:           url,
			Provider:      provider,
		},
	}
	manager.Packages[name] = pkg
	return dumpYaml(filepath.Join(manager.config.PackagesFolder, name+".yaml"), &pkg)
}

func (manager *ManagerImpl) Outdated() error {
	for _, pkg := range manager.Packages {
		logger := manager.logger.With().Str("pkg", pkg.Name).Logger()
		currentVersion, ok := manager.StateFile.Packages[pkg.Name]
		if !ok || currentVersion == "" {
			continue
		}
		provider, ok := manager.Providers[pkg.Provider]
		if !ok {
			return fmt.Errorf("%w: %s", ErrProviderNotFound, pkg.Provider)
		}
		version, err := provider.GetLatest(pkg)
		if err != nil {
			return err
		}
		if version == currentVersion {
			continue
		}
		logger.Info().Msgf("find package version %s", version)
		fmt.Fprintf(manager.stdout, "%s: %s => %s\n", pkg.Name, currentVersion, version)
	}
	return nil
}

func (manager *ManagerImpl) Install(name string, force bool) (err error) {
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
	if currentVersion != "" && !force {
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

	var path string
	if pkg.DownloadUrl != "" {
		path, err = manager.FetchFromDownloadURL(pkg, version, manager.tmpDir)
	} else {
		path, err = provider.FetchPackage(pkg, version, manager.tmpDir)
	}
	if err != nil {
		return err
	}

	if pkg.ArchiveFormat != "" {
		path, err = manager.extractPackage(&pkg, version, path)
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

func (manager *ManagerImpl) inPackageList(pkgName string, pkgNames []string) bool {
	for _, name := range pkgNames {
		if pkgName == name {
			return true
		}
	}
	return false
}

func (manager *ManagerImpl) Update(packageNames []string) (err error) {
	var selectedPackages []Package
	if len(packageNames) > 0 {
		for _, pkg := range manager.Packages {
			if manager.inPackageList(pkg.Name, packageNames) {
				selectedPackages = append(selectedPackages, pkg)
			}
		}
	} else {
		for _, pkg := range manager.Packages {
			selectedPackages = append(selectedPackages, pkg)
		}
	}
	for _, pkg := range selectedPackages {
		err := manager.update(&pkg)
		if err != nil {
			logger := manager.logger.With().Str("pkg", pkg.Name).Logger()
			logger.Error().Msgf("cannot update package: %s. Skipping...", err)
		}
	}
	return nil
}

func (manager *ManagerImpl) FetchFromDownloadURL(pkg Package, version string, cacheDir string) (path string, err error) {
	url := pkg.patternExpand(pkg.DownloadUrl, version)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return path, err
	}
	defer resp.Body.Close()
	var filename string
	if pkg.ArchiveFormat != "" {
		filename = fmt.Sprintf("%s.%s", pkg.Name, pkg.ArchiveFormat)
	} else {
		filename = pkg.Name
	}

	path = filepath.Join(cacheDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return path, err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return path, err
	}
	return path, nil
}

func (manager *ManagerImpl) update(pkg *Package) (err error) {
	logger := manager.logger.With().Str("pkg", pkg.Name).Logger()
	provider, ok := manager.Providers[pkg.Provider]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, pkg.Provider)
	}
	currentVersion, ok := manager.StateFile.Packages[pkg.Name]
	if !ok || currentVersion == "" {
		logger.Info().Msg("package is not installed")
		return nil
	}
	version, err := provider.GetLatest(*pkg)
	if err != nil {
		return err
	}
	logger.Info().Msgf("find package version %s", version)
	if version == currentVersion {
		logger.Info().Msgf("version is up to date :)")
		return nil
	}

	if !manager.config.Quiet {
		fmt.Printf("%s %s => %s\n", pkg.Name, currentVersion, version)
	}

	manager.tmpDir, err = ioutil.TempDir("", "bpm-*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(manager.tmpDir)
		manager.tmpDir = ""
	}()
	var path string
	if pkg.DownloadUrl != "" {
		path, err = manager.FetchFromDownloadURL(*pkg, version, manager.tmpDir)
	} else {
		path, err = provider.FetchPackage(*pkg, version, manager.tmpDir)
	}
	if err != nil {
		return err
	}

	if pkg.ArchiveFormat != "" {
		path, err = manager.extractPackage(pkg, version, path)
		if err != nil {
			return err
		}
	}

	err = manager.install(pkg, version, path)
	if err != nil {
		return nil
	}

	manager.StateFile.Packages[pkg.Name] = version
	return nil
}

func (manager *ManagerImpl) install(pkg *Package, version string, sourceFile string) error {
	targetFile := filepath.Join(manager.config.BinFolder, pkg.Name)
	manager.logger.Debug().Msgf("install file %s to %s", sourceFile, targetFile)
	// first copy the new file to target file
	inputFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer func() {
		inputFile.Close()
	}()
	targetPathWithVersion := filepath.Join(manager.config.BinFolder, fmt.Sprintf("%s-%s", pkg.Name, version))
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

func (manager *ManagerImpl) extractPackage(pkg *Package, version string, sourceFile string) (string, error) {
	manager.logger.Info().Msgf("extract package %s (format %s)", pkg.Name, pkg.ArchiveFormat)
	switch pkg.ArchiveFormat {
	case "tar":
		return manager.extractTar(pkg, version, sourceFile)
	case "tar.gz":
		return manager.extractTarGZ(pkg, version, sourceFile)
	case "zip":
		return manager.extractZip(pkg, version, sourceFile)
	default:
		return "", fmt.Errorf("unknown archive format %s", pkg.ArchiveFormat)
	}
}

func (manager *ManagerImpl) extractTar(pkg *Package, version string, sourcePath string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()
	return manager.extractTarReader(pkg, version, sourceFile)
}
func (manager *ManagerImpl) extractTarReader(pkg *Package, version string, reader io.Reader) (string, error) {
	tarReader := tar.NewReader(reader)
	var outputPath string
	binPattern, err := regexp.Compile(pkg.patternExpand(pkg.BinPattern, version))
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

func (manager *ManagerImpl) extractTarGZ(pkg *Package, version string, sourceFile string) (string, error) {
	file, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}

	return manager.extractTarReader(pkg, version, reader)
}

func (manager *ManagerImpl) extractZip(pkg *Package, version string, sourceFile string) (string, error) {
	var outputPath string
	binPattern, err := regexp.Compile(pkg.patternExpand(pkg.BinPattern, version))
	if err != nil {
		return outputPath, err
	}

	reader, err := zip.OpenReader(sourceFile)
	if err != nil {
		return outputPath, err
	}
	defer reader.Close()
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

func (manager *ManagerImpl) migrateStateFile() error {
	manager.logger.Debug().Msgf("migrate state file")
	version := SchemaVersion{
		Version: 1,
	}
	stateFilePath := filepath.Join(manager.config.StateFolder, "state.yaml")
	err := loadYaml(stateFilePath, &version)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if version.Version == StateFileVersion {
		manager.logger.Debug().Msgf("no state file migration needed")
		return nil
	}

	return fmt.Errorf("%w: %d", ErrUnknownStateFileVersion, version.Version)
}

func (manager *ManagerImpl) migratePackageFile(path string) (err error) {
	manager.logger.Debug().Msgf("migrate package file %s", path)

	version := SchemaVersion{
		Version: 1,
	}
	err = loadYaml(path, &version)
	if os.IsNotExist(err) {
		manager.logger.Debug().Msgf("file %s does not exist, no migration needed", path)
		return nil
	} else if err != nil {
		return err
	}

	if version.Version == PackageSchemaVersion {
		manager.logger.Debug().Msgf("no package file migration needed for %s", path)
		return nil
	}
	var pkg Package

	switch version.Version {
	case 1:
		manager.logger.Debug().Msgf("migrate 1 to %d", PackageSchemaVersion)
		pkgV1 := PackageV1{}
		err = loadYaml(path, &pkgV1)
		if err != nil {
			return err
		}
		pkg = Package{
			PackageV2: PackageV2{
				SchemaVersion: 2,
				Name:          pkgV1.Name,
				Provider:      pkgV1.Provider,
				URL:           pkgV1.URL,
				GOOS:          make(map[string]string),
				GOARCH:        make(map[string]string),
				AssetPattern:  pkgV1.AssetPattern,
				ArchiveFormat: pkgV1.ArchiveFormat,
				BinPattern:    pkgV1.BinPattern,
				DownloadUrl:   pkgV1.DownloadUrl,
			},
		}
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		pkg.GOOS[goos] = pkgV1.GOOS
		pkg.GOARCH[goarch] = pkgV1.GOARCH
	default:
		return fmt.Errorf("%w: %d", ErrUnknownPackageFileVersion, version.Version)
	}

	return dumpYaml(path, &pkg)
}

func (manager *ManagerImpl) migratePackageFiles() (err error) {
	err = filepath.Walk(manager.config.PackagesFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".yaml") {
			err = manager.migratePackageFile(path)
		}
		return err
	})
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		manager.logger.Warn().Msgf("got error from loading packages: %s", err)
	}
	return err
}

func (manager *ManagerImpl) Migrate() error {
	err := manager.migrateStateFile()
	if err != nil {
		return err
	}
	err = manager.migratePackageFiles()
	if err != nil {
		return err
	}
	return manager.LoadState()
}
