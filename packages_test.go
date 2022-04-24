package bpm

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const dummyProviderName = "dummy"

func dummyPackage() *Package {
	return &Package{
		PackageV2{
			SchemaVersion: 2,
			Name:          "testName",
			GOOS:          make(map[string]string),
			GOARCH:        make(map[string]string),
			Provider:      dummyProviderName,
		},
	}
}

func TestPackagePatternExpand(t *testing.T) {
	pkg := dummyPackage()

	type patternExpandTest struct {
		input   string
		output  string
		version string
	}
	version := "1.0.0"

	tests := []patternExpandTest{
		{
			input:  "",
			output: "",
		},
		{
			input:  "name",
			output: "name",
		},
		{
			input:  "name-${name}",
			output: fmt.Sprintf("name-%s", pkg.Name),
		},
		{
			input:   "name-${name}-${version}",
			output:  fmt.Sprintf("name-%s-%s", pkg.Name, version),
			version: version,
		},
		{
			input:   "name-${goos}",
			output:  fmt.Sprintf("name-%s", runtime.GOOS),
			version: version,
		},
		{
			input:   "name-${goarch}",
			output:  fmt.Sprintf("name-%s", runtime.GOARCH),
			version: version,
		},
		{
			input:  "name-${missing}",
			output: "name-",
		},
	}

	for _, test := range tests {
		assert.Equal(t, pkg.patternExpand(test.input, test.version), test.output)
	}

	goarchOverride := "goarchOverride"
	goosOverride := "goosOverride"

	pkg.GOARCH[runtime.GOARCH] = goarchOverride
	pkg.GOOS[runtime.GOOS] = goosOverride

	tests = []patternExpandTest{
		{
			input:   "name-${goos}",
			output:  fmt.Sprintf("name-%s", goosOverride),
			version: version,
		},
		{
			input:   "name-${goarch}",
			output:  fmt.Sprintf("name-%s", goarchOverride),
			version: version,
		},
	}
	for _, test := range tests {
		assert.Equal(t, pkg.patternExpand(test.input, test.version), test.output)
	}
}

var (
	pkgTestStringVersion1 = `
---
schema_version: 1
name: testName
`
	pkgTestStringVersion2 = `
schema_version: 2
name: testName
`
	pkgTestStringMissingSchemaVersion = `
name: testName
`
	pkgTestStringSchemaFieldError = `
schema_version: string
`
	pkgTestStringFieldError = fmt.Sprintf(`
schema_version: %d
name: {}
`, PackageSchemaVersion)
)

func TestUnmarshalPackage(t *testing.T) {

	tests := []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "version-1",
			input: pkgTestStringVersion1,
			err:   ErrMigrateNeeded,
		},
		{
			name:  "version-2",
			input: pkgTestStringVersion2,
			err:   nil,
		},
		{
			name:  "missing-schema",
			input: pkgTestStringMissingSchemaVersion,
			err:   ErrMigrateNeeded,
		},
		{
			name:  "schema-version-error",
			input: pkgTestStringSchemaFieldError,
			err:   ErrPackageLoadError,
		},
		{
			name:  "schema-error",
			input: pkgTestStringFieldError,
			err:   ErrPackageLoadError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(test.input))
			decoder := yaml.NewDecoder(reader)
			pkg := Package{}
			err := decoder.Decode(&pkg)
			assert.ErrorIs(t, err, test.err)
		})
	}
}
