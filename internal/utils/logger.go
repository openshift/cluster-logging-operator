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
func InitLogger(component string) logr.Logger {
	verbosity := defaultLoggerVerbosity
	if rawVerbosity, ok := os.LookupEnv(envLogLevel); ok {
		envVerbosity, err := strconv.Atoi(rawVerbosity)
		if err != nil {
			stdlog.Panicf("%q must be an integer", envLogLevel)
		}

		verbosity = envVerbosity
	}
	return log.NewLogger(component, log.WithVerbosity(verbosity))
}

// InitStaticLogger creates a logger and optionally initializes the verbosity with the value in LOG_LEVEL
// and replaces the static logger with the newly-created logger.
func InitStaticLogger(component string) logr.Logger {
	logger := InitLogger(component)
	static.SetLogger(logger)
	return logger
}
