# Changelog

## 0.4.3 (2024-01-07)

### Fix

- **deps**: update module github.com/google/go-github/v56 to v57

## 0.4.2 (2023-11-10)

### Fix

- **deps**: update module github.com/google/go-github/v55 to v56

## 0.4.1 (2023-11-10)

### Fix

- **deps**: update module github.com/google/go-github/v55 to v56
- **deps**: update module github.com/rs/zerolog to v1.31.0
- **deps**: update module github.com/google/go-github/v54 to v55

## 0.4.0 (2023-08-27)

### Feat

- **manager**: add tar.xz support

## 0.3.3 (2023-08-27)

### Fix

- **deps**: update module github.com/google/go-github/v52 to v54
- **deps**: update module github.com/rs/zerolog to v1.30.0
- **deps**: update module github.com/stretchr/testify to v1.8.4
- **deps**: update module github.com/google/go-github/v53 to v53.1.0
- **deps**: update module github.com/google/go-github/v52 to v53
- **deps**: update module github.com/stretchr/testify to v1.8.3

## 0.3.2 (2023-05-18)

### Fix

- **deps**: update module github.com/google/go-github/v51 to v52
- **deps**: update module github.com/google/go-github/v50 to v51
- **deps**: update module github.com/masterminds/semver/v3 to v3.2.1
- **deps**: update module github.com/rs/zerolog to v1.29.1
- **deps**: update module github.com/google/go-github/v50 to v50.2.0
- **deps**: update module github.com/creasty/defaults to v1.7.0

## 0.3.1 (2023-02-25)

### Fix

- **deps**: update module github.com/stretchr/testify to v1.8.2
- **deps**: update module github.com/google/go-github/v50 to v50.1.0

## 0.3.0 (2023-02-23)

* build(goreleaser): fix deprecations
* fix(deps): update module gopkg.in/yaml.v2 to v3
* Bump golang.org/x/net from 0.0.0-20211112202133-69e39bad7dc2 to 0.7.0
* fix(deps): update module github.com/google/go-github/v36 to v50
* fix(deps): update module gopkg.in/yaml.v2 to v3
* chore(deps): update github/codeql-action action to v2
* chore(deps): update actions/setup-go action to v3
* Add remove command
* fix(deps): use yaml.v3 instead of yaml.v2 in code
* chore(deps): update actions/checkout action to v3
* fix(deps): update module github.com/stretchr/testify to v1.8.1
* fix(deps): update module github.com/masterminds/semver/v3 to v3.2.0
* fix(deps): update module github.com/rs/zerolog to v1.29.0
* chore(deps): add renovate.json

## 0.2.0 (2022-06-04)

* Update dependencies
* Update README
* Fix update does always update same package
* Fix error message for unknown package file version
* Check for error cases in cmd run func
* Add tests
* Add codeql github workflow
* Create github workflow
* Add debugging info to rate-limits
* Allow to pass github token
* allow to update only named packages
* add outdated command
* add migrate command
* convert goos and goarch to maps to use same package definition on different platforms. This needs schema migration (package schema is now 2).

## 0.1.1 (2021-11-18)

* match asset pattern case sensitive

## 0.1.0 (2021-11-18)

* add version command
* add a download_url parameter and add version to patternExpand
* show updated packages in update command
* use default config if config is missing and search in /home/mpease
* check schema versions
* add update command to update all packages
* implement install with extracting archives
* initial version
