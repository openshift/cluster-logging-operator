package mustgather

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/yaml"
)

// Client wraps Kubernetes client functionality for must-gather operations
type Client struct {
	clientset      *kubernetes.Clientset
	dynamicClient  dynamic.Interface
	config         *rest.Config
	logger         *Logger
}

// NewClient creates a new Kubernetes client for must-gather operations
func NewClient(logger *Logger) (*Client, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
		logger:        logger,
	}, nil
}

// GetPods returns a list of pods matching the label selector
func (c *Client) GetPods(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	return podList.Items, nil
}

// GetNamespaces returns all namespaces
func (c *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, len(nsList.Items))
	for i, ns := range nsList.Items {
		namespaces[i] = ns.Name
	}
	return namespaces, nil
}

// ExecInPod executes a command in a pod container and returns the output
func (c *Client) ExecInPod(ctx context.Context, namespace, pod, container string, command []string) (string, error) {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr io.Writer
	stdoutBuf := &writeBuffer{}
	stderrBuf := &writeBuffer{}
	stdout = stdoutBuf
	stderr = stderrBuf

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})

	if err != nil {
		return "", fmt.Errorf("exec failed: %w, stderr: %s", err, stderrBuf.String())
	}

	return stdoutBuf.String(), nil
}

// GetResource gets a resource and writes it as YAML to the destination file
func (c *Client) GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name, destPath string) error {
	var resource runtime.Object
	var err error

	if namespace != "" {
		resource, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		resource, err = c.dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	return c.WriteResourceToFile(resource, destPath)
}

// ListResources lists resources and writes them to destination directory
// Each resource is written as a separate file, similar to oc adm inspect
func (c *Client) ListResources(ctx context.Context, gvr schema.GroupVersionResource, namespace, destDir string, opts metav1.ListOptions) error {
	var listObj *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		listObj, err = c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, opts)
	} else {
		listObj, err = c.dynamicClient.Resource(gvr).List(ctx, opts)
	}

	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	// Only create directory if we have items to write
	if len(listObj.Items) == 0 {
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Write each resource as a separate file
	for _, item := range listObj.Items {
		name := item.GetName()
		if name == "" {
			name = item.GetGenerateName()
		}

		itemPath := filepath.Join(destDir, fmt.Sprintf("%s.yaml", name))
		if err := c.WriteResourceToFile(&item, itemPath); err != nil {
			c.logger.Log("WARNING: Failed to write %s: %v", name, err)
			continue
		}
	}

	return nil
}

// WriteResourceToFile writes a Kubernetes resource as YAML to a file
func (c *Client) WriteResourceToFile(resource runtime.Object, destPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(destPath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// GetDynamicClient returns the underlying dynamic client
func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// GetClientset returns the underlying clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// writeBuffer is a simple buffer that implements io.Writer
type writeBuffer struct {
	data []byte
}

func (w *writeBuffer) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *writeBuffer) String() string {
	return string(w.data)
}
