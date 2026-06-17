package mustgather

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// LogStoreCollector collects log store resources (Elasticsearch or LokiStack)
type LogStoreCollector struct {
	client        *Client
	logger        *Logger
	logStoreType  string // "elasticsearch" or "lokistack"
	namespace     string
}

// NewLogStoreCollector creates a new log store resource collector
func NewLogStoreCollector(client *Client, logger *Logger, logStoreType, namespace string) *LogStoreCollector {
	return &LogStoreCollector{
		client:       client,
		logger:       logger,
		logStoreType: logStoreType,
		namespace:    namespace,
	}
}

// Name returns the name of this collector
func (l *LogStoreCollector) Name() string {
	return "LogStoreCollector"
}

// Collect performs the collection of log store resources
func (l *LogStoreCollector) Collect(ctx context.Context, config *Config) error {
	l.logger.Log("BEGIN gather_logstore_resources ...")
	l.logger.Log("Gathering data for logstore component")

	esFolder := filepath.Join(config.BaseCollectionPath, "cluster-logging", "es")
	if err := os.MkdirAll(esFolder, 0755); err != nil {
		return fmt.Errorf("failed to create logstore folder: %w", err)
	}

	if l.logStoreType == "elasticsearch" {
		if err := l.collectElasticsearch(ctx, config, esFolder); err != nil {
			l.logger.Log("WARNING: Failed to collect Elasticsearch resources: %v", err)
		}
	}

	if l.logStoreType == "lokistack" {
		if err := l.collectLokiStack(ctx, config); err != nil {
			l.logger.Log("WARNING: Failed to collect LokiStack resources: %v", err)
		}
	}

	l.logger.Log("END gather_logstore_resources ...")
	return nil
}

// collectElasticsearch collects Elasticsearch-specific resources
func (l *LogStoreCollector) collectElasticsearch(ctx context.Context, config *Config, esFolder string) error {
	l.logger.Log("-- Checking Elasticsearch health")

	// Get Elasticsearch pods
	esPods, err := l.client.GetPods(ctx, l.namespace, "component=elasticsearch")
	if err != nil {
		return fmt.Errorf("failed to get Elasticsearch pods: %w", err)
	}

	for _, pod := range esPods {
		l.logger.Log("---- Elasticsearch pod: %s", pod.Name)

		// Get environment info
		if err := l.getEnv(ctx, pod.Name, esFolder); err != nil {
			l.logger.Log("WARNING: Failed to get env for pod %s: %v", pod.Name, err)
		}

		// List ES storage
		if err := l.listESStorage(ctx, pod.Name, esFolder); err != nil {
			l.logger.Log("WARNING: Failed to list storage for pod %s: %v", pod.Name, err)
		}
	}

	// Get Elasticsearch cluster status from one running pod
	runningPod := l.getFirstRunningPod(esPods)
	if runningPod != "" {
		l.logger.Log("-- Getting Elasticsearch cluster info from pod %s", runningPod)
		if err := l.getElasticsearchStatus(ctx, runningPod, esFolder); err != nil {
			l.logger.Log("WARNING: Failed to get Elasticsearch status: %v", err)
		}
	}

	// Gather Elasticsearch CR
	l.logger.Log("-- Gather Elasticsearch CR")
	esGVR := schema.GroupVersionResource{
		Group:    "logging.openshift.io",
		Version:  "v1",
		Resource: "elasticsearches",
	}
	if err := l.client.ListResources(ctx, esGVR, l.namespace, esFolder, metav1.ListOptions{}); err != nil {
		l.logger.Log("WARNING: Failed to collect Elasticsearch CR: %v", err)
	}

	return nil
}

// collectLokiStack collects LokiStack resources
func (l *LogStoreCollector) collectLokiStack(ctx context.Context, config *Config) error {
	l.logger.Log("Gathering Lokistack resources")
	l.logger.Log("-- Gather Lokistack CR")

	lokiGVR := schema.GroupVersionResource{
		Group:    "loki.grafana.com",
		Version:  "v1",
		Resource: "lokistacks",
	}

	lokiFolder := filepath.Join(config.BaseCollectionPath, "cluster-logging", "lokistack")
	if err := l.client.ListResources(ctx, lokiGVR, l.namespace, lokiFolder, metav1.ListOptions{}); err != nil {
		return fmt.Errorf("failed to collect LokiStack CR: %w", err)
	}

	return nil
}

