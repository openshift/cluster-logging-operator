package utils

import (
	"bytes"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type ExecConfig struct {
	Pod            *v1.Pod
	ContainerName  string
	Command        []string
	KubeConfigPath string
	MasterURL      string
	StdOut         bool
	StdErr         bool
	Tty            bool
}

func PodExec(config *ExecConfig) (*bytes.Buffer, *bytes.Buffer, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	esPod := config.Pod
	if esPod.Status.Phase != v1.PodRunning {
		return nil, nil, fmt.Errorf("elasticsearch pod [%s] found but isn't running", esPod.Name)
	}

	client := k8sclient.GetKubeClient()
	execRequest := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(esPod.Name).
		Namespace(esPod.Namespace).
		SubResource("exec")

	execRequest.VersionedParams(&v1.PodExecOptions{
		Container: config.ContainerName,
		Command:   config.Command,
		Stdout:    config.StdOut,
		Stderr:    config.StdErr,
	}, scheme.ParameterCodec)

	restClientConfig, err := clientcmd.BuildConfigFromFlags(config.MasterURL, config.KubeConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error when creating rest client command: %v", err)
	}
	exec, err := remotecommand.NewSPDYExecutor(restClientConfig, "POST", execRequest.URL())
	if err != nil {
		return nil, nil, fmt.Errorf("error when creating remote command executor: %v", err)
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    config.Tty,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("remote execution failed: %v", err)
	}

	return &execOut, &execErr, nil
}
