package mustgather

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

// MonitoringCollector collects Prometheus rules and alerts
type MonitoringCollector struct {
	client *Client
	logger *Logger
}

// NewMonitoringCollector creates a new monitoring collector
func NewMonitoringCollector(client *Client, logger *Logger) *MonitoringCollector {
	return &MonitoringCollector{
		client: client,
		logger: logger,
	}
}

// Name returns the name of this collector
func (m *MonitoringCollector) Name() string {
	return "MonitoringCollector"
}

// Collect performs the collection of monitoring resources
func (m *MonitoringCollector) Collect(ctx context.Context, config *Config) error {
	m.logger.Log("BEGIN gathering alerts ...")

	monitoringPath := filepath.Join(config.BaseCollectionPath, "monitoring")
	if err := os.MkdirAll(monitoringPath, 0755); err != nil {
		return fmt.Errorf("failed to create monitoring folder: %w", err)
	}

	// Get Prometheus pods
	promPods, err := m.client.GetPods(ctx, "openshift-monitoring", "prometheus=k8s")
	if err != nil {
		m.logger.Log("WARNING: Failed to get Prometheus pods: %v", err)
		return nil
	}

	m.logger.Log("INFO: Found %d Prometheus replicas", len(promPods))

	// Get first ready pod
	readyPod := m.getFirstReadyPromPod(promPods)
	if readyPod == "" {
		m.logger.Log("WARNING: No ready Prometheus pod found")
		return nil
	}

	// Get Prometheus rules
	m.logger.Log("INFO: Getting rules from %s", readyPod)
	if err := m.promGet(ctx, readyPod, "rules", monitoringPath); err != nil {
		m.logger.Log("WARNING: Failed to get Prometheus rules: %v", err)
	}

	m.logger.Log("END gathering alerts")
	return nil
}

// promGet makes HTTP GET requests to prometheus /api/v1/<object>
func (m *MonitoringCollector) promGet(ctx context.Context, pod, object, monitoringPath string) error {
	resultPath := filepath.Join(monitoringPath, "prometheus", object)
	if err := os.MkdirAll(filepath.Dir(resultPath), 0755); err != nil {
		return fmt.Errorf("failed to create result directory: %w", err)
	}

	// Execute curl command in Prometheus pod
	cmd := []string{"/bin/bash", "-c",
		fmt.Sprintf("curl -sG http://localhost:9090/api/v1/%s", object)}

	output, err := m.client.ExecInPod(ctx, "openshift-monitoring", pod, "prometheus", cmd)
	if err != nil {
		// Write error to stderr file
		stderrFile := fmt.Sprintf("%s.stderr", resultPath)
		os.WriteFile(stderrFile, []byte(err.Error()), 0644)
		return err
	}

	// Write output to json file
	jsonFile := fmt.Sprintf("%s.json", resultPath)
	return os.WriteFile(jsonFile, []byte(output), 0644)
}

// getFirstReadyPromPod returns the first ready Prometheus pod
func (m *MonitoringCollector) getFirstReadyPromPod(pods []corev1.Pod) string {
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			// Check if all containers are ready
			allReady := true
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					allReady = false
					break
				}
			}
			if allReady {
				return pod.Name
			}
		}
	}
	return ""
}
