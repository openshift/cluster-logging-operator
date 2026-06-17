package cluster

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects cluster-scoped resources
type Collector struct {
	client *client.Client
	logger api.Logger
}

// NewCollector creates a new cluster resource collector
func NewCollector(c *client.Client, logger api.Logger) *Collector {
	return &Collector{
		client: c,
		logger: logger,
	}
}

// Name returns the name of this collector
func (c *Collector) Name() string {
	return "ClusterCollector"
}

// Collect performs the collection of cluster-scoped resources
func (c *Collector) Collect(ctx context.Context, baseCollectionPath string) error {
	c.logger.Log("BEGIN inspecting cluster resources...")

	// Define cluster-scoped resources to collect (matching /tmp/foo reference)
	clusterResources := []schema.GroupVersionResource{
		// Core resources
		{Group: "", Version: "v1", Resource: "nodes"},

		// RBAC
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},

		// API Extensions
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},

		// OpenShift Config
		{Group: "config.openshift.io", Version: "v1", Resource: "clusterversions"},
	}

	destDir := filepath.Join(baseCollectionPath, "cluster-scoped-resources")

	var wg sync.WaitGroup
	for _, gvr := range clusterResources {
		wg.Add(1)
		go func(g schema.GroupVersionResource) {
			defer wg.Done()

			c.logger.Log("-- BEGIN inspecting cluster resource %s ...", g.Resource)

			// Use "core" for core resources (empty group) to match reference structure
			group := g.Group
			if group == "" {
				group = "core"
			}

			resourceDir := filepath.Join(destDir, group, g.Resource)

			if err := c.client.ListResources(ctx, g, "", resourceDir, metav1.ListOptions{}); err != nil {
				c.logger.Log("WARNING: Failed to collect %s: %v", g.Resource, err)
			}
		}(gvr)
	}

	wg.Wait()

	c.logger.Log("END inspecting cluster resources")
	return nil
}
