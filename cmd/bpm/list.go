package main

import (
	"bpm"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type ListSubCommand struct {
	Opts ListSubCommandOpts
}
type ListSubCommandOpts struct {
}

func init() {
	subCommands["list"] = &ListSubCommand{}
}

func (cmd *ListSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("list", "list packages", "list packages", &cmd.Opts)
	return err
}

func (cmd *ListSubCommand) Run(logger zerolog.Logger, manager *bpm.Manager) error {
	return manager.List()
}
