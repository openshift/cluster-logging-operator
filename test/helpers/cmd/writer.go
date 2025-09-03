package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
)

// Writer writes to stdin of a running exec.Cmd.
type Writer struct {
	io.WriteCloser // Pipe to stdin
	*exec.Cmd
	stderr bytes.Buffer
}

// NewWriter starts an exec.Cmd and returns an io.WriteCloser for its stdin.
func NewExecWriter(cmd *exec.Cmd) (*Writer, error) {
	p, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	w := &Writer{WriteCloser: p, Cmd: cmd}
	if cmd.Stderr == nil {
		// Capture stderr because exec.Start() doesn't fill in exec.ExitError.Stderr.
		cmd.Stderr = &w.stderr
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return w, nil
}

// Close the writer's stdin and wait for its command to exit.
func (w *Writer) Close() error {
	_ = w.WriteCloser.Close()
	err := w.Wait()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(w.stderr.String()))
	}
	return nil
}

// PodWriter returns an io.WriteCloser that appends to a file in a pod container.
func PodWriter(pod *corev1.Pod, container, filename string) (io.WriteCloser, error) {
	return NewExecWriter(runtime.ExecContainer(pod, container, "sh", "-c", "cat >"+filename))
}

// PodWrite is a shortcut for NewPodWriter(), Write(), Close()
func PodWrite(pod *corev1.Pod, container, filename string, data []byte) error {
	w, err := PodWriter(pod, container, filename)
	if err == nil {
		defer w.Close()
		_, err = w.Write(data)
	}
	return err
}
