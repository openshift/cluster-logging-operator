package mustgather

import (
	"context"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// UIPluginCollector collects UIPlugin and Console resources
type UIPluginCollector struct {
	client *Client
	logger *Logger
}

// NewUIPluginCollector creates a new UIPlugin collector
func NewUIPluginCollector(client *Client, logger *Logger) *UIPluginCollector {
	return &UIPluginCollector{
		client: client,
		logger: logger,
	}
}

// Name returns the name of this collector
func (u *UIPluginCollector) Name() string {
	return "UIPluginCollector"
}

// Collect performs the collection of UIPlugin resources
func (u *UIPluginCollector) Collect(ctx context.Context, config *Config) error {
	u.logger.Log("BEGIN gathering uiplugin and console resources ...")

	uipluginGVR := schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "uiplugins",
	}

	destDir := filepath.Join(config.BaseCollectionPath, "cluster-scoped-resources", "console.openshift.io", "uiplugins")

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

	consoleDestDir := filepath.Join(config.BaseCollectionPath, "cluster-scoped-resources", "config.openshift.io", "clusteroperators")

	if err := u.client.GetResource(ctx, coGVR, "", "console", filepath.Join(consoleDestDir, "console.yaml")); err != nil {
		u.logger.Log("WARNING: Failed to collect console ClusterOperator: %v", err)
	}

	u.logger.Log("END gathering uiplugin and console resources")
	return nil
}
