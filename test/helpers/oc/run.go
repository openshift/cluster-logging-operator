package oc

import (
	"fmt"
	"strings"
)

// Runner is interface for collecting arguments for Run command
type Runner interface {
	Command

	// argument for option --config
	WithConfig(string) Runner
	// argument for option -n
	WithNamespace(string) Runner

	// argument for pod name
	Name(string) Runner
	// argument for --image
	Image(string) Runner

	// argument for --env
	Env(key, value string) Runner
	// argument for --labels
	Labels(string) Runner
	// argument for --port
	Port(string) Runner
	// argument for --restart
	Restart(string) Runner
	// argument for --overrides
	Overrides(string) Runner

	// argument for --command
	Command() Runner
	// argument for --rm
	Rm() Runner
	// argument for --attach
	Attach() Runner
	// argument for --stdin
	Stdin() Runner
	// argument for --tty
	Tty() Runner

	// command and arguments to be executed in pod
	// arguments for option --
	WithCmd(string, ...string) Runner
}

type run struct {
	*runner
	namespace string

	name  string
	image string

	envVars   map[string]string
	labels    string
	port      string
	restart   string
	overrides string
	command   bool
	rm        bool
	attach    bool
	stdin     bool
	tty       bool
	cmd       string
	cmdArgs   []string
}

// Run creates an 'oc run' command
func Run() Runner {
	r := &run{
		runner:  &runner{},
		envVars: make(map[string]string),
	}
	r.collectArgsFunc = r.args
	return r
}

func (r *run) WithConfig(cfg string) Runner {
	r.configPath = cfg
	return r
}

func (r *run) WithNamespace(namespace string) Runner {
	r.namespace = namespace
	return r
}

func (r *run) Name(name string) Runner {
	r.name = name
	return r
}

func (r *run) Image(image string) Runner {
	r.image = image
	return r
}

func (r *run) Env(key, value string) Runner {
	r.envVars[key] = value
	return r
}

func (r *run) Labels(labels string) Runner {
	r.labels = labels
	return r
}

func (r *run) Port(port string) Runner {
	r.port = port
	return r
}

func (r *run) Restart(restart string) Runner {
	r.restart = restart
	return r
}

func (r *run) Overrides(overrides string) Runner {
	r.overrides = overrides
	return r
}

func (r *run) Command() Runner {
	r.command = true
	return r
}

func (r *run) Rm() Runner {
	r.rm = true
	return r
}

func (r *run) Attach() Runner {
	r.attach = true
	return r
}

func (r *run) Stdin() Runner {
	r.stdin = true
	return r
}

func (r *run) Tty() Runner {
	r.tty = true
	return r
}

func (r *run) WithCmd(command string, args ...string) Runner {
	r.cmd = command
	r.cmdArgs = args
	return r
}

func (r *run) args() []string {
	var args []string
	if r.namespace != "" {
		args = append(args, "-n", r.namespace)
	}
	args = append(args, "run", r.name)

	if r.image != "" {
		args = append(args, "--image="+r.image)
	}

	for key, value := range r.envVars {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}

	if r.labels != "" {
		args = append(args, "--labels="+r.labels)
	}

	if r.port != "" {
		args = append(args, "--port="+r.port)
	}

	if r.restart != "" {
		args = append(args, "--restart="+r.restart)
	}

	if r.overrides != "" {
		args = append(args, "--overrides="+r.overrides)
	}

	if r.command {
		args = append(args, "--command")
	}

	if r.rm {
		args = append(args, "--rm")
	}

	if r.attach {
		args = append(args, "--attach")
	}

	if r.stdin {
		args = append(args, "--stdin")
	}

	if r.tty {
		args = append(args, "--tty")
	}

	if r.cmd != "" {
		args = append(args, "--", r.cmd)
		args = append(args, r.cmdArgs...)
	}

	return args
}

func (r *run) String() string {
	return fmt.Sprintf("%s %s", r.runner.String(), strings.Join(r.args(), " "))
}
