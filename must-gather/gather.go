package mustgather

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Gather is the main must-gather orchestrator
type Gather struct {
	config *Config
	client *Client
	logger *Logger
}

// NewGather creates a new must-gather orchestrator
func NewGather(baseCollectionPath, loggingNamespace string, logWriter io.Writer) (*Gather, error) {
	logger := NewLogger(logWriter)
	client, err := NewClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Set up paths
	absPath, err := filepath.Abs(baseCollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	config := &Config{
		BaseCollectionPath: absPath,
		LoggingNamespace:   loggingNamespace,
		LogFileName:        "gather-debug.log",
		Logger:             logWriter,
		Context:            context.Background(),
	}

	return &Gather{
		config: config,
		client: client,
		logger: logger,
	}, nil
}

// Run executes the must-gather collection
func (g *Gather) Run(ctx context.Context) error {
	g.logger.Log("..... Cluster Logging must-gather script started .....")
	g.logger.Log("must-gather logs are located at: '%s'", filepath.Join(g.config.BaseCollectionPath, g.config.LogFileName))

	// Ensure base collection path exists
	if err := os.MkdirAll(g.config.BaseCollectionPath, 0755); err != nil {
		return fmt.Errorf("failed to create base collection path: %w", err)
	}

	// Discover namespaces
	namespaces, err := g.discoverNamespaces(ctx)
	if err != nil {
		g.logger.Log("WARNING: Failed to discover namespaces: %v", err)
		namespaces = []string{g.config.LoggingNamespace}
	}

	// Create collectors
	collectors := g.createCollectors(ctx, namespaces)

	// Run collectors concurrently
	results := g.runCollectors(ctx, collectors)

	// Log results
	g.logResults(results)

	return nil
}

// discoverNamespaces discovers all relevant namespaces for collection
func (g *Gather) discoverNamespaces(ctx context.Context) ([]string, error) {
	namespaceSet := make(map[string]bool)

	// Standard namespaces
	standardNamespaces := []string{
		"openshift-operator-lifecycle-manager",
		g.config.LoggingNamespace,
		"openshift-operators-redhat",
		"openshift-operators",
		"openshift-monitoring", // Contains Prometheus pods needed by monitoring collector
	}

	for _, ns := range standardNamespaces {
		namespaceSet[ns] = true
	}

	// Find multi-forwarder namespaces
	clfGVR := schema.GroupVersionResource{
		Group:    "observability.openshift.io",
		Version:  "v1",
		Resource: "clusterlogforwarders",
	}

	// List all ClusterLogForwarders across all namespaces
	clfListUnstructured, err := g.client.dynamicClient.Resource(clfGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		g.logger.Log("WARNING: Failed to list ClusterLogForwarders: %v", err)
	} else {
		for _, item := range clfListUnstructured.Items {
			ns := item.GetNamespace()
			if ns != "" && ns != g.config.LoggingNamespace {
				namespaceSet[ns] = true
				g.logger.Log("Adding namespace '%s' to cluster resources list", ns)
			}
		}
	}

	// Convert set to slice
	namespaces := make([]string, 0, len(namespaceSet))
	for ns := range namespaceSet {
		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}

// createCollectors creates all collectors needed for the gathering
func (g *Gather) createCollectors(ctx context.Context, namespaces []string) []Collector {
	collectors := make([]Collector, 0)

	// Cluster-scoped resources collector
	collectors = append(collectors, NewClusterCollector(g.client, g.logger))

	// Namespace collectors
	collectors = append(collectors, NewNamespaceCollector(g.client, g.logger, namespaces))

	// UIPlugin collector (if installed)
	if g.isUIPluginInstalled(ctx) {
		collectors = append(collectors, NewUIPluginCollector(g.client, g.logger))
	}

	// Monitoring collector
	collectors = append(collectors, NewMonitoringCollector(g.client, g.logger))

	// LogStore collectors (LokiStack only)
	if g.isLokiStackInstalled(ctx) {
		collectors = append(collectors, NewLogStoreCollector(g.client, g.logger, "lokistack", g.config.LoggingNamespace))
	}

	return collectors
}

// runCollectors runs all collectors concurrently
func (g *Gather) runCollectors(ctx context.Context, collectors []Collector) []Result {
	var wg sync.WaitGroup
	resultsChan := make(chan Result, len(collectors))

	for _, collector := range collectors {
		wg.Add(1)
		go func(c Collector) {
			defer wg.Done()

			start := time.Now()
			err := c.Collect(ctx, g.config)
			duration := time.Since(start)

			resultsChan <- Result{
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
	results := make([]Result, 0, len(collectors))
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// logResults logs the results of all collectors
func (g *Gather) logResults(results []Result) {
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

	uiPluginList, err := g.client.dynamicClient.Resource(uipluginGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false
	}

	return len(uiPluginList.Items) > 0
}

// isLokiStackInstalled checks if LokiStack is installed
func (g *Gather) isLokiStackInstalled(ctx context.Context) bool {
	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	lokiList, err := g.client.dynamicClient.Resource(lokiGVR).Namespace(g.config.LoggingNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false
	}

	return len(lokiList.Items) > 0
}
