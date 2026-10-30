package api

import (
	"context"
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Config holds the configuration for must-gather collection
type Config struct {
	// DestDir is the root directory where all collected data will be stored
	DestDir Path

	// LoggingNamespace is the namespace where cluster logging operator is deployed
	LoggingNamespace string

	// LogFileName is the name of the debug log file
	LogFileName string

	// Logger is where log output should be written
	Logger io.Writer

	// SinceSeconds restricts pod log collection to lines no older than this many seconds.
	// Corresponds to --since passed to oc adm must-gather.
	SinceSeconds *int64

	// SinceTime restricts pod log collection to lines after this absolute timestamp.
	// Corresponds to --since-time passed to oc adm must-gather.
	SinceTime *metav1.Time
}

// Collector defines the interface for all must-gather collectors
type Collector interface {
	// Collect performs the collection and returns an error if collection fails
	// Accepts optional GroupVersionResource parameters to collect specific resources
	Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error

	// Name returns the name of this collector
	Name() string
}

// Result represents the result of a collection operation
type Result struct {
	CollectorName string
	Error         error
	Duration      time.Duration
}