// getEnv gets environment variables from a pod
func (l *LogStoreCollector) getEnv(ctx context.Context, podName, destFolder string) error {
	envFile := filepath.Join(destFolder, podName)
	var output strings.Builder

	// Get environment variables
	envOutput, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", []string{"env"})
	if err != nil {
		return err
	}

	output.WriteString("-- Environment Variables\n")
	output.WriteString(envOutput)

	return os.WriteFile(envFile, []byte(output.String()), 0644)
}

// listESStorage lists Elasticsearch storage information
func (l *LogStoreCollector) listESStorage(ctx context.Context, podName, destFolder string) error {
	// Get mount path from pod spec
	pods, err := l.client.clientset.CoreV1().Pods(l.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
	})
	if err != nil || len(pods.Items) == 0 {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	pod := pods.Items[0]
	var mountPath string

	// Find elasticsearch-storage volume mount
	for _, container := range pod.Spec.Containers {
		if container.Name == "elasticsearch" {
			for _, mount := range container.VolumeMounts {
				if mount.Name == "elasticsearch-storage" {
					mountPath = mount.MountPath
					break
				}
			}
		}
	}

	if mountPath == "" {
		return fmt.Errorf("elasticsearch-storage mount not found")
	}

	// List files
	lsOutput, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", []string{"ls", "-lR", mountPath})
	if err != nil {
		return err
	}

	storageFile := filepath.Join(destFolder, podName)
	var output strings.Builder
	output.WriteString("-- Persistence files\n")
	output.WriteString(lsOutput)

	// Get disk usage
	dfOutput, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", []string{"df", "-h", mountPath})
	if err == nil {
		storageFile = filepath.Join(destFolder, fmt.Sprintf("%s-storage", podName))
		var storageOutput strings.Builder
		storageOutput.WriteString("-- Persistence storage size\n")
		storageOutput.WriteString(dfOutput)
		os.WriteFile(storageFile, []byte(storageOutput.String()), 0644)
	}

	return os.WriteFile(filepath.Join(destFolder, podName), []byte(output.String()), 0644)
}

// getElasticsearchStatus gets Elasticsearch cluster status
func (l *LogStoreCollector) getElasticsearchStatus(ctx context.Context, podName, esFolder string) error {
	clusterFolder := filepath.Join(esFolder, "cluster-elasticsearch")
	if err := os.MkdirAll(clusterFolder, 0755); err != nil {
		return err
	}

	curlCmd := []string{"curl", "-s", "--max-time", "20",
		"--key", "/etc/elasticsearch/secret/admin-key",
		"--cert", "/etc/elasticsearch/secret/admin-cert",
		"--cacert", "/etc/elasticsearch/secret/admin-ca",
		"https://localhost:9200"}

	// Get various ES API endpoints
	catItems := []string{"health", "nodes", "aliases", "thread_pool"}
	for _, item := range catItems {
		cmd := append(curlCmd, fmt.Sprintf("/_cat/%s?v", item))
		output, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", cmd)
		if err != nil {
			l.logger.Log("WARNING: Failed to get %s: %v", item, err)
			continue
		}
		os.WriteFile(filepath.Join(clusterFolder, fmt.Sprintf("%s.cat", item)), []byte(output), 0644)
	}

	// Get indices
	cmd := append(curlCmd, "/_cat/indices?v&bytes=m")
	output, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", cmd)
	if err == nil {
		os.WriteFile(filepath.Join(clusterFolder, "indices.cat"), []byte(output), 0644)
	}

	// Get cluster health
	cmd = append(curlCmd, "/_cat/health?h=status")
	healthOutput, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", cmd)
	if err == nil {
		health := strings.TrimSpace(healthOutput)
		if health != "green" {
			l.logger.Log("Gathering additional cluster information. Cluster status is %s", health)

			// Get recovery, shards, pending tasks
			additionalItems := []string{"recovery", "shards", "pending_tasks"}
			for _, item := range additionalItems {
				cmd := append(curlCmd, fmt.Sprintf("/_cat/%s?v", item))
				output, err := l.client.ExecInPod(ctx, l.namespace, podName, "elasticsearch", cmd)
				if err == nil {
					os.WriteFile(filepath.Join(clusterFolder, fmt.Sprintf("%s.cat", item)), []byte(output), 0644)
				}
			}
		}
	}

	return nil
}

// getFirstRunningPod returns the first running pod from a list
func (l *LogStoreCollector) getFirstRunningPod(pods []corev1.Pod) string {
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			return pod.Name
		}
	}
	return ""
}
