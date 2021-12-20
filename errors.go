package bpm

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrPackageNotFound           = errors.New("package not found")
	ErrProviderNotFound          = errors.New("package provider not found")
	ErrProviderConfig            = errors.New("provider config is not valid")
	ErrProvider                  = errors.New("provider error")
	ErrProviderFetch             = errors.New("error fetching package")
	ErrMigrateNeeded             = fmt.Errorf("migration needed. Call `%s migrate` to migrate files", os.Args[0])
	ErrUnknownStateFileVersion   = errors.New("unknown state file version")
	ErrUnknownPackageFileVersion = errors.New("unknown state file version")
)
