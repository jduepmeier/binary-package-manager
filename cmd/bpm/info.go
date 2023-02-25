package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type InfoSubCommand struct {
	Opts InfoSubCommandOpts
}
type InfoSubCommandOpts struct {
	Args struct {
		Name string
	} `positional-args:"yes" required:"yes"`
}

func init() {
	subCommands["info"] = &InfoSubCommand{}
}

func (cmd *InfoSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("info", "show infos about a package", "shows info about a package", &cmd.Opts)
	return err
}

func (cmd *InfoSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Info(cmd.Opts.Args.Name)
}
