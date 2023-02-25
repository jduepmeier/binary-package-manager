package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type OutdatedSubCommand struct {
	Opts OutdatedSubCommandOpts
}
type OutdatedSubCommandOpts struct{}

func init() {
	subCommands["outdated"] = &OutdatedSubCommand{}
}

func (cmd *OutdatedSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("outdated", "outdated packages", "show packages with updates", &cmd.Opts)
	return err
}

func (cmd *OutdatedSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Outdated()
}
