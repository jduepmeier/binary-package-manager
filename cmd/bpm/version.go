package main

import (
	"bpm"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

type VersionSubCommand struct {
	Opts VersionSubCommandOpts
}
type VersionSubCommandOpts struct{}

func init() {
	subCommands["version"] = &VersionSubCommand{}
}

func (cmd *VersionSubCommand) AddCommand(parser *flags.Parser) error {
	_, err := parser.AddCommand("version", "show version of the package manager", "show version of the package manager", &cmd.Opts)
	return err
}

func (cmd *VersionSubCommand) Run(logger zerolog.Logger, manager *bpm.Manager) error {
	fmt.Printf("%s - %s\n", os.Args[0], build)
	return nil
}
