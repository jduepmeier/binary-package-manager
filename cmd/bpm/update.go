package main

import (
	"bpm"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type UpdateSubCommand struct {
	Opts UpdateSubCommandOpts
}
type UpdateSubCommandOpts struct {
	Args struct {
		Packages []string
	} `positional-args:"true"`
}

func init() {
	subCommands["update"] = &UpdateSubCommand{}
}

func (cmd *UpdateSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("update", "updates packages", "updates packages. If no package is given update all", &cmd.Opts)
	return err
}

func (cmd *UpdateSubCommand) Run(logger zerolog.Logger, manager *bpm.Manager) error {
	return manager.Update(cmd.Opts.Args.Packages)
}
