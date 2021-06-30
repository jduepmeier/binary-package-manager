package bpm

import "github.com/rs/zerolog"

type PackageProvider interface {
	GetLatest(pkg Package) (version string, err error)
	FetchPackage(pkg Package, cacheDir string) (path string, err error)
}

type Package struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	URL      string `yaml:"url"`
	GOOS     string `yaml:"goos"`
	GOARCH   string `yaml:"goarch"`
}

type StateFile struct {
	Version  int
	Packages map[string]string `yaml:"packages"`
}

type NewPackageProviderFunc = func(logger zerolog.Logger) PackageProvider

var (
	PackageProviders = make(map[string]NewPackageProviderFunc)
)
