package main

import (
	"bpm"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type UpdateSubCommand struct {
	Opts UpdateSubCommandOpts
}
type UpdateSubCommandOpts struct{}

func init() {
	subCommands["update"] = &UpdateSubCommand{}
}

func (cmd *UpdateSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("update", "updates all package", "updates all packages", &cmd.Opts)
	return err
}

func (cmd *UpdateSubCommand) Run(logger zerolog.Logger, manager *bpm.Manager) error {
	return manager.Update()
}
