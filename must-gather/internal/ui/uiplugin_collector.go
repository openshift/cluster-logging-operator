package ui

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// UIPluginCollector collects UIPlugin and Console resources
type UIPluginCollector struct {
	client  *client.Client
	logger  api.Logger
	destDir string
}

// NewUIPluginCollector creates a new UIPlugin collector
func NewUIPluginCollector(c *client.Client, logger api.Logger, destDir string) *UIPluginCollector {
	return &UIPluginCollector{
		client:  c,
		logger:  logger,
		destDir: destDir,
	}
}

// Name returns the name of this collector
func (u *UIPluginCollector) Name() string {
	return "UIPluginCollector"
}

// Collect performs the collection of UIPlugin resources
func (u *UIPluginCollector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	defer u.logger.Begin("gathering uiplugin and console resources ...")()

	uipluginGVR := schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "uiplugins",
	}

	destDir := filepath.Join(u.destDir, "cluster-scoped-resources", "console.openshift.io", "uiplugins")

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

	consoleDestDir := filepath.Join(u.destDir, "cluster-scoped-resources", "config.openshift.io", "clusteroperators")

	if err := u.client.GetResource(ctx, coGVR, "", "console", filepath.Join(consoleDestDir, "console.yaml")); err != nil {
		u.logger.Warn("Failed to collect console ClusterOperator: %v", err)
	}

	return nil
}
