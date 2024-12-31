package bpm

import (
	"fmt"
	"os"
	"runtime"

	"github.com/creasty/defaults"
	"github.com/rs/zerolog"
)

const (
	PackageSchemaVersion = 2
)

type PackageProvider interface {
	GetLatest(pkg Package) (version string, err error)
	FetchPackage(pkg Package, version string, cacheDir string) (path string, err error)
}

type Package struct {
	PackageV2 `yaml:",inline"`
}

type PackageV2 struct {
	SchemaVersion int               `yaml:"schema_version" default:"1"`
	Name          string            `yaml:"name"`
	Provider      string            `yaml:"provider"`
	URL           string            `yaml:"url"`
	GOOS          map[string]string `yaml:"goos"`
	GOARCH        map[string]string `yaml:"goarch"`
	AssetPattern  string            `yaml:"asset_pattern" default:"${goos}-${goarch}"`
	ArchiveFormat string            `yaml:"archive_format" default:""`
	BinPattern    string            `yaml:"bin_pattern" default:"${name}"`
	DownloadURL   string            `yaml:"download_url" default:""`
	TagFilter     string            `yaml:"tag_filter" default:""`
	PreReleases   bool              `yaml:"pre_releases"`
}

type PackageV1 struct {
	SchemaVersion int    `yaml:"schema_version" default:"1"`
	Name          string `yaml:"name"`
	Provider      string `yaml:"provider"`
	URL           string `yaml:"url"`
	GOOS          string `yaml:"goos"`
	GOARCH        string `yaml:"goarch"`
	AssetPattern  string `yaml:"asset_pattern" default:"${goos}-${goarch}"`
	ArchiveFormat string `yaml:"archive_format" default:""`
	BinPattern    string `yaml:"bin_pattern" default:"${name}"`
	DownloadURL   string `yaml:"download_url" default:""`
}

func (pkg *Package) SetDefaults() {
	pkg.GOOS = make(map[string]string)   // strings.ToLower(runtime.GOOS)
	pkg.GOARCH = make(map[string]string) // strings.ToLower(runtime.GOARCH)
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
		return fmt.Errorf("%w: %s", ErrPackageLoadError, err)
	}

	if obj.SchemaVersion != PackageSchemaVersion {
		return ErrMigrateNeeded
	}

	if err := unmarshal((*plain)(pkg)); err != nil {
		return fmt.Errorf("%w: %s", ErrPackageLoadError, err)
	}
	return nil
}

type StateFile struct {
	Version  int
	Packages map[string]string `yaml:"packages"`
}

type NewPackageProviderFunc = func(logger zerolog.Logger, config *Config) PackageProvider

var PackageProviders = make(map[string]NewPackageProviderFunc)

func (pkg *Package) patternExpand(pattern string, version string) string {
	mapper := func(placeHolderName string) string {
		switch placeHolderName {
		case "goos":
			goos, ok := pkg.GOOS[runtime.GOOS]
			if ok {
				return goos
			}
			return runtime.GOOS
		case "goarch":
			goarch, ok := pkg.GOARCH[runtime.GOARCH]
			if ok {
				return goarch
			}
			return runtime.GOARCH
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
