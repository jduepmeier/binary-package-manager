package bpm

import "errors"

var (
	ErrPackageNotFound  = errors.New("package not found")
	ErrProviderNotFound = errors.New("package provider not found")
	ErrProviderConfig   = errors.New("provider config is not valid")
	ErrProvider         = errors.New("provider error")
	ErrProviderFetch    = errors.New("error fetching package")
)
