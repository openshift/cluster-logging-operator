package oc

import (
	"fmt"
	"strings"
)

// Execer is interface for collecting arguments for Exec command
type Execer interface {
	Command

	// argument for option --config
	WithConfig(string) Execer
	// argument for option -n
	WithNamespace(string) Execer

	// argument for podname
	Pod(string) Execer
	// a command can be composed to get the pod name. e.g. oc.Get or oc.Literal
	WithPodGetter(Command) Execer

	// argument for option -c
	Container(string) Execer
	// a command can be composed to get the container name. e.g. oc.Get or oc.Literal
	WithContainerGetter(Command) Execer

	// command and argument to be executed in pod
	// arguments for option --
	WithCmd(string, ...string) Execer
}

type exec struct {
	*runner
	namespace string

	podname   string
	podGetter Command

	container       string
	containerGetter Command

	command     string
	commandargs []string
}

// Exec creates an 'oc exec' command
func Exec() Execer {
	e := &exec{
		runner: &runner{},
	}
	e.collectArgsFunc = e.args
	return e
}

func (e *exec) WithConfig(cfg string) Execer {
	e.configPath = cfg
	return e
}

func (e *exec) WithNamespace(namespace string) Execer {
	e.namespace = namespace
	return e
}

func (e *exec) Pod(podname string) Execer {
	e.podname = podname
	return e
}

func (e *exec) WithPodGetter(cmd Command) Execer {
	e.podGetter = cmd
	return e
}

func (e *exec) Container(container string) Execer {
	e.container = strings.ToLower(container)
	return e
}

func (e *exec) WithContainerGetter(cmd Command) Execer {
	e.containerGetter = cmd
	return e
}

func (e *exec) WithCmd(command string, args ...string) Execer {
	e.command = command
	e.commandargs = args
	return e
}

func (e *exec) String() string {
	namespaceStr := ""
	if e.namespace != "" {
		namespaceStr = fmt.Sprintf("-n %s", e.namespace)
	}
	podStr := e.podname
	if e.podGetter != nil {
		podStr = fmt.Sprintf("$(%s)", e.podGetter.String())
	}
	containerStr := ""
	if e.container != "" {
		containerStr = fmt.Sprintf("-c %s", e.container)
	}
	if e.containerGetter != nil {
		containerStr = fmt.Sprintf("-c %s", e.containerGetter.String())
	}
	return sanitizeArgStr(fmt.Sprintf("%s %s exec %s %s -- %s %s", e.runner.String(), namespaceStr, podStr, containerStr, e.command, strings.Join(e.commandargs, " ")))
}

// creates command args to be used by runner
func (e *exec) args() []string {
	var err error
	namespaceStr := ""
	if e.namespace != "" {
		namespaceStr = fmt.Sprintf("-n %s", e.namespace)
	}
	if e.podGetter != nil {
		e.podname, err = e.podGetter.Run()
		if err != nil {
			e.runner.err = err
		}
	}
	if e.containerGetter != nil {
		e.container, err = e.containerGetter.Run()
		if err != nil {
			e.runner.err = err
		}
	}
	containerStr := ""
	if e.container != "" {
		containerStr = fmt.Sprintf("-c %s", e.container)
	}
	ocargs := sanitizeArgs(fmt.Sprintf("%s exec %s %s", namespaceStr, e.podname, containerStr))
	ocargs = append(ocargs, "--", e.command)
	ocargs = append(ocargs, e.commandargs...)
	return ocargs
}
