package collection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/namespace"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects cluster logging operator resources
type Collector struct {
	client    *client.Client
	logger    api.Logger
	namespace string
	destDir   api.Path
}

// NewCollector creates a new CLO resource collector
func NewCollector(c *client.Client, logger api.Logger, loggingNamespace string, destDir api.Path) *Collector {
	return &Collector{
		client:    c,
		logger:    logger,
		namespace: loggingNamespace,
		destDir:   destDir,
	}
}

// Name returns the name of this collector
func (c *Collector) Name() string {
	return "LogCollectorCollector"
}

// Collect performs the collection of CLO resources
func (c *Collector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	defer c.logger.Begin("gather cluster-logging-operator from namespace: %s ...", c.namespace)()

	// Collect operator-specific artifacts (only if operator is deployed)
	operatorFound, err := c.CollectForOperator(ctx)
	if err != nil {
		c.logger.Warn("failed to collect operator artifacts in namespace %s: %v", c.namespace, err)
	}

	// Discover namespaces with ClusterLogForwarders (independent of operator presence)
	namespaces, err := c.discoverNamespaces(ctx)
	if err != nil {
		return err
	}

	// If no namespaces found and operator not present, nothing to collect
	if len(namespaces) == 0 && !operatorFound {
		c.logger.Info("No cluster-logging-operator or ClusterLogForwarders found, skipping collection")
		return nil
	}

	// Collect namespace resources with collection-specific GVRs
	gvrs = append(gvrs, clfGVR, lfmeGVR)
	nsCollector := namespace.NewCollector(c.client, c.logger, namespaces, c.destDir)
	return nsCollector.Collect(ctx, gvrs...)
}

// discoverNamespaces discovers all relevant namespaces for collection
func (c *Collector) discoverNamespaces(ctx context.Context) ([]string, error) {
	namespaceSet := make(map[string]bool)

	// Always include the logging namespace itself
	namespaceSet[c.namespace] = true

	// List all ClusterLogForwarders across all namespaces
	clfListUnstructured, err := c.client.DynamicClient.Resource(clfGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.logger.Warn("Failed to list ClusterLogForwarders: %v", err)
	} else {
		for _, item := range clfListUnstructured.Items {
			ns := item.GetNamespace()
			if ns != "" {
				namespaceSet[ns] = true
				if ns != c.namespace {
					c.logger.Log("Adding namespace '%s' to cluster resources list", ns)
				}
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

func (c *Collector) CollectForOperator(ctx context.Context) (bool, error) {
	defer c.logger.Begin("<gather_cluster_logging_operator_resources> from namespace: %s ...", c.namespace)()

	// Try to get CLO pods with standard Kubernetes label first
	pods, err := c.client.GetPods(ctx, c.namespace, "app.kubernetes.io/name=cluster-logging-operator")
	if err != nil {
		c.logger.Warn("Failed to get CLO pods with standard label: %v", err)
		return false, nil
	}

	// If not found with standard label, try legacy label
	if len(pods) == 0 {
		pods, err = c.client.GetPods(ctx, c.namespace, "name=cluster-logging-operator")
		if err != nil {
			c.logger.Warn("Failed to get CLO pods with legacy label: %v", err)
			return false, nil
		}
	}

	// Return early if no operator pods found with either label
	if len(pods) == 0 {
		c.logger.Info("No cluster-logging-operator pods found in namespace %s, skipping operator-specific collection", c.namespace)
		return false, nil
	}

	cloFolder := filepath.Join(c.destDir.String(), "namespaces", c.namespace, "core", "pods")
	if err := os.MkdirAll(cloFolder, 0755); err != nil {
		return true, fmt.Errorf("failed to create CLO folder: %w", err)
	}

	c.logger.Log("Gathering data for 'cluster-logging-operator' from namespace: %s", c.namespace)
	for _, pod := range pods {
		c.logger.Log("Inspecting %s", pod.Name)
		if err := c.getEnv(ctx, c.namespace, pod.Name, cloFolder, "Dockerfile-.*operator*"); err != nil {
			c.logger.Warn("Failed to get env for pod %s: %v", pod.Name, err)
		}
	}

	return true, nil
}

// getEnv gets environment variables and build info from a pod
func (c *Collector) getEnv(ctx context.Context, namespace, podName, destFolder, dockerfilePattern string) error {
	defer c.logger.Begin("get_env ...")()
	c.logger.Log("---- Env for %s", podName)

	destDir := filepath.Join(destFolder, podName)
	envFile := filepath.Join(destDir, "env.txt")
	var output strings.Builder

	// Get pod to find containers
	pod, err := c.client.GetPod(ctx, namespace, podName)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	for _, container := range pod.Spec.Containers {
		c.logger.Log("----- Inspecting container %s", container.Name)

		// Try to get dockerfile info
		lsOutput, err := c.client.PodExec(ctx, namespace, podName, container.Name, []string{"ls", "/root/buildinfo"})
		if err == nil {
			// Find dockerfile matching pattern
			re := regexp.MustCompile(dockerfilePattern)
			lines := strings.Split(lsOutput, "\n")
			for _, line := range lines {
				if re.MatchString(line) {
					c.logger.Log("----- Getting buildInfo")
					output.WriteString(fmt.Sprintf("Image info: %s\n", line))

					// Get build date
					buildDateOutput, err := c.client.PodExec(ctx, namespace, podName, container.Name,
						[]string{"grep", "-o", "\"build-date\"=\"[^[:blank:]]*\"", "/root/buildinfo/" + line})
					if err == nil {
						output.WriteString(buildDateOutput)
					} else {
						c.logger.Log("---- Unable to get build date")
					}
					break
				}
			}
		}

		// Get environment variables
		c.logger.Log("----- Getting environment variables")
		output.WriteString("-- Environment Variables\n")

		envOutput, err := c.client.PodExec(ctx, namespace, podName, container.Name, []string{"env"})
		if err == nil {
			envLines := strings.Split(envOutput, "\n")
			sort.Strings(envLines)
			for _, line := range envLines {
				if line != "" {
					output.WriteString(line + "\n")
				}
			}
		}
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}
	if err := os.WriteFile(envFile, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}
