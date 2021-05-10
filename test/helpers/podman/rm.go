package podman

import (
	"fmt"
	"strings"
)

// Remover is interface for collecting arguments for the 'rm' command
type Remover interface {
	Command

	WithContainer(containerId ...string) Remover
}

type remove struct {
	*runner
	ids []string
}

// RM creates an 'podman rm' command
func RM() Remover {
	sub := &remove{
		runner: &runner{},
		ids:    []string{},
	}
	sub.collectArgsFunc = sub.args
	return sub
}

func (sub *remove) WithContainer(ids ...string) Remover {
	sub.ids = append(sub.ids, ids...)
	return sub
}

func (sub *remove) args() []string {
	args := []string{"rm"}
	return append(args, sub.ids...)
}

func (sub *remove) String() string {
	return fmt.Sprintf("%s %s", sub.runner.String(), strings.Join(sub.args(), " "))
}
