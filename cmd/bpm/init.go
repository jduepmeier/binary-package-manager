package main

import (
	"bpm"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type InitSubCommand struct {
	Opts InitSubCommandOpts
}
type InitSubCommandOpts struct{}

func init() {
	subCommands["init"] = &InitSubCommand{}
}

func (cmd *InitSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("init", "init the package manager", "init the package manager", &cmd.Opts)
	return err
}

func (cmd *InitSubCommand) Run(logger zerolog.Logger, manager *bpm.Manager) error {
	return manager.Init()
}
