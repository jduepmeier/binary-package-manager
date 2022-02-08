package bpm

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dummyPackage() *Package {
	return &Package{
		PackageV2{
			SchemaVersion: 1,
			Name:          "testName",
			GOOS:          make(map[string]string),
			GOARCH:        make(map[string]string),
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
