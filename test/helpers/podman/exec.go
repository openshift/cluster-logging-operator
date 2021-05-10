package podman

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	osexec "os/exec"
	"strings"
)

// Execer is interface for collecting arguments for the Exec command
type Execer interface {
	Command

	ToContainer(containerId string) Execer
	WithCmd(command string, args ...string) Execer

	Reader() (*cmd.Reader, error)
}

type execer struct {
	*runner
	id          string
	cmd         string
	commandArgs []string
}

// Exec creates an 'podman exec' command
func Exec() Execer {
	sub := &execer{
		runner: &runner{},
	}
	sub.collectArgsFunc = sub.args
	return sub
}

func (sub *execer) WithCmd(command string, args ...string) Execer {
	sub.cmd = command
	sub.commandArgs = args
	return sub
}
func (sub *execer) ToContainer(id string) Execer {
	sub.id = id
	return sub
}

func (sub *execer) args() []string {
	args := []string{"exec", sub.id}
	args = append(args, sub.cmd)
	if len(sub.commandArgs) > 0 {
		args = append(args, sub.commandArgs...)
	}

	return args
}

func (sub *execer) String() string {
	return fmt.Sprintf("%s %s", sub.runner.String(), strings.Join(sub.args(), " "))
}

func (sub *execer) Reader() (*cmd.Reader, error) {
	args := sub.args()
	command := osexec.Command(CMD, args...)
	return cmd.NewReader(command)
}
