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
	destDir   string
}

// NewCollector creates a new LokiStack collector
func NewCollector(c *client.Client, logger api.Logger, namespace, destDir string) *Collector {
	return &Collector{
		client:    c,
		logger:    logger,
		namespace: namespace,
		destDir:   destDir,
	}
}

// Name returns the name of this collector
func (l *Collector) Name() string {
	return "LogStoreCollector"
}

// Collect performs the collection of LokiStack resources
func (l *Collector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	l.logger.Log("BEGIN gather_logstore_resources ...")

	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	// Check if LokiStack is installed (cluster-wide check)
	lokiList, err := l.client.DynamicClient.Resource(lokiGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		l.logger.Log("INFO: LokiStack CRD not available, skipping logstore collection")
		return nil
	}

	if len(lokiList.Items) == 0 {
		l.logger.Log("INFO: No LokiStack resources found in any namespace, skipping logstore collection")
		return nil
	}

	l.logger.Log("Gathering Lokistack resources")
	l.logger.Log("-- Gather Lokistack CR")

	// Write to namespace directory like other resources
	lokiFolder := filepath.Join(l.destDir, "namespaces", l.namespace, "loki.grafana.com", "lokistacks")
	if err := l.client.ListResources(ctx, lokiGVR, l.namespace, lokiFolder, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect LokiStack CR: %w", err)
	}

	l.logger.Log("END gather_logstore_resources ...")
	return nil
}
