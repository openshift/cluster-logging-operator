package metrics

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects Prometheus rules and alerts
type Collector struct {
	client  *client.Client
	logger  api.Logger
	destDir api.Path
}

// NewCollector creates a new monitoring collector
func NewCollector(c *client.Client, logger api.Logger, destDir api.Path) *Collector {
	return &Collector{
		client:  c,
		logger:  logger,
		destDir: destDir,
	}
}

// Name returns the name of this collector
func (m *Collector) Name() string {
	return "MonitoringCollector"
}

// Collect performs the collection of monitoring resources
func (m *Collector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	defer m.logger.Begin("gathering alerts ...")()

	monitoringPath := filepath.Join(m.destDir.String(), "monitoring")
	if err := os.MkdirAll(monitoringPath, 0755); err != nil {
		return fmt.Errorf("failed to create monitoring folder: %w", err)
	}

	// Get Prometheus pods
	promPods, err := m.client.GetPods(ctx, "openshift-monitoring", "prometheus=k8s")
	if err != nil {
		m.logger.Warn("Failed to get Prometheus pods: %v", err)
		return nil
	}

	m.logger.Info("Found %d Prometheus replicas", len(promPods))

	// Get first ready pod
	readyPod := m.getFirstReadyPromPod(promPods)
	if readyPod == "" {
		m.logger.Warn("No ready Prometheus pod found")
		return nil
	}

	// Get Prometheus rules
	m.logger.Info("Getting rules from %s", readyPod)
	if err := m.promGet(ctx, readyPod, "rules", monitoringPath); err != nil {
		m.logger.Warn("Failed to get Prometheus rules: %v", err)
	}

	return nil
}

// promGet makes HTTP GET requests to prometheus /api/v1/<object>
func (m *Collector) promGet(ctx context.Context, pod, object, monitoringPath string) error {
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
func (m *Collector) getFirstReadyPromPod(pods []corev1.Pod) string {
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
