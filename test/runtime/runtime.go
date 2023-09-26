// package runtime provides conveniences based on "k8s.io/apimachinery/pkg/runtime"
package runtime

import (
	"os/exec"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

// Exec returns an `oc exec` Cmd to run cmd on o.
func Exec(o runtime.Object, cmd string, args ...string) *exec.Cmd {
	return ExecContainer(o, "", cmd, args...)
}

func ExecOc(o runtime.Object, container, cmd string, args ...string) (string, error) {
	m := runtime.Meta(o)
	return oc.Exec().WithNamespace(m.GetNamespace()).Pod(m.GetName()).Container(strings.ToLower(container)).WithCmd(cmd, args...).Run()
}

// ExecContainer returns an `oc exec` Cmd to run cmd on o.
func ExecContainer(o runtime.Object, container, cmd string, args ...string) *exec.Cmd {
	m := runtime.Meta(o)
	ocCmd := []string{
		"exec",
		"-i",
		"-n", m.GetNamespace(),
	}
	if container != "" {
		ocCmd = append(ocCmd, "-c", strings.ToLower(container))
	}
	ocCmd = append(ocCmd,
		runtime.GroupVersionKind(o).Kind+"/"+m.GetName(),
		"--",
		cmd)
	ocCmd = append(ocCmd, args...)
	return exec.Command("oc", ocCmd...)
}
