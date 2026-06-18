package cluster

import (
	"context"
	"sync"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ArtifactRoot = "cluster-scoped-resources"
)

// Collector collects cluster-scoped resources
type Collector struct {
	client  *client.Client
	logger  api.Logger
	destDir api.Path
}

// NewCollector creates a new cluster resource collector
func NewCollector(c *client.Client, logger api.Logger, destDir api.Path) *Collector {
	return &Collector{
		client:  c,
		logger:  logger,
		destDir: destDir,
	}
}

// Name returns the name of this collector
func (c *Collector) Name() string {
	return "ClusterCollector"
}

// Collect performs the collection of cluster-scoped resources
func (c *Collector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	defer c.logger.Begin("inspecting cluster resources...")()

	// Use provided GVRs or default cluster-scoped resources
	clusterResources := gvrs
	if len(clusterResources) == 0 {
		// Default cluster-scoped resources to collect (matching /tmp/foo reference)
		clusterResources = []schema.GroupVersionResource{
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
	}

	basePath := api.NewArtifactPath(c.destDir.String(), ArtifactRoot)

	var wg sync.WaitGroup
	for _, gvr := range clusterResources {
		wg.Add(1)
		go func(g schema.GroupVersionResource) {
			defer wg.Done()
			defer c.logger.Begin("-- inspecting cluster resource %s ...", g.Resource)()

			// Use "core" for core resources (empty group) to match reference structure
			gvr := g
			if gvr.Group == "" {
				gvr.Group = "core"
			}

			resourcePath := basePath.ForResource(gvr)

			if err := c.client.ListResources(ctx, g, "", resourcePath, metav1.ListOptions{}); err != nil {
				c.logger.Warn("Failed to collect %s: %v", g.Resource, err)
			}
		}(gvr)
	}

	wg.Wait()

	return nil
}
