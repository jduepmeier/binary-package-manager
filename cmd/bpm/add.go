package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type AddSubCommand struct {
	Opts AddSubCommandOpts
}
type AddSubCommandOpts struct {
	Args struct {
		Name string
		URL  string
	} `positional-args:"yes" required:"yes"`
}

func init() {
	subCommands["add"] = &AddSubCommand{}
}

func (cmd *AddSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("add", "add a package", "add a package", &cmd.Opts)
	return err
}

func (cmd *AddSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Add(cmd.Opts.Args.Name, cmd.Opts.Args.URL)
}
