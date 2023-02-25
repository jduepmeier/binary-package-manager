package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type RemoveSubCommand struct {
	Opts RemoveSubCommandOpts
}
type RemoveSubCommandOpts struct {
	Args struct {
		Name string
	} `positional-args:"yes" required:"yes"`
}

func init() {
	subCommands["remove"] = &RemoveSubCommand{}
}

func (cmd *RemoveSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("remove", "remove a package", "remove a package", &cmd.Opts)
	return err
}

func (cmd *RemoveSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Remove(cmd.Opts.Args.Name)
}
