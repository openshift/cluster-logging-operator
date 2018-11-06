package k8shandler

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
	pod            *v1.Pod
	containerName  string
	command        []string
	kubeConfigPath string
	masterURL      string
	stdOut         bool
	stdErr         bool
	tty            bool
}

func PodExec(config *ExecConfig) (*bytes.Buffer, *bytes.Buffer, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	esPod := config.pod
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
		Container: config.containerName,
		Command:   config.command,
		Stdout:    config.stdOut,
		Stderr:    config.stdErr,
	}, scheme.ParameterCodec)

	restClientConfig, err := clientcmd.BuildConfigFromFlags(config.masterURL, config.kubeConfigPath)
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
		Tty:    config.tty,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("remote execution failed: %v", err)
	}

	return &execOut, &execErr, nil
}
