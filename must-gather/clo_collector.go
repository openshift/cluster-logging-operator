package mustgather

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CLOCollector collects cluster logging operator resources
type CLOCollector struct {
	client *Client
	logger *Logger
}

// NewCLOCollector creates a new CLO resource collector
func NewCLOCollector(client *Client, logger *Logger) *CLOCollector {
	return &CLOCollector{
		client: client,
		logger: logger,
	}
}

// Name returns the name of this collector
func (c *CLOCollector) Name() string {
	return "CLOCollector"
}

// Collect performs the collection of CLO resources
func (c *CLOCollector) Collect(ctx context.Context, config *Config) error {
	c.logger.Log("BEGIN <gather_cluster_logging_operator_resources> from namespace: %s ...", config.LoggingNamespace)

	cloFolder := filepath.Join(config.BaseCollectionPath, "cluster-logging", "clo")
	if err := os.MkdirAll(cloFolder, 0755); err != nil {
		return fmt.Errorf("failed to create CLO folder: %w", err)
	}

	// Only collect from openshift-logging namespace
	if config.LoggingNamespace == "openshift-logging" {
		c.logger.Log("Gathering data for 'cluster-logging-operator' from namespace: %s", config.LoggingNamespace)

		// Get CLO pods
		pods, err := c.client.GetPods(ctx, config.LoggingNamespace, "name=cluster-logging-operator")
		if err != nil {
			c.logger.Log("WARNING: Failed to get CLO pods: %v", err)
		} else {
			for _, pod := range pods {
				c.logger.Log("Inspecting %s", pod.Name)
				if err := c.getEnv(ctx, config.LoggingNamespace, pod.Name, cloFolder, "Dockerfile-.*operator*"); err != nil {
					c.logger.Log("WARNING: Failed to get env for pod %s: %v", pod.Name, err)
				}
			}
		}
	}

	// Get version from CSV
	c.logger.Log("Gathering 'version' from logging namespace: %s", config.LoggingNamespace)
	if err := c.getVersion(ctx, config.LoggingNamespace, cloFolder); err != nil {
		c.logger.Log("WARNING: Failed to get version: %v", err)
	}

	c.logger.Log("END <gather_cluster_logging_operator_resources> from namespace: %s ...", config.LoggingNamespace)
	return nil
}

// getEnv gets environment variables and build info from a pod
func (c *CLOCollector) getEnv(ctx context.Context, namespace, podName, destFolder, dockerfilePattern string) error {
	c.logger.Log("BEGIN get_env ...")
	c.logger.Log("---- Env for %s", podName)

	envFile := filepath.Join(destFolder, podName)
	var output strings.Builder

	// Get pod to find containers
	pods, err := c.client.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
	})
	if err != nil || len(pods.Items) == 0 {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	pod := pods.Items[0]

	for _, container := range pod.Spec.Containers {
		c.logger.Log("----- Inspecting container %s", container.Name)

		// Try to get dockerfile info
		lsOutput, err := c.client.ExecInPod(ctx, namespace, podName, container.Name, []string{"ls", "/root/buildinfo"})
		if err == nil {
			// Find dockerfile matching pattern
			re := regexp.MustCompile(dockerfilePattern)
			lines := strings.Split(lsOutput, "\n")
			for _, line := range lines {
				if re.MatchString(line) {
					c.logger.Log("----- Getting buildInfo")
					output.WriteString(fmt.Sprintf("Image info: %s\n", line))

					// Get build date
					buildDateOutput, err := c.client.ExecInPod(ctx, namespace, podName, container.Name,
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

		envOutput, err := c.client.ExecInPod(ctx, namespace, podName, container.Name, []string{"env"})
		if err == nil {
			envLines := strings.Split(envOutput, "\n")
			// Sort would be nice but let's keep it simple
			for _, line := range envLines {
				if line != "" {
					output.WriteString(line + "\n")
				}
			}
		}
	}

	if err := os.WriteFile(envFile, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	c.logger.Log("END get_env ...")
	return nil
}

// getVersion gets the CLO version from CSV
func (c *CLOCollector) getVersion(ctx context.Context, namespace, destFolder string) error {
	csvGVR := schema.GroupVersionResource{
		Group:    "operators.coreos.com",
		Version:  "v1alpha1",
		Resource: "clusterserviceversions",
	}

	csvListUnstructured, err := c.client.dynamicClient.Resource(csvGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CSVs: %w", err)
	}

	// Find cluster-logging CSV
	for _, item := range csvListUnstructured.Items {
		name := item.GetName()
		if strings.Contains(name, "cluster-logging") || strings.Contains(name, "clusterlogging") {
			displayName, _, _ := unstructured.NestedString(item.Object, "spec", "displayName")
			version, _, _ := unstructured.NestedString(item.Object, "spec", "version")

			versionContent := fmt.Sprintf("%s/must-gather\n%s\n", displayName, version)
			versionFile := filepath.Join(destFolder, "version")

			if err := os.WriteFile(versionFile, []byte(versionContent), 0644); err != nil {
				return fmt.Errorf("failed to write version file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("no cluster-logging CSV found")
}
