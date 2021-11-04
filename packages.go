package bpm

import (
	"os"
	"runtime"
	"strings"

	"github.com/creasty/defaults"
	"github.com/rs/zerolog"
)

const (
	PackageSchemaVersion = 1
)

type PackageProvider interface {
	GetLatest(pkg Package) (version string, err error)
	FetchPackage(pkg Package, version string, cacheDir string) (path string, err error)
}

type Package struct {
	SchemaVersion int    `yaml:"schema_version" default:"1"`
	Name          string `yaml:"name"`
	Provider      string `yaml:"provider"`
	URL           string `yaml:"url"`
	GOOS          string `yaml:"goos"`
	GOARCH        string `yaml:"goarch"`
	AssetPattern  string `yaml:"asset_pattern" default:"${goos}-${goarch}"`
	ArchiveFormat string `yaml:"archive_format" default:""`
	BinPattern    string `yaml:"bin_pattern" default:"${name}"`
	DownloadUrl   string `yaml:"download_url" default:""`
}

func (pkg *Package) SetDefaults() {
	pkg.GOOS = strings.ToLower(runtime.GOOS)
	pkg.GOARCH = strings.ToLower(runtime.GOARCH)
}

func (pkg *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(pkg)

	type plain Package

	type T struct {
		SchemaVersion int `yaml:"schema_version"`
	}
	obj := &T{
		SchemaVersion: 1,
	}

	if err := unmarshal((*T)(obj)); err != nil {
		return err
	}

	if obj.SchemaVersion != PackageSchemaVersion {
		return ErrMigrateNeeded
	}

	if err := unmarshal((*plain)(pkg)); err != nil {
		return err
	}
	return nil
}

type StateFile struct {
	Version  int
	Packages map[string]string `yaml:"packages"`
}

type NewPackageProviderFunc = func(logger zerolog.Logger) PackageProvider

var (
	PackageProviders = make(map[string]NewPackageProviderFunc)
)

func (pkg *Package) patternExpand(pattern string, version string) string {
	mapper := func(placeHolderName string) string {
		switch placeHolderName {
		case "goos":
			return pkg.GOOS
		case "goarch":
			return pkg.GOARCH
		case "name":
			return pkg.Name
		case "version":
			return version
		default:
			return ""
		}
	}
	return os.Expand(pattern, mapper)
}
