package mustgather

import (
	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"context"
	"fmt"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// LogStoreCollector collects log store resources (LokiStack)
type LogStoreCollector struct {
	client        *client.Client
	logger        *api.Logger
	logStoreType  string // "lokistack"
	namespace     string
}

// NewLogStoreCollector creates a new log store resource collector
func NewLogStoreCollector(c *client.Client, logger *api.Logger, logStoreType, namespace string) *LogStoreCollector {
	return &LogStoreCollector{
		client:       c,
		logger:       logger,
		logStoreType: logStoreType,
		namespace:    namespace,
	}
}

// Name returns the name of this collector
func (l *LogStoreCollector) Name() string {
	return "LogStoreCollector"
}

// Collect performs the collection of log store resources
func (l *LogStoreCollector) Collect(ctx context.Context, config *api.Config) error {
	l.logger.Log("BEGIN gather_logstore_resources ...")
	l.logger.Log("Gathering data for logstore component")

	if l.logStoreType == "lokistack" {
		if err := l.collectLokiStack(ctx, config); err != nil {
			l.logger.Log("WARNING: Failed to collect LokiStack resources: %v", err)
		}
	}

	l.logger.Log("END gather_logstore_resources ...")
	return nil
}

// collectLokiStack collects LokiStack resources
func (l *LogStoreCollector) collectLokiStack(ctx context.Context, config *api.Config) error {
	l.logger.Log("Gathering Lokistack resources")
	l.logger.Log("-- Gather Lokistack CR")

	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	lokiFolder := filepath.Join(config.DestDir, "cluster-logging", "lokistack")
	if err := l.client.ListResources(ctx, lokiGVR, l.namespace, lokiFolder, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect LokiStack CR: %w", err)
	}

	return nil
}
