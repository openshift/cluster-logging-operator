package lokistack

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects LokiStack resources
type Collector struct {
	client    *client.Client
	logger    api.Logger
	namespace string
	destDir   api.Path
}

// NewCollector creates a new LokiStack collector
func NewCollector(c *client.Client, logger api.Logger, namespace string, destDir api.Path) *Collector {
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
	defer l.logger.Begin("gather_logstore_resources ...")()

	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	// Check if LokiStack is installed (cluster-wide check)
	lokiList, err := l.client.DynamicClient.Resource(lokiGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		// Only skip if CRD doesn't exist (NotFound or NoKindMatchError)
		// Return other errors (RBAC, API issues, etc.) to caller
		if kerrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			l.logger.Info("LokiStack CRD not available, skipping logstore collection")
			return nil
		}
		return fmt.Errorf("failed to check for LokiStack resources: %w", err)
	}

	if len(lokiList.Items) == 0 {
		l.logger.Info("No LokiStack resources found in any namespace, skipping logstore collection")
		return nil
	}

	l.logger.Log("Gathering Lokistack resources")
	l.logger.Log("-- Gather Lokistack CR")

	// Write to namespace directory like other resources
	lokiFolder := l.destDir.Add("namespaces", l.namespace).ForResource(lokiGVR)
	if err := l.client.ListResources(ctx, lokiGVR, l.namespace, lokiFolder, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect LokiStack CR: %w", err)
	}

	return nil
}
