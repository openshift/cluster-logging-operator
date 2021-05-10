package podman

import (
	"fmt"
	"strings"
)

// Stopper is interface for collecting arguments for the Stop command
type Stopper interface {
	Command

	WithContainer(containerId ...string) Stopper
}

type stop struct {
	*runner
	ids []string
}

// Stop creates an 'podman stop' command
func Stop() Stopper {
	sub := &stop{
		runner: &runner{},
		ids:    []string{},
	}
	sub.collectArgsFunc = sub.args
	return sub
}

func (sub *stop) WithContainer(ids ...string) Stopper {
	sub.ids = append(sub.ids, ids...)
	return sub
}

func (sub *stop) args() []string {
	args := []string{"stop"}
	return append(args, sub.ids...)
}

func (sub *stop) String() string {
	return fmt.Sprintf("%s %s", sub.runner.String(), strings.Join(sub.args(), " "))
}
