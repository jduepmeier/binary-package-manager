package main

import (
	"bpm"
	"os"

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
	Run(logger zerolog.Logger, manager *bpm.Manager) error
}

var (
	subCommands = map[string]SubCommand{}
)

const (
	EXIT_SUCCESS      = 0
	EXIT_ERROR        = 1
	EXIT_CONFIG_ERROR = 2
)

func run() int {
	opts := opts{
		LogLevel: "warn",
		Config:   "",
		Quiet:    false,
	}
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	parser := flags.NewParser(&opts, flags.Default)
	for _, cmd := range subCommands {
		err := cmd.AddCommand(parser)
		if err != nil {
			logger.Err(err).Msg("")
			return EXIT_CONFIG_ERROR
		}
	}
	_, err := parser.Parse()
	if err != nil {
		if err == flags.ErrHelp {
			return EXIT_SUCCESS
		}
		return EXIT_CONFIG_ERROR
	}
	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil {
		logger.Err(err).Msg("")
		return EXIT_CONFIG_ERROR
	}
	logger = logger.Level(level)
	logger.Debug().Msg("starting up")

	manager, err := bpm.NewManager(opts.Config, logger)
	if err != nil {
		logger.Err(err).Msg("cannot create manager instance")
		return EXIT_CONFIG_ERROR
	}
	manager.Config.Quiet = opts.Quiet

	cmd := subCommands[parser.Active.Name]

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

func main() {
	os.Exit(run())
}
