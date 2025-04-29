package util

import (
	"github.com/henrikvtcodes/tungsten/config"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	LogLevel zerolog.Level
	Logger   = log.Logger
	writer   io.Writer
)

func init() {
	// https://github.com/rs/zerolog#leveled-logging
	logLevel, err := zerolog.ParseLevel(os.Getenv("TUNGSTEN_LOG_LEVEL"))
	if err != nil {
		// If the log level is not set or invalid, default to InfoLevel
		logLevel = zerolog.WarnLevel
	}
	LogLevel = logLevel

	format := os.Getenv("TUNGSTEN_LOG_FORMAT")
	switch strings.ToLower(format) {
	case "json":
		writer = os.Stdout
	case "pretty":
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	default:
		writer = os.Stdout
	}

	// Initialize the logger with default settings
	zerolog.SetGlobalLevel(logLevel)
	loggerCtx := zerolog.New(writer).With().Timestamp()
	if config.DevMode {
		loggerCtx = loggerCtx.Caller()
	}
	Logger = loggerCtx.Logger().Level(logLevel)
	Logger.Debug().Msgf("Logger initialized with level %s ", logLevel.String())
}
