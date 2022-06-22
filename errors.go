package bpm

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrPackageNotFound           = errors.New("package not found")
	ErrPackageLoadError          = errors.New("cannot load package")
	ErrPackageNotInstalled       = errors.New("package is not installed")
	ErrPackageRemove             = errors.New("cannot remove package")
	ErrProviderNotFound          = errors.New("package provider not found")
	ErrProviderConfig            = errors.New("provider config is not valid")
	ErrProvider                  = errors.New("provider error")
	ErrProviderFetch             = errors.New("error fetching package")
	ErrMigrateNeeded             = fmt.Errorf("migration needed. Call `%s migrate` to migrate files", os.Args[0])
	ErrUnknownStateFileVersion   = errors.New("unknown state file version")
	ErrUnknownPackageFileVersion = errors.New("unknown package file version")
	ErrConfigLoad                = errors.New("cannot load config file")
	ErrYamlDump                  = errors.New("cannot dump content as yaml")
	ErrManagerCreate             = errors.New("cannot create new manager")
)
