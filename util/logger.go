package util

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger = log.Logger
	writer io.Writer
)

func init() {
	logLevel, err := zerolog.ParseLevel(os.Getenv("TUNGSTEN_LOG_LEVEL"))
	if err != nil {
		// If the log level is not set or invalid, default to InfoLevel
		logLevel = zerolog.DebugLevel
	}

	format := os.Getenv("TUNGSTEN_LOG_FORMAT")
	switch strings.ToLower(format) {
	case "json":
		writer = os.Stdout
	case "pretty":
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	default:
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	// Initialize the logger with default settings
	zerolog.SetGlobalLevel(logLevel)
	Logger = zerolog.New(writer).With().Timestamp().Logger().Level(logLevel);
	Logger.Debug().Msgf("Logger initialized with level %s ", logLevel.String())
}