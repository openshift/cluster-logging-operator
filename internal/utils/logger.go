package utils

import (
	stdlog "log"
	"os"
	"strconv"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/ViaQ/logerr/v2/log/static"
	"github.com/go-logr/logr"
)

const (
	envLogLevel            = "LOG_LEVEL"
	defaultLoggerVerbosity = 0
)

// InitLogger creates a logger and optionally initializes the verbosity with the value in LOG_LEVEL.
//
// It also replaces the static logger with the newly-created logger.
func InitLogger(component string) logr.Logger {
	verbosity := defaultLoggerVerbosity
	if rawVerbosity, ok := os.LookupEnv(envLogLevel); ok {
		envVerbosity, err := strconv.Atoi(rawVerbosity)
		if err != nil {
			stdlog.Panicf("%q must be an integer", envLogLevel)
		}

		verbosity = envVerbosity
	}

	logger := log.NewLogger(component, log.WithVerbosity(verbosity))
	static.SetLogger(logger)
	return logger
}
