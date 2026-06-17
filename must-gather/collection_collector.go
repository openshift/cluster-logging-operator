package mustgather

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CollectionCollector collects collector (vector) resources
type CollectionCollector struct {
	client    *Client
	logger    *Logger
	namespace string
}

// NewCollectionCollector creates a new collection resource collector
func NewCollectionCollector(client *Client, logger *Logger, namespace string) *CollectionCollector {
	return &CollectionCollector{
		client:    client,
		logger:    logger,
		namespace: namespace,
	}
}

// Name returns the name of this collector
func (c *CollectionCollector) Name() string {
	return "CollectionCollector"
}

// Collect performs the collection of collector resources
func (c *CollectionCollector) Collect(ctx context.Context, config *Config) error {
	c.logger.Log("- BEGIN <gather_collection_resources> for namespace: %s ...", c.namespace)

	// Get ClusterLogForwarder resources
	c.logger.Log("-- Exporting ClusterLogForwarder.observability.openshift.io resources")

	clfGVR := schema.GroupVersionResource{
		Group:    "observability.openshift.io",
		Version:  "v1",
		Resource: "clusterlogforwarders",
	}

	clfListUnstructured, err := c.client.dynamicClient.Resource(clfGVR).Namespace(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.logger.Log("WARNING: Failed to list ClusterLogForwarders: %v", err)
		return nil
	}

	// If no ClusterLogForwarders, nothing to collect
	if len(clfListUnstructured.Items) == 0 {
		c.logger.Log("- END <gather_collection_resources> for namespace: %s (no ClusterLogForwarders found)", c.namespace)
		return nil
	}

	// Create directory only if we have ClusterLogForwarders
	collectorFolder := filepath.Join(config.BaseCollectionPath, "cluster-logging", "namespaces", c.namespace)
	if err := os.MkdirAll(collectorFolder, 0755); err != nil {
		return fmt.Errorf("failed to create collector folder: %w", err)
	}

	// Process each ClusterLogForwarder
	for _, item := range clfListUnstructured.Items {
		collectorName := item.GetName()
		c.logger.Log("-- Gathering data for ClusterLogForwarder: %s", collectorName)

		// Save the CLF resource
		if err := c.client.ListResources(ctx, clfGVR, c.namespace, collectorFolder, metav1.ListOptions{}); err != nil {
			c.logger.Log("WARNING: Failed to save ClusterLogForwarder: %v", err)
		}

		// Get daemonset YAML
		c.logger.Log("--- Gathering DaemonSet %s", collectorName)
		dsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
		if err := c.client.GetResource(ctx, dsGVR, c.namespace, collectorName, filepath.Join(collectorFolder, fmt.Sprintf("daemonset_%s.yaml", collectorName))); err != nil {
			c.logger.Log("INFO: DaemonSet %s not found or failed: %v", collectorName, err)
		}

		// Get collector pods YAML and related events
		pods, err := c.client.GetPods(ctx, c.namespace, fmt.Sprintf("app.kubernetes.io/instance=%s,app.kubernetes.io/component=collector", collectorName))
		if err != nil {
			c.logger.Log("WARNING: Failed to get collector pods: %v", err)
		} else {
			for _, pod := range pods {
				c.logger.Log("--- Gathering collector pod: %s", pod.Name)

				// Save pod YAML
				podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
				if err := c.client.GetResource(ctx, podGVR, c.namespace, pod.Name, filepath.Join(collectorFolder, fmt.Sprintf("pod_%s.yaml", pod.Name))); err != nil {
					c.logger.Log("WARNING: Failed to get pod %s: %v", pod.Name, err)
				}

				// Save pod events
				if err := c.getPodEvents(ctx, c.namespace, pod.Name, collectorFolder); err != nil {
					c.logger.Log("INFO: Failed to get events for pod %s: %v", pod.Name, err)
				}

				// Save pod logs
				if err := c.getPodLogs(ctx, c.namespace, pod.Name, collectorFolder); err != nil {
					c.logger.Log("INFO: Failed to get logs for pod %s: %v", pod.Name, err)
				}
			}
		}

		// Get vector.toml from configmap
		configName := fmt.Sprintf("%s-config", collectorName)
		c.logger.Log("-- Gathering %s#vector.toml from namespace: %s", configName, c.namespace)
		if err := c.getVectorConfig(ctx, c.namespace, configName, collectorFolder); err != nil {
			c.logger.Log("INFO: ConfigMap %s not found or has no vector.toml: %v", configName, err)
		}
	}

	c.logger.Log("- END <gather_collection_resources> for namespace: %s ...", c.namespace)
	return nil
}

// getPodEvents gets events related to a pod
func (c *CollectionCollector) getPodEvents(ctx context.Context, namespace, podName, destFolder string) error {
	events, err := c.client.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		return err
	}

	if len(events.Items) == 0 {
		return nil
	}

	eventFile := filepath.Join(destFolder, fmt.Sprintf("pod_%s_events.yaml", podName))
	return c.client.WriteResourceToFile(events, eventFile)
}

// getPodLogs gets logs from all containers in a pod
func (c *CollectionCollector) getPodLogs(ctx context.Context, namespace, podName, destFolder string) error {
	pod, err := c.client.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Get logs for each container
	for _, container := range pod.Spec.Containers {
		logOpts := &corev1.PodLogOptions{
			Container: container.Name,
			TailLines: int64Ptr(1000), // Last 1000 lines
		}

		req := c.client.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
		logs, err := req.Stream(ctx)
		if err != nil {
			c.logger.Log("INFO: Failed to get logs for container %s in pod %s: %v", container.Name, podName, err)
			continue
		}
		defer logs.Close()

		logData, err := io.ReadAll(logs)
		if err != nil {
			c.logger.Log("WARNING: Failed to read logs for container %s in pod %s: %v", container.Name, podName, err)
			continue
		}

		logFile := filepath.Join(destFolder, fmt.Sprintf("pod_%s_container_%s.log", podName, container.Name))
		if err := os.WriteFile(logFile, logData, 0644); err != nil {
			c.logger.Log("WARNING: Failed to write log file for container %s: %v", container.Name, err)
		}
	}

	return nil
}

func int64Ptr(i int64) *int64 {
	return &i
}

// getVectorConfig extracts vector.toml from configmap
func (c *CollectionCollector) getVectorConfig(ctx context.Context, namespace, configName, destFolder string) error {
	cm, err := c.client.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	vectorToml, ok := cm.Data["vector.toml"]
	if !ok {
		return fmt.Errorf("vector.toml not found in configmap")
	}

	configFile := filepath.Join(destFolder, fmt.Sprintf("configmap_%s_vector.toml", configName))
	return os.WriteFile(configFile, []byte(vectorToml), 0644)
}
