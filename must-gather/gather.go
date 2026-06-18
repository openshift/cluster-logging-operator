package mustgather

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/cluster"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/collection"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/logstore/lokistack"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/metrics"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/namespace"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/ui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Standard namespaces
	standardNamespaces = []string{
		"openshift-operator-lifecycle-manager",
		"openshift-operators-redhat",
		"openshift-operators",
		"openshift-monitoring", // Contains Prometheus pods needed by monitoring collector
	}
)

// Gather is the main must-gather orchestrator
type Gather struct {
	config *api.Config
	client *client.Client
	logger api.Logger
}

// NewGather creates a new must-gather orchestrator
func NewGather(baseCollectionPath, loggingNamespace string, logWriter io.Writer) (*Gather, error) {
	logger := api.NewLogger(logWriter)
	k8sClient, err := client.NewClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Set up paths
	absPath, err := filepath.Abs(baseCollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	config := &api.Config{
		DestDir:          absPath,
		LoggingNamespace: loggingNamespace,
		LogFileName:      "gather-debug.log",
		Logger:           logWriter,
	}

	return &Gather{
		config: config,
		client: k8sClient,
		logger: logger,
	}, nil
}

// Run executes the must-gather collection
func (g *Gather) Run(ctx context.Context) error {
	g.logger.Log("..... Cluster Logging must-gather script started .....")
	g.logger.Log("must-gather logs are located at: '%s'", filepath.Join(g.config.DestDir, g.config.LogFileName))

	// Ensure base collection path exists
	if err := os.MkdirAll(g.config.DestDir, 0755); err != nil {
		return fmt.Errorf("failed to create base collection path: %w", err)
	}

	// Create collectors
	collectors := g.createCollectors()

	// Run collectors concurrently
	results := g.runCollectors(ctx, collectors)

	// Log results
	g.logResults(results)

	return nil
}

// createCollectors creates all collectors needed for the gathering
func (g *Gather) createCollectors() []api.Collector {
	collectors := make([]api.Collector, 0)

	// Cluster-scoped resources collector
	collectors = append(collectors, cluster.NewCollector(g.client, g.logger, g.config.DestDir))

	// Namespace collectors
	collectors = append(collectors, namespace.NewCollector(g.client, g.logger, standardNamespaces, g.config.DestDir))

	// Log Collection collector
	collectors = append(collectors, collection.NewCollector(g.client, g.logger, g.config.LoggingNamespace, g.config.DestDir))

	// UIPlugin collector (if installed)
	if g.isUIPluginInstalled(context.Background()) {
		collectors = append(collectors, ui.NewUIPluginCollector(g.client, g.logger, g.config.DestDir))
	}

	// Monitoring collector
	collectors = append(collectors, metrics.NewCollector(g.client, g.logger, g.config.DestDir))

	// LogStore collector (checks for LokiStack installation internally)
	collectors = append(collectors, lokistack.NewCollector(g.client, g.logger, g.config.LoggingNamespace, g.config.DestDir))

	return collectors
}

// runCollectors runs all collectors concurrently
func (g *Gather) runCollectors(ctx context.Context, collectors []api.Collector) []api.Result {
	var wg sync.WaitGroup
	resultsChan := make(chan api.Result, len(collectors))

	for _, collector := range collectors {
		wg.Add(1)
		go func(c api.Collector) {
			defer wg.Done()

			start := time.Now()
			// Call Collect with no GVRs to use defaults
			err := c.Collect(ctx)
			duration := time.Since(start)

			resultsChan <- api.Result{
				CollectorName: c.Name(),
				Error:         err,
				Duration:      duration,
			}
		}(collector)
	}

	// Wait for all collectors to finish
	wg.Wait()
	close(resultsChan)

	// Collect results
	results := make([]api.Result, 0, len(collectors))
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// logResults logs the results of all collectors
func (g *Gather) logResults(results []api.Result) {
	g.logger.Log("=== Must-gather collection complete ===")
	for _, result := range results {
		if result.Error != nil {
			g.logger.Log("FAILED: %s (took %v): %v", result.CollectorName, result.Duration, result.Error)
		} else {
			g.logger.Log("SUCCESS: %s (took %v)", result.CollectorName, result.Duration)
		}
	}
}

// isUIPluginInstalled checks if the UIPlugin is installed
func (g *Gather) isUIPluginInstalled(ctx context.Context) bool {
	uipluginGVR := schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "uiplugins",
	}

	uiPluginList, err := g.client.DynamicClient.Resource(uipluginGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false
	}

	return len(uiPluginList.Items) > 0
}

