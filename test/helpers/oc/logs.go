package oc

import (
	"fmt"
	"strings"
)

// Logs is interface for collecting arguments for Logs command
type LogsCommand interface {
	Command

	// argument for option -n
	WithNamespace(string) LogsCommand

	// argument for podname
	WithPod(string) LogsCommand

	// argument for option -c
	WithContainer(string) LogsCommand
}

type logConfig struct {
	*runner
	namespace string

	podname   string
	container string
}

// Exec creates an 'oc exec' command
func Logs() LogsCommand {
	e := &logConfig{
		runner: &runner{},
	}
	e.collectArgsFunc = e.args
	return e
}

func (e *logConfig) WithNamespace(namespace string) LogsCommand {
	e.namespace = namespace
	return e
}

func (e *logConfig) WithPod(podname string) LogsCommand {
	e.podname = podname
	return e
}

func (e *logConfig) WithContainer(container string) LogsCommand {
	e.container = strings.ToLower(container)
	return e
}

// creates command args to be used by runner
func (e *logConfig) args() []string {
	namespaceStr := ""
	if e.namespace != "" {
		namespaceStr = fmt.Sprintf("-n %s", e.namespace)
	}
	containerStr := ""
	if e.container != "" {
		containerStr = fmt.Sprintf("-c %s", e.container)
	}
	ocargs := sanitizeArgs(fmt.Sprintf("%s logs %s %s", namespaceStr, e.podname, containerStr))
	return ocargs
}
