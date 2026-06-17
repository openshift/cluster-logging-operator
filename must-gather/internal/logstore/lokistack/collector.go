package lokistack

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects LokiStack resources
type Collector struct {
	client    *client.Client
	logger    api.Logger
	namespace string
}

// NewCollector creates a new LokiStack collector
func NewCollector(c *client.Client, logger api.Logger, namespace string) *Collector {
	return &Collector{
		client:    c,
		logger:    logger,
		namespace: namespace,
	}
}

// Name returns the name of this collector
func (l *Collector) Name() string {
	return "LogStoreCollector"
}

// Collect performs the collection of LokiStack resources
func (l *Collector) Collect(ctx context.Context, destDir, loggingNamespace string) error {
	l.logger.Log("BEGIN gather_logstore_resources ...")
	l.logger.Log("Gathering Lokistack resources")
	l.logger.Log("-- Gather Lokistack CR")

	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	// Write to namespace directory like other resources
	lokiFolder := filepath.Join(destDir, "namespaces", l.namespace, "loki.grafana.com", "lokistacks")
	if err := l.client.ListResources(ctx, lokiGVR, l.namespace, lokiFolder, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect LokiStack CR: %w", err)
	}

	l.logger.Log("END gather_logstore_resources ...")
	return nil
}
