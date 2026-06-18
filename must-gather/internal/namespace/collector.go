package namespace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Collector collects namespace-scoped resources
type Collector struct {
	client     *client.Client
	logger     api.Logger
	namespaces []string
	destDir    string
}

// NewCollector creates a new namespace resource collector
func NewCollector(c *client.Client, logger api.Logger, namespaces []string, destDir string) *Collector {
	return &Collector{
		client:     c,
		logger:     logger,
		namespaces: namespaces,
		destDir:    destDir,
	}
}

// Name returns the name of this collector
func (n *Collector) Name() string {
	return "Collector"
}

// Collect performs the collection of namespace-scoped resources
func (n *Collector) Collect(ctx context.Context, gvrs ...schema.GroupVersionResource) error {
	defer n.logger.Begin("inspecting namespaced resources...")()

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
	}

	// Append any additional GVRs provided as arguments
	namespacedResources = append(namespacedResources, gvrs...)

	var wg sync.WaitGroup

	for _, ns := range n.namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			defer n.logger.Begin("-- inspecting namespace %s ...", namespace)()

			// First collect the namespace itself
			nsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
			nsDir := filepath.Join(n.destDir, "namespaces", namespace)

			if err := n.client.GetResource(ctx, nsGVR, "", namespace, filepath.Join(nsDir, "namespace.yaml")); err != nil {
				n.logger.Log("WARNING: Failed to collect namespace %s: %v", namespace, err)
				return
			}

			// Collect resources in the namespace in parallel
			var resourceWg sync.WaitGroup
			for _, gvr := range namespacedResources {
				resourceWg.Add(1)
				go func(g schema.GroupVersionResource) {
					defer resourceWg.Done()

					// Use "core" for core resources (empty group) to match reference structure
					group := g.Group
					if group == "" {
						group = "core"
					}

					resourceDir := filepath.Join(nsDir, group, g.Resource)

					if err := n.client.ListResources(ctx, g, namespace, resourceDir, metav1.ListOptions{}); err != nil {
						// Some resources may not exist in all namespaces, just log and continue
						n.logger.Log("INFO: Skipped %s in namespace %s: %v", g.Resource, namespace, err)
					}
				}(gvr)
			}
			resourceWg.Wait()

			// Collect pod logs for all pods in the namespace
			n.logger.Log("-- Collecting pod logs for namespace %s ...", namespace)
			if err := n.collectPodLogs(ctx, namespace, nsDir); err != nil {
				n.logger.Log("WARNING: Failed to collect pod logs for namespace %s: %v", namespace, err)
			}
		}(ns)
	}

	wg.Wait()

	return nil
}

// collectPodLogs collects logs for all pods in a namespace
func (n *Collector) collectPodLogs(ctx context.Context, namespace, nsDir string) error {
	// Get all pods in the namespace
	pods, err := n.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	var wg sync.WaitGroup
	for _, pod := range pods.Items {
		wg.Add(1)
		go func(p corev1.Pod) {
			defer wg.Done()

			podDir := filepath.Join(nsDir, "core", "pods", p.Name)

			// Save pod YAML
			podYamlPath := filepath.Join(podDir, fmt.Sprintf("%s.yaml", p.Name))
			if err := n.client.WriteResourceToFile(&p, podYamlPath); err != nil {
				n.logger.Log("WARNING: Failed to save pod YAML for %s: %v", p.Name, err)
			}

			// Collect logs for each container
			for _, container := range p.Spec.Containers {
				containerDir := filepath.Join(podDir, container.Name, "logs")

				// Collect current logs
				if err := n.collectContainerLog(ctx, namespace, p.Name, container.Name, containerDir, "current.log", false); err != nil {
					n.logger.Log("INFO: Failed to get current logs for pod %s container %s: %v", p.Name, container.Name, err)
				}

				// Collect previous logs (from restarts)
				if err := n.collectContainerLog(ctx, namespace, p.Name, container.Name, containerDir, "previous.log", true); err != nil {
					// Previous logs may not exist if the container hasn't restarted, don't log as warning
					continue
				}
			}

			// Collect logs for init containers if any
			for _, container := range p.Spec.InitContainers {
				containerDir := filepath.Join(podDir, container.Name, "logs")

				// Collect current logs
				if err := n.collectContainerLog(ctx, namespace, p.Name, container.Name, containerDir, "current.log", false); err != nil {
					n.logger.Log("INFO: Failed to get current logs for init container %s in pod %s: %v", container.Name, p.Name, err)
				}

				// Collect previous logs (from restarts)
				if err := n.collectContainerLog(ctx, namespace, p.Name, container.Name, containerDir, "previous.log", true); err != nil {
					// Previous logs may not exist if the container hasn't restarted
					continue
				}
			}
		}(pod)
	}

	wg.Wait()
	return nil
}

// collectContainerLog collects a single container log
func (n *Collector) collectContainerLog(ctx context.Context, namespace, podName, containerName, destDir, filename string, previous bool) error {
	logOpts := &corev1.PodLogOptions{
		Container: containerName,
		Previous:  previous,
	}

	req := n.client.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
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
