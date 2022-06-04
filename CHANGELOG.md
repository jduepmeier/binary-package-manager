# Changelog

## Unreleased

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
