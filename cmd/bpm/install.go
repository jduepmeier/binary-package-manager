package main

import (
	"github.com/jduepmeier/binary-package-manager"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type InstallSubCommand struct {
	Opts InstallSubCommandOpts
}
type InstallSubCommandOpts struct {
	Force bool `long:"force" short:"f" description:"force install"`
	Args  struct {
		Name string
	} `positional-args:"yes" required:"yes"`
}

func init() {
	subCommands["install"] = &InstallSubCommand{}
}

func (cmd *InstallSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("install", "installs a package", "installs a package", &cmd.Opts)
	return err
}

func (cmd *InstallSubCommand) Run(logger zerolog.Logger, manager bpm.Manager) error {
	return manager.Install(cmd.Opts.Args.Name, cmd.Opts.Force)
}
