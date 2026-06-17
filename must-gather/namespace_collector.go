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

// NamespaceCollector collects namespace-scoped resources
type NamespaceCollector struct {
	client     *Client
	logger     *Logger
	namespaces []string
}

// NewNamespaceCollector creates a new namespace resource collector
func NewNamespaceCollector(client *Client, logger *Logger, namespaces []string) *NamespaceCollector {
	return &NamespaceCollector{
		client:     client,
		logger:     logger,
		namespaces: namespaces,
	}
}

// Name returns the name of this collector
func (n *NamespaceCollector) Name() string {
	return "NamespaceCollector"
}

// Collect performs the collection of namespace-scoped resources
func (n *NamespaceCollector) Collect(ctx context.Context, config *Config) error {
	n.logger.Log("BEGIN inspecting namespaced resources...")

	// Define namespace-scoped resources to collect (matching oc adm inspect behavior)
	namespacedResources := []schema.GroupVersionResource{
		// Core resources
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "serviceaccounts"},
		{Group: "", Version: "v1", Resource: "events"},
		{Group: "", Version: "v1", Resource: "endpoints"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "replicationcontrollers"},

		// Apps
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},

		// Batch
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1", Resource: "cronjobs"},

		// RBAC
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},

		// Networking
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
		{Group: "discovery.k8s.io", Version: "v1", Resource: "endpointslices"},

		// Autoscaling
		{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},

		// Policy
		{Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"},

		// OpenShift Apps
		{Group: "apps.openshift.io", Version: "v1", Resource: "deploymentconfigs"},

		// OpenShift Build
		{Group: "build.openshift.io", Version: "v1", Resource: "builds"},
		{Group: "build.openshift.io", Version: "v1", Resource: "buildconfigs"},

		// OpenShift Image
		{Group: "image.openshift.io", Version: "v1", Resource: "imagestreams"},

		// OpenShift Route
		{Group: "route.openshift.io", Version: "v1", Resource: "routes"},

		// OpenShift Monitoring
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors"},
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "podmonitors"},
		{Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheusrules"},

		// OVN Kubernetes
		{Group: "k8s.ovn.org", Version: "v1", Resource: "egressfirewalls"},
		{Group: "k8s.ovn.org", Version: "v1", Resource: "egressqoses"},
		{Group: "k8s.ovn.org", Version: "v1", Resource: "userdefinednetworks"},

		// Operators
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "installplans"},
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "subscriptions"},
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions"},

		// Cluster Logging specific
		{Group: "observability.openshift.io", Version: "v1", Resource: "logfilemetricexporters"},
		{Group: "observability.openshift.io", Version: "v1", Resource: "clusterlogforwarders"},
	}

	for _, ns := range n.namespaces {
		n.logger.Log("-- BEGIN inspecting namespace %s ...", ns)

		// First collect the namespace itself
		nsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
		nsDir := filepath.Join(config.BaseCollectionPath, "namespaces", ns)

		if err := n.client.GetResource(ctx, nsGVR, "", ns, filepath.Join(nsDir, "namespace.yaml")); err != nil {
			n.logger.Log("WARNING: Failed to collect namespace %s: %v", ns, err)
			continue
		}

		// Collect resources in the namespace
		for _, gvr := range namespacedResources {
			// Use "core" for core resources (empty group) to match reference structure
			group := gvr.Group
			if group == "" {
				group = "core"
			}

			resourceDir := filepath.Join(nsDir, group, gvr.Resource)

			if err := n.client.ListResources(ctx, gvr, ns, resourceDir, metav1.ListOptions{}); err != nil {
				// Some resources may not exist in all namespaces, just log and continue
				n.logger.Log("INFO: Skipped %s in namespace %s: %v", gvr.Resource, ns, err)
				continue
			}
		}

		// Collect pod logs for all pods in the namespace
		n.logger.Log("-- Collecting pod logs for namespace %s ...", ns)
		if err := n.collectPodLogs(ctx, ns, nsDir); err != nil {
			n.logger.Log("WARNING: Failed to collect pod logs for namespace %s: %v", ns, err)
		}

		n.logger.Log("-- END inspecting namespace %s", ns)
	}

	n.logger.Log("END inspecting namespaced resources")
	return nil
}

// collectPodLogs collects logs for all pods in a namespace
func (n *NamespaceCollector) collectPodLogs(ctx context.Context, namespace, nsDir string) error {
	// Get all pods in the namespace
	pods, err := n.client.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		podDir := filepath.Join(nsDir, "pods", pod.Name)

		// Save pod YAML
		podYamlPath := filepath.Join(podDir, fmt.Sprintf("%s.yaml", pod.Name))
		if err := n.client.WriteResourceToFile(&pod, podYamlPath); err != nil {
			n.logger.Log("WARNING: Failed to save pod YAML for %s: %v", pod.Name, err)
		}

		// Collect logs for each container
		for _, container := range pod.Spec.Containers {
			containerDir := filepath.Join(podDir, container.Name, container.Name, "logs")

			// Collect current logs
			if err := n.collectContainerLog(ctx, namespace, pod.Name, container.Name, containerDir, "current.log", false); err != nil {
				n.logger.Log("INFO: Failed to get current logs for pod %s container %s: %v", pod.Name, container.Name, err)
			}

			// Collect previous logs (from restarts)
			if err := n.collectContainerLog(ctx, namespace, pod.Name, container.Name, containerDir, "previous.log", true); err != nil {
				// Previous logs may not exist if the container hasn't restarted, don't log as warning
				continue
			}
		}

		// Collect logs for init containers if any
		for _, container := range pod.Spec.InitContainers {
			containerDir := filepath.Join(podDir, container.Name, container.Name, "logs")

			// Collect current logs
			if err := n.collectContainerLog(ctx, namespace, pod.Name, container.Name, containerDir, "current.log", false); err != nil {
				n.logger.Log("INFO: Failed to get current logs for init container %s in pod %s: %v", container.Name, pod.Name, err)
			}

			// Collect previous logs (from restarts)
			if err := n.collectContainerLog(ctx, namespace, pod.Name, container.Name, containerDir, "previous.log", true); err != nil {
				// Previous logs may not exist if the container hasn't restarted
				continue
			}
		}
	}

	return nil
}

// collectContainerLog collects a single container log
func (n *NamespaceCollector) collectContainerLog(ctx context.Context, namespace, podName, containerName, destDir, filename string, previous bool) error {
	logOpts := &corev1.PodLogOptions{
		Container: containerName,
		Previous:  previous,
	}

	req := n.client.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer logs.Close()

	logData, err := io.ReadAll(logs)
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	// Only create directory and write file if we have log data
	if len(logData) == 0 {
		return nil
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	logFile := filepath.Join(destDir, filename)
	if err := os.WriteFile(logFile, logData, 0644); err != nil {
		return fmt.Errorf("failed to write log file %s: %w", logFile, err)
	}

	return nil
}
