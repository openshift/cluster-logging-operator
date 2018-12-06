package utils

import (
	"bytes"

	"k8s.io/api/core/v1"
)

func ElasticsearchExec(pod *v1.Pod, command []string) (*bytes.Buffer, *bytes.Buffer, error) {
	// when running in a pod, use the values provided for the sa
	// this is primarily used when testing
	kubeConfigPath := LookupEnvWithDefault("KUBERNETES_CONFIG", "")
	masterURL := "https://kubernetes.default.svc"
	if kubeConfigPath == "" {
		// ExecConfig requires both are "", or both have a real value
		masterURL = ""
	}
	config := &ExecConfig{
		Pod:            pod,
		ContainerName:  "elasticsearch",
		Command:        command,
		KubeConfigPath: kubeConfigPath,
		MasterURL:      masterURL,
		StdOut:         true,
		StdErr:         true,
		Tty:            false,
	}
	return PodExec(config)
}
