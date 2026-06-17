package api

import (
	"context"
	"io"
	"time"
)

// Logger interface for logging operations
type Logger interface {
	Log(format string, args ...interface{})
}

// Config holds the configuration for must-gather collection
type Config struct {
	// DestDir is the root directory where all collected data will be stored
	DestDir string

	// LoggingNamespace is the namespace where cluster logging operator is deployed
	LoggingNamespace string

	// LogFileName is the name of the debug log file
	LogFileName string

	// Logger is where log output should be written
	Logger io.Writer
}

// Collector defines the interface for all must-gather collectors
type Collector interface {
	// Collect performs the collection and returns an error if collection fails
	Collect(ctx context.Context) error

	// Name returns the name of this collector
	Name() string
}

// Result represents the result of a collection operation
type Result struct {
	CollectorName string
	Error         error
	Duration      time.Duration
}
