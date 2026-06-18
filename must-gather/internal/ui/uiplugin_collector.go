package ui

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	uipluginGVR = schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "uiplugins",
	}
)

// UIPluginCollector collects UIPlugin and Console resources
type UIPluginCollector struct {
	client  *client.Client
	logger  api.Logger
	destDir api.Path
}

// NewUIPluginCollector creates a new UIPlugin collector
func NewUIPluginCollector(c *client.Client, logger api.Logger, destDir api.Path) *UIPluginCollector {
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

	// Check if UIPlugin is installed
	uiPluginList, err := u.client.DynamicClient.Resource(uipluginGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		u.logger.Info("UIPlugin CRD not available, skipping uiplugin collection")
		return nil
	}

	if len(uiPluginList.Items) == 0 {
		u.logger.Info("No UIPlugin resources found, skipping uiplugin collection")
		return nil
	}

	destDir := u.destDir.Add(cluster.ArtifactRoot).ForResource(uipluginGVR)

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

	consoleDestDir := u.destDir.Add(cluster.ArtifactRoot).ForResource(coGVR)

	if err := u.client.GetResource(ctx, coGVR, "", "console", consoleDestDir.Add("console.yaml")); err != nil {
		u.logger.Warn("Failed to collect console ClusterOperator: %v", err)
	}

	return nil
}
