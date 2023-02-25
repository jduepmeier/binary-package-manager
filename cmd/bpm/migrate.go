package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type MigrateSubCommand struct {
	Opts MigrateSubCommandOpts
}
type MigrateSubCommandOpts struct{}

func init() {
	subCommands["migrate"] = &MigrateSubCommand{}
}

func (cmd *MigrateSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("migrate", "migrate packages", "migrate packages", &cmd.Opts)
	return err
}

func (cmd *MigrateSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Migrate()
}
