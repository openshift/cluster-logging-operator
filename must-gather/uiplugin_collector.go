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

// UIPluginCollector collects UIPlugin and Console resources
type UIPluginCollector struct {
	client *client.Client
	logger *api.Logger
}

// NewUIPluginCollector creates a new UIPlugin collector
func NewUIPluginCollector(c *client.Client, logger *api.Logger) *UIPluginCollector {
	return &UIPluginCollector{
		client: c,
		logger: logger,
	}
}

// Name returns the name of this collector
func (u *UIPluginCollector) Name() string {
	return "UIPluginCollector"
}

// Collect performs the collection of UIPlugin resources
func (u *UIPluginCollector) Collect(ctx context.Context, config *api.Config) error {
	u.logger.Log("BEGIN gathering uiplugin and console resources ...")

	uipluginGVR := schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "uiplugins",
	}

	destDir := filepath.Join(config.DestDir, "cluster-scoped-resources", "console.openshift.io", "uiplugins")

	// Collect UIPlugin resources
	if err := u.client.ListResources(ctx, uipluginGVR, "", destDir, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect UIPlugin resources: %w", err)
	}

	// Collect Console ClusterOperator
	coGVR := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "clusteroperators",
	}

	consoleDestDir := filepath.Join(config.DestDir, "cluster-scoped-resources", "config.openshift.io", "clusteroperators")

	if err := u.client.GetResource(ctx, coGVR, "", "console", filepath.Join(consoleDestDir, "console.yaml")); err != nil {
		u.logger.Log("WARNING: Failed to collect console ClusterOperator: %v", err)
	}

	u.logger.Log("END gathering uiplugin and console resources")
	return nil
}
