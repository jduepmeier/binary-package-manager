# binary-package-manager

This is a small package manager that fetches binaries from different sources (github, gitlab, ...)


## Install

To install the tool copy the binary from releases to a folder inside your `PATH` variable.
Alternative install it from source:

```bash
make

# installs inside $GOPATH/bin
make install
```

## Usage

First init bpm:

```bash
bpm init
```

To add packages create the corresponding yaml files inside `~/.config/bpm/packages`.
For kickstarting the file use:

```bash
bpm add <package> <repo-url>
```

Currently only github is supported.

See [package.example.yaml](package.example.yaml) for all available options.

Install a package with:

```bash
bpm install <package>
```

To update all packages use:

```bash
bpm update
```


## Release Notes

See [CHANGELOG.md](CHANGELOG.md).