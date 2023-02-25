package main

import (
	"github.com/jduepmeier/binary-package-manager"
	"errors"
	"fmt"
	"io"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

var (
	build = "dev"
)

type opts struct {
	LogLevel string `short:"l" long:"loglevel" description:"loglevel to set"`
	Config   string `short:"c" long:"config" description:"path to config"`
	Quiet    bool   `short:"q" long:"quiet" description:"do not output on stdout"`
}

type SubCommand interface {
	AddCommand(parser *flags.Parser) error
	Run(logger zerolog.Logger, manager bpm.Manager) error
}

var (
	subCommands = map[string]SubCommand{}
)

const (
	EXIT_SUCCESS      = 0
	EXIT_ERROR        = 1
	EXIT_CONFIG_ERROR = 2
)

func run(managerCreateFunc bpm.ManagerCreateFunc, parserOut io.Writer, loggerOut io.Writer, args []string) int {
	opts := opts{
		LogLevel: "warn",
		Config:   "",
		Quiet:    false,
	}
	logger := zerolog.New(loggerOut).With().Timestamp().Logger()
	parser := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	for _, cmd := range subCommands {
		err := cmd.AddCommand(parser)
		if err != nil {
			logger.Err(err).Msg("")
			return EXIT_CONFIG_ERROR
		}
	}
	_, err := parser.ParseArgs(args)
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) {
			if flagsErr.Type == flags.ErrHelp {
				parser.WriteHelp(parserOut)
				return EXIT_SUCCESS
			}
		}
		fmt.Fprintf(parserOut, "%s\n", err)
		parser.WriteHelp(parserOut)
		return EXIT_CONFIG_ERROR
	}
	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil {
		logger.Err(err).Msg("")
		return EXIT_CONFIG_ERROR
	}
	logger = logger.Level(level)
	logger.Debug().Msg("starting up")

	cmd := subCommands[parser.Active.Name]
	migrate := false
	if parser.Active.Name == "migrate" {
		logger.Info().Msgf("migrate active")
		migrate = true
	}
	manager, err := managerCreateFunc(opts.Config, logger, migrate)
	if err != nil {
		logger.Err(err).Msg("cannot create manager instance")
		return EXIT_CONFIG_ERROR
	}
	manager.Config().Quiet = opts.Quiet

	logger.Debug().Msgf("execute command %s", parser.Active.Name)
	err = cmd.Run(logger, manager)
	if err != nil {
		logger.Err(err).Msg("")
		return EXIT_ERROR
	}

	err = manager.SaveState()
	if err != nil {
		logger.Err(err).Msg("cannot save state")
		return EXIT_ERROR
	}

	return EXIT_SUCCESS
}
