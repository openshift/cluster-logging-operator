package podman

import (
	"fmt"
	"strings"
)

// Getter is interface for collecting arguments for Get command
type Runner interface {
	Command

	Named(name string) Runner
	Detached(bool) Runner

	AsPrivileged(bool) Runner

	WithVolume(source, dest string) Runner
	WithEnvVar(name, value string) Runner
	WithImage(image string) Runner
	WithNetwork(name string) Runner

	// command and argument to be executed in pod
	// arguments for option --
	WithCmd(string, ...string) Runner
}

type run struct {
	*runner
	privileged bool
	detached   bool

	volumes     []string
	envvars     []string
	networks    []string
	image       string
	cmd         string
	name        string
	commandArgs []string
}

// Get creates an 'podman run' command
func Run() Runner {
	r := &run{
		runner:      &runner{},
		volumes:     []string{},
		envvars:     []string{},
		commandArgs: []string{},
	}
	r.collectArgsFunc = r.args
	return r
}

func (r *run) Named(name string) Runner {
	r.name = name
	return r
}
func (r *run) Detached(detached bool) Runner {
	r.detached = detached
	return r
}
func (r *run) AsPrivileged(privileged bool) Runner {
	r.privileged = privileged
	return r
}

func (r *run) WithEnvVar(name, value string) Runner {
	r.envvars = append(r.envvars, fmt.Sprintf("%s=%s", name, value))
	return r
}
func (r *run) WithVolume(source, destination string) Runner {
	r.volumes = append(r.volumes, fmt.Sprintf("%s:%s", source, destination))
	return r
}
func (r *run) WithNetwork(name string) Runner {
	r.networks = append(r.networks, name)
	return r
}

func (r *run) WithImage(image string) Runner {
	r.image = image
	return r
}
func (r *run) WithCmd(command string, args ...string) Runner {
	r.cmd = command
	r.commandArgs = args
	return r
}

func (r *run) args() []string {
	args := []string{"run"}
	if r.name != "" {
		args = append(args, "--name", r.name)
	}
	if r.detached {
		args = append(args, "-d")
	}
	if r.privileged {
		args = append(args, "--privileged")
	}
	for _, env := range r.envvars {
		args = append(args, "-e", env)
	}
	for _, v := range r.volumes {
		args = append(args, "-v", v)
	}
	for _, n := range r.networks {
		args = append(args, "--network", n)
	}
	args = append(args, r.image)
	if r.cmd != "" {
		args = append(args, r.cmd)
		if len(r.commandArgs) > 0 {
			args = append(args, r.commandArgs...)
		}
	}
	return args
}

func (r *run) String() string {
	return fmt.Sprintf("%s %s", r.runner.String(), strings.Join(r.args(), " "))
}
